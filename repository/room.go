//go:generate go tool mockgen -source=$GOFILE -destination=mockrepository/$GOFILE -package=mockrepository
package repository

import (
	"context"
	"errors"

	"github.com/traPtitech/rucQ/model"
)

var (
	ErrRoomNotFound            = errors.New("room not found")
	ErrUserOrRoomGroupNotFound = errors.New("user or room group not found")
)

type RoomRepository interface {
	GetRooms() ([]model.Room, error)
	GetRoomByID(id uint) (*model.Room, error)
	GetRoomByUserID(ctx context.Context, campID uint, userID string) (*model.Room, error)
	CreateRoom(ctx context.Context, room *model.Room) error
	UpdateRoom(ctx context.Context, roomID uint, room *model.Room) error
	DeleteRoom(ctx context.Context, roomID uint) error
}
