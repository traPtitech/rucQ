//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package repository

import "github.com/traPtitech/rucQ/model"

type RoomRepository interface {
	GetRooms() ([]model.Room, error)
	GetRoomByID(id uint) (*model.Room, error)
	CreateRoom(room *model.Room) error
	UpdateRoom(room *model.Room) error
}
