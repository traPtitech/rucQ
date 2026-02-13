package gormrepository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
)

func (r *Repository) SetRoomStatus(
	ctx context.Context,
	roomID uint,
	status *model.RoomStatus,
	operatorID string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&model.Room{}, roomID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repository.ErrRoomNotFound
			}
			return err
		}

		status.RoomID = roomID
		status.UpdatedAt = time.Now()

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "room_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"type", "topic", "updated_at"}),
		}).Create(status).Error; err != nil {
			return err
		}

		log := model.RoomStatusLog{
			RoomID:     roomID,
			Type:       status.Type,
			Topic:      status.Topic,
			OperatorID: operatorID,
		}

		if err := tx.Create(&log).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *Repository) GetRoomStatusLogs(
	ctx context.Context,
	roomID uint,
) ([]model.RoomStatusLog, error) {
	logs, err := gorm.G[model.RoomStatusLog](r.db).
		Where("room_id = ?", roomID).
		Order("updated_at ASC").
		Find(ctx)

	if err != nil {
		return nil, err
	}

	return logs, nil
}
