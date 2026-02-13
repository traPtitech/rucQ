//go:generate go tool mockgen -source=$GOFILE -destination=mockrepository/$GOFILE -package=mockrepository
package repository

import (
	"context"

	"github.com/traPtitech/rucQ/model"
)

type RoomStatusRepository interface {
	SetRoomStatus(ctx context.Context, roomID uint, status *model.RoomStatus, operatorID string) error
	GetRoomStatusLogs(ctx context.Context, roomID uint) ([]model.RoomStatusLog, error)
}
