package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type MessageType string

const (
	JOIN_ROOM     MessageType = "join_room"
	LEAVE_ROOM    MessageType = "leave_room"
	OFFER         MessageType = "offer"
	ANSWER        MessageType = "answer"
	ICE_CANDIDATE MessageType = "ice_candidate"
	USER_JOINED   MessageType = "user_joined"
	USER_LEFT     MessageType = "user_left"
	ROOM_USERS    MessageType = "room_users"
)

type Message struct {
	Type     MessageType `json:"type"`
	RoomID   string      `json:"room_id,omitempty"`
	UserID   string      `json:"user_id,omitempty"`
	ToUserID string      `json:"to_user_id,omitempty"`
	Data     interface{} `json:"data,omitempty"`
}

type Client struct {
	ID        string
	Conn      *websocket.Conn
	RoomID    string
	PeerConns map[string]*webrtc.PeerConnection
	Send      chan Message
	mutex     sync.RWMutex
}

type Room struct {
	ID      string
	Clients map[string]*Client
	mutex   sync.RWMutex
}

type Hub struct {
	Rooms      map[string]*Room
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan Message
	mutex      sync.RWMutex
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[string]*Room),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.Broadcast:
			h.broadcastMessage(message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.Rooms[client.RoomID] == nil {
		h.Rooms[client.RoomID] = &Room{
			ID:      client.RoomID,
			Clients: make(map[string]*Client),
		}
	}

	room := h.Rooms[client.RoomID]
	room.mutex.Lock()
	room.Clients[client.ID] = client
	room.mutex.Unlock()

	h.sendRoomUsers(client.RoomID, client.ID)

	h.notifyUserJoined(client.RoomID, client.ID)

	log.Printf("Client %s joined room %s", client.ID, client.RoomID)
}

func (h *Hub) unregisterClient(client *Client) {
	h.mutex.RLock()
	room := h.Rooms[client.RoomID]
	h.mutex.RUnlock()

	if room != nil {
		room.mutex.Lock()
		if _, ok := room.Clients[client.ID]; ok {
			delete(room.Clients, client.ID)
			close(client.Send)

			client.mutex.Lock()
			for _, pc := range client.PeerConns {
				pc.Close()
			}
			client.mutex.Unlock()
		}
		room.mutex.Unlock()

		h.notifyUserLeft(client.RoomID, client.ID)

		log.Printf("Client %s left room %s", client.ID, client.RoomID)
	}
}

func (h *Hub) broadcastMessage(message Message) {
	h.mutex.RLock()
	room := h.Rooms[message.RoomID]
	h.mutex.RUnlock()

	if room != nil {
		room.mutex.RLock()
		defer room.mutex.RUnlock()

		if message.ToUserID != "" {
			if client, ok := room.Clients[message.ToUserID]; ok {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(room.Clients, message.ToUserID)
				}
			}
		} else {
			for userID, client := range room.Clients {
				if userID != message.UserID {
					select {
					case client.Send <- message:
					default:
						close(client.Send)
						delete(room.Clients, userID)
					}
				}
			}
		}
	}
}

func (h *Hub) sendRoomUsers(roomID, userID string) {
	h.mutex.RLock()
	room := h.Rooms[roomID]
	h.mutex.RUnlock()

	if room != nil {
		room.mutex.RLock()
		users := make([]string, 0, len(room.Clients))
		for id := range room.Clients {
			if id != userID {
				users = append(users, id)
			}
		}
		room.mutex.RUnlock()

		if client, ok := room.Clients[userID]; ok {
			message := Message{
				Type: ROOM_USERS,
				Data: users,
			}
			select {
			case client.Send <- message:
			default:
			}
		}
	}
}

func (h *Hub) notifyUserJoined(roomID, userID string) {
	message := Message{
		Type:   USER_JOINED,
		RoomID: roomID,
		UserID: userID,
	}
	h.Broadcast <- message
}

func (h *Hub) notifyUserLeft(roomID, userID string) {
	message := Message{
		Type:   USER_LEFT,
		RoomID: roomID,
		UserID: userID,
	}
	h.Broadcast <- message
}

func createPeerConnection() (*webrtc.PeerConnection, error) {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	return webrtc.NewPeerConnection(config)
}

func (h *Hub) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	userID := r.URL.Query().Get("user_id")
	roomID := r.URL.Query().Get("room_id")

	if userID == "" || roomID == "" {
		conn.Close()
		return
	}

	client := &Client{
		ID:        userID,
		Conn:      conn,
		RoomID:    roomID,
		PeerConns: make(map[string]*webrtc.PeerConnection),
		Send:      make(chan Message, 256),
	}

	h.Register <- client

	go client.writePump()
	go client.readPump(h)
}

func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		var message Message
		err := c.Conn.ReadJSON(&message)
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		message.RoomID = c.RoomID
		message.UserID = c.ID

		switch message.Type {
		case OFFER:
			c.handleOffer(message, hub)
		case ANSWER:
			c.handleAnswer(message, hub)
		case ICE_CANDIDATE:
			c.handleICECandidate(message, hub)
		}
	}
}

func (c *Client) writePump() {
	defer c.Conn.Close()

	for message := range c.Send {
		if err := c.Conn.WriteJSON(message); err != nil {
			log.Println("Write error:", err)
			return
		}
	}
	c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func (c *Client) handleOffer(message Message, hub *Hub) {
	toUserID := message.ToUserID
	if toUserID == "" {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.PeerConns[toUserID] == nil {
		pc, err := createPeerConnection()
		if err != nil {
			log.Println("Error creating peer connection:", err)
			return
		}
		c.PeerConns[toUserID] = pc

		pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
			if candidate != nil {
				candidateMsg := Message{
					Type:     ICE_CANDIDATE,
					RoomID:   c.RoomID,
					UserID:   c.ID,
					ToUserID: toUserID,
					Data:     candidate.ToJSON(),
				}
				hub.Broadcast <- candidateMsg
			}
		})
	}

	hub.Broadcast <- message
}

func (c *Client) handleAnswer(message Message, hub *Hub) {
	hub.Broadcast <- message
}

func (c *Client) handleICECandidate(message Message, hub *Hub) {
	hub.Broadcast <- message
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	hub := NewHub()
	go hub.Run()

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", hub.handleWebSocket)

	log.Println("Server starting on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
