import React, { useState, useEffect, useRef } from 'react';
import { Loader } from 'lucide-react';

const App = () => {
  const [currentRoom, setCurrentRoom] = useState(null);
  const [username, setUsername] = useState(localStorage.getItem('username'));
  const [rooms, setRooms] = useState([]);
  const [messages, setMessages] = useState([]);
  const [socket, setSocket] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [typingUsers, setTypingUsers] = useState(new Set());
  
  const messageInputRef = useRef(null);
  const fileInputRef = useRef(null);
  const typingTimeoutRef = useRef(null);
  const reconnectAttemptsRef = useRef(0);
  const maxReconnectAttempts = 5;
  const baseDelay = 1000;

  useEffect(() => {
    init();
    setupReconnection();
    return () => {
      if (socket) socket.close();
    };
  }, []);

  const setupReconnection = () => {
    window.addEventListener('online', () => {
      if (currentRoom) {
        reconnectAttemptsRef.current = 0;
        attemptReconnect();
      }
    });

    const attemptReconnect = () => {
      if (reconnectAttemptsRef.current >= maxReconnectAttempts) {
        alert('Unable to reconnect to chat. Please refresh the page.');
        return;
      }

      const delay = baseDelay * Math.pow(2, reconnectAttemptsRef.current);
      setTimeout(() => {
        const token = localStorage.getItem('chatToken');
        if (token) {
          connectWebSocket(token);
          reconnectAttemptsRef.current++;
        }
      }, delay);
    };

    const intervalId = setInterval(() => {
      if (socket && socket.readyState === WebSocket.CLOSED) {
        attemptReconnect();
      }
    }, 5000);

    return () => clearInterval(intervalId);
  };

  const loadMessages = async (roomId) => {
    try {
      const token = localStorage.getItem('chatToken');
      const response = await fetch(`http://localhost:8081/rooms/${roomId}/messages`, {
        headers: {
          'Authorization': token ? `Bearer ${token}` : '',
          'Content-Type': 'application/json'
        }
      });
      
      if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
      
      const messagesData = await response.json();
      setMessages(messagesData.reverse());
    } catch (error) {
      setError(`Error loading messages: ${error.message}`);
    }
  };

  const init = async () => {
    const token = localStorage.getItem('chatToken');
    if (token && username) {
      loadRooms();
    } else {
      promptLogin();
    }
  };

  const loadRooms = async () => {
    try {
      setLoading(true);
      const response = await fetch('http://localhost:8081/rooms');
      if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
      const roomsData = await response.json();
      setRooms(roomsData);
    } catch (error) {
      setError(`Error loading rooms: ${error.message}`);
    } finally {
      setLoading(false);
    }
  };

  const connectWebSocket = (token) => {
    if (socket) socket.close();

    const newSocket = new WebSocket(`ws://localhost:8081/ws?token=${token}&room_id=${currentRoom}`);
    
    newSocket.onmessage = (event) => {
      const message = JSON.parse(event.data);
      if (message.type === 'typing') {
        setTypingUsers(prev => new Set([...prev, message.username]));
      } else if (message.type === 'stop_typing') {
        setTypingUsers(prev => {
          const newSet = new Set(prev);
          newSet.delete(message.username);
          return newSet;
        });
      } else {
        setMessages(prev => [message, ...prev]);
      }
    };

    newSocket.onopen = () => {
      console.log('Connected to chat server');
      reconnectAttemptsRef.current = 0;
    };

    newSocket.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    newSocket.onclose = () => {
      console.log('Disconnected from chat server');
    };

    setSocket(newSocket);
  };

  const handleJoinRoom = async (roomId) => {
    if (!username) {
      promptLogin();
      return;
    }

    const password = prompt('Enter room password:');
    if (!password) return;

    try {
      const response = await fetch(`http://localhost:8081/rooms/${roomId}/join`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      });

      if (response.ok) {
        const { token } = await response.json();
        localStorage.setItem('chatToken', token);
        setCurrentRoom(roomId);
        connectWebSocket(token);
        loadMessages(roomId);
      } else {
        alert('Invalid password');
      }
    } catch (error) {
      alert('Failed to join room');
    }
  };

  const handleSendMessage = () => {
    if (!messageInputRef.current?.value.trim() || !socket) return;

    socket.send(JSON.stringify({
      type: 'text',
      content: messageInputRef.current.value.trim(),
      roomId: currentRoom
    }));

    messageInputRef.current.value = '';
  };

  const handleTyping = () => {
    if (!socket || !currentRoom) return;

    if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);

    socket.send(JSON.stringify({
      type: 'typing',
      roomId: currentRoom,
      username
    }));

    typingTimeoutRef.current = setTimeout(() => {
      socket.send(JSON.stringify({
        type: 'stop_typing',
        roomId: currentRoom,
        username
      }));
    }, 1000);
  };

  const promptLogin = () => {
    const newUsername = prompt('Enter your username to start chatting:');
    if (newUsername) {
      localStorage.setItem('username', newUsername);
      setUsername(newUsername);
      loadRooms();
    } else {
      setTimeout(promptLogin, 500);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('chatToken');
    localStorage.removeItem('username');
    setUsername(null);
    setCurrentRoom(null);
    if (socket) socket.close();
    setMessages([]);
    promptLogin();
  };

  const handleCreateRoom = async (name, password) => {
    try {
      const response = await fetch('http://localhost:8081/rooms', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ name, password })
      });

      if (response.ok) {
        loadRooms();
      } else {
        const error = await response.json();
        alert(error.message || 'Failed to create room');
      }
    } catch (error) {
      console.error('Error creating room:', error);
      alert('Failed to create room');
    }
  };

  return (
    <div className="container">
      <RoomsList
        rooms={rooms}
        currentRoom={currentRoom}
        username={username}
        loading={loading}
        error={error}
        onJoinRoom={handleJoinRoom}
        onLogout={handleLogout}
        onCreateRoom={handleCreateRoom}
      />
      <ChatArea
        messages={messages}
        currentRoom={currentRoom}
        username={username}
        typingUsers={typingUsers}
        messageInputRef={messageInputRef}
        fileInputRef={fileInputRef}
        onSendMessage={handleSendMessage}
        onTyping={handleTyping}
      />
    </div>
  );
};

