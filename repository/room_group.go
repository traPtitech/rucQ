//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package repository

import (
	"context"

	"github.com/traPtitech/rucQ/model"
)

type RoomGroupRepository interface {
	CreateRoomGroup(ctx context.Context, roomGroup *model.RoomGroup) error
	UpdateRoomGroup(ctx context.Context, roomGroupID uint, roomGroup *model.RoomGroup) error
	GetRoomGroupByID(ctx context.Context, roomGroupID uint) (*model.RoomGroup, error)
}
