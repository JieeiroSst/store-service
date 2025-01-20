package services

import (
	"github.com/JIeeiroSst/room-service/internal/core/domain/models"
	"github.com/JIeeiroSst/room-service/internal/core/ports"
	"golang.org/x/crypto/bcrypt"
)

type roomService struct {
	roomRepo    ports.RoomRepository
	messageRepo ports.MessageRepository
	authService ports.AuthService
}

func NewRoomService(
	roomRepo ports.RoomRepository,
	messageRepo ports.MessageRepository,
	authService ports.AuthService,
) ports.RoomService {
	return &roomService{
		roomRepo:    roomRepo,
		messageRepo: messageRepo,
		authService: authService,
	}
}

func (s *roomService) CreateRoom(room *models.Room) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(room.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	room.Password = string(hashedPassword)
	return s.roomRepo.Create(room)
}

func (s *roomService) ListRooms() ([]models.Room, error) {
	return s.roomRepo.FindAll()
}

func (s *roomService) JoinRoom(id uint, password, username string) (string, error) {
	room, err := s.roomRepo.FindByID(id)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(room.Password), []byte(password)); err != nil {
		return "", err
	}

	return s.authService.GenerateToken(room.ID, username)
}

func (s *roomService) GetMessages(roomID uint) ([]models.Message, error) {
	return s.messageRepo.FindByRoomID(roomID)
}
