package gormrepository

import (
	"context"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
)

func (r *Repository) GetRooms() ([]model.Room, error) {
	var rooms []model.Room

	if err := r.db.Preload("Members").Find(&rooms).Error; err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *Repository) GetRoomByID(id uint) (*model.Room, error) {
	var room model.Room

	if err := r.db.Preload("Members").Where(&model.Room{
		Model: gorm.Model{
			ID: id,
		},
	}).First(&room).Error; err != nil {
		return nil, err
	}

	return &room, nil
}

func (r *Repository) CreateRoom(ctx context.Context, room *model.Room) error {
	if err := r.db.WithContext(ctx).Create(room).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) UpdateRoom(room *model.Room) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(room).Association("Members").Replace(room.Members); err != nil {
			return err
		}

		return tx.Model(room).Updates(&model.Room{
			Name: room.Name,
		}).Error
	})
}