const RoomsList = ({ rooms, currentRoom, username, loading, error, onJoinRoom, onLogout, onCreateRoom }) => {
  const [newRoomName, setNewRoomName] = useState('');
  const [newRoomPassword, setNewRoomPassword] = useState('');

  const handleSubmit = (e) => {
    e.preventDefault();
    if (!newRoomName.trim() || !newRoomPassword.trim()) {
      alert('Please enter both room name and password');
      return;
    }

    onCreateRoom(newRoomName.trim(), newRoomPassword.trim());
    setNewRoomName('');
    setNewRoomPassword('');
  };

  return (
    <div className="rooms-list">
      <div className="rooms-header">
        <h2>Chat Rooms</h2>
        {username && (
          <div className="mt-2">
            Logged in as: <strong>{username}</strong>
            <button onClick={onLogout} className="ml-2">Logout</button>
          </div>
        )}
      </div>
      <div className="rooms-container">
        {loading ? (
          <div className="loading">
            <Loader className="animate-spin" />
            Loading rooms...
          </div>
        ) : error ? (
          <div className="error">{error}</div>
        ) : (
          rooms.map(room => (
            <div
              key={room.id}
              className={`room-item ${currentRoom === room.id ? 'active' : ''}`}
              onClick={() => onJoinRoom(room.id)}
            >
              <div className="room-name">{room.name}</div>
              <div className="room-users">{room.activeUsers || 0} online</div>
            </div>
          ))
        )}
      </div>
      <div className="create-room-form">
        <h3>Create New Room</h3>
        <form onSubmit={handleSubmit}>
          <div className="mt-2">
            <input
              type="text"
              value={newRoomName}
              onChange={(e) => setNewRoomName(e.target.value)}
              placeholder="Room name"
              className="w-full"
            />
          </div>
          <div className="mt-2">
            <input
              type="password"
              value={newRoomPassword}
              onChange={(e) => setNewRoomPassword(e.target.value)}
              placeholder="Room password"
              className="w-full"
            />
          </div>
          <div className="mt-2">
            <button type="submit">Create Room</button>
          </div>
        </form>
      </div>
    </div>
  );
};

