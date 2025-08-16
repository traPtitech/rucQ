//go:generate go tool mockgen -source=$GOFILE -destination=mockrepository/$GOFILE -package=mockrepository
package repository

import (
	"context"
	"errors"

	"github.com/traPtitech/rucQ/model"
)

var ErrRoomGroupNotFound = errors.New("room group not found")

type RoomGroupRepository interface {
	CreateRoomGroup(ctx context.Context, roomGroup *model.RoomGroup) error
	UpdateRoomGroup(ctx context.Context, roomGroupID uint, roomGroup *model.RoomGroup) error
	GetRoomGroupByID(ctx context.Context, roomGroupID uint) (*model.RoomGroup, error)
	GetRoomGroups(ctx context.Context, campID uint) ([]model.RoomGroup, error)
}
