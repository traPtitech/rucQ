package repository

import "github.com/traP-jp/rucQ/backend/model"

type RoomRepository interface {
	GetRooms() ([]model.Room, error)
	GetRoomByID(id uint) (*model.Room, error)
	CreateRoom(room *model.Room) error
	UpdateRoom(room *model.Room) error
}
