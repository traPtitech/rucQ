package gormrepository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
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

func (r *Repository) GetRoomByUserID(ctx context.Context, userID string) (*model.Room, error) {
	var room model.Room

	if err := r.db.
		WithContext(ctx).
		Joins("JOIN room_members ON room_members.room_id = rooms.id").
		Where("room_members.user_id = ?", userID).
		Preload("Members").
		First(&room).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrRoomNotFound
		}

		return nil, err
	}

	return &room, nil
}

func (r *Repository) CreateRoom(ctx context.Context, room *model.Room) error {
	if err := r.db.
		WithContext(ctx).
		Omit("Members.*"). // 関係は更新するがユーザーの新規作成はされないようにする
		Create(room).
		Error; err != nil {
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			return repository.ErrUserOrRoomGroupNotFound
		}

		return err
	}

	return nil
}

func (r *Repository) UpdateRoom(ctx context.Context, roomID uint, room *model.Room) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		room.ID = roomID

		rowsAffected, err := gorm.G[*model.Room](tx).Omit("Members").Updates(ctx, room)

		if err != nil {
			if errors.Is(err, gorm.ErrForeignKeyViolated) {
				return repository.ErrRoomGroupNotFound
			}

			return err
		}

		if rowsAffected == 0 {
			return repository.ErrRoomNotFound
		}

		if err := tx.WithContext(ctx).
			Model(room).
			Omit("Members.*"). // ユーザーの新規作成はされないようにする
			Association("Members").
			Replace(room.Members); err != nil {
			if errors.Is(err, gorm.ErrForeignKeyViolated) {
				return repository.ErrUserNotFound
			}

			return err
		}

		return nil
	})
}
