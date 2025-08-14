package gormrepository

import (
	"context"
	"errors"

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
	rowsAffected, err := gorm.G[*model.RoomGroup](r.db).
		Where("id = ?", roomGroupID).
		Updates(ctx, roomGroup)
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return model.ErrNotFound
	}

	return nil
}

func (r *Repository) GetRoomGroupByID(
	ctx context.Context,
	roomGroupID uint,
) (*model.RoomGroup, error) {
	roomGroup, err := gorm.G[model.RoomGroup](r.db).
		Preload("Rooms", nil).
		Where("id = ?", roomGroupID).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	return &roomGroup, nil
}