const ChatArea = ({
  messages,
  currentRoom,
  username,
  typingUsers,
  messageInputRef,
  fileInputRef,
  onSendMessage,
  onTyping
}) => {
  const messagesEndRef = useRef(null);
  const messagesContainerRef = useRef(null);
  const [autoScroll, setAutoScroll] = useState(true);

  useEffect(() => {
    if (autoScroll) {
      scrollToBottom();
    }
  }, [messages, autoScroll]);

  const handleScroll = () => {
    if (!messagesContainerRef.current) return;
    
    const { scrollTop, scrollHeight, clientHeight } = messagesContainerRef.current;
    const isScrolledNearBottom = scrollHeight - (scrollTop + clientHeight) < 100;
    setAutoScroll(isScrolledNearBottom);
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  return (
    <div className="chat-container">
      <div className="chat-header">
        <h2>{currentRoom ? `Room ${currentRoom}` : 'Select a Room'}</h2>
      </div>
      <div 
        className="chat-messages"
        ref={messagesContainerRef}
        onScroll={handleScroll}
        style={{
          overflowY: 'auto',
          maxHeight: 'calc(100vh - 200px)',
          paddingRight: '10px'
        }}
      >
        {!currentRoom ? (
          <div className="loading">Join a room to start chatting</div>
        ) : (
          <>
            {messages.map((message, index) => (
              <Message key={index} message={message} isOwn={message.username === username} />
            ))}
            <div ref={messagesEndRef} />
          </>
        )}
      </div>
      {!autoScroll && messages.length > 0 && (
        <button
          className="scroll-bottom-btn"
          onClick={() => {
            scrollToBottom();
            setAutoScroll(true);
          }}
          style={{
            position: 'absolute',
            bottom: '80px',
            right: '20px',
            padding: '8px',
            borderRadius: '50%',
            backgroundColor: '#007bff',
            color: 'white',
            border: 'none',
            cursor: 'pointer',
            boxShadow: '0 2px 5px rgba(0,0,0,0.2)',
            zIndex: 10
          }}
        >
          ↓
        </button>
      )}
      {typingUsers.size > 0 && (
        <div className="typing-indicator">
          {Array.from(typingUsers).join(', ')} {typingUsers.size === 1 ? 'is' : 'are'} typing...
        </div>
      )}
      <div className="input-container">
        <input
          ref={messageInputRef}
          type="text"
          placeholder="Type a message..."
          disabled={!currentRoom}
          onKeyPress={(e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
              e.preventDefault();
              onSendMessage();
            }
            onTyping();
          }}
        />
        <input
          ref={fileInputRef}
          type="file"
          className="file-upload"
          accept="image/*"
        />
        <button
          className="upload-btn"
          disabled={!currentRoom}
          onClick={() => fileInputRef.current?.click()}
        >
          Upload Image
        </button>
        <button
          disabled={!currentRoom}
          onClick={onSendMessage}
        >
          Send
        </button>
      </div>
    </div>
  );
};

const Message = ({ message, isOwn }) => {
  const formatDate = (timestamp) => {
    const date = new Date(timestamp);
    return date.toLocaleString('en-US', {
      hour: 'numeric',
      minute: 'numeric',
      hour12: true,
      month: 'short',
      day: 'numeric'
    });
  };

  return (
    <div className={`message ${isOwn ? 'sent' : 'received'}`}>
      <div className="username">{message.username}</div>
      <div className="content">
        {message.type === 'image' ? (
          <img src={message.content} alt="Shared image" />
        ) : (
          message.content
        )}
      </div>
      <div className="timestamp">{formatDate(message.timestamp)}</div>
      </div>
  );
};

export default App;