//go:generate go tool mockgen -source=$GOFILE -destination=mockrepository/$GOFILE -package=mockrepository
package repository

import "github.com/traPtitech/rucQ/model"

type RoomRepository interface {
	GetRooms() ([]model.Room, error)
	GetRoomByID(id uint) (*model.Room, error)
	CreateRoom(room *model.Room) error
	UpdateRoom(room *model.Room) error
}
