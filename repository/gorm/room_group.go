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

func (r *Repository) UpdateRoomGroup(
	ctx context.Context,
	roomGroupID uint,
	roomGroup *model.RoomGroup,
) error {
	if _, err := gorm.G[*model.RoomGroup](r.db).Where(&model.RoomGroup{
		Model: gorm.Model{
			ID: roomGroupID,
		},
	}).Updates(ctx, roomGroup); err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetRoomGroupByID(ctx context.Context, roomGroupID uint) (*model.RoomGroup, error) {
	roomGroup, err := gorm.G[model.RoomGroup](r.db).
		Preload("Rooms", nil).
		Where("id = ?", roomGroupID).
		First(ctx)

	if err != nil {
		return nil, err
	}

	return &roomGroup, nil
}
