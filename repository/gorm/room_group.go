package gorm

import (
	"context"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
)

func (r *Repository) CreateRoomGroup(ctx context.Context, roomGroup *model.RoomGroup) error {
	if err := gorm.G[model.RoomGroup](r.db).Create(ctx, roomGroup); err != nil {
		return err
	}

	return nil
}
