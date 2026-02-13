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
	status model.RoomStatus,
	operatorID string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		newStatus := status
		newStatus.RoomID = roomID
		newStatus.UpdatedAt = time.Now()

		if err := tx.
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "room_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"type", "topic", "updated_at"}),
			}).
			Create(&newStatus).Error; err != nil {
			if errors.Is(err, gorm.ErrForeignKeyViolated) {
				return repository.ErrRoomNotFound
			}
			return err
		}

		log := model.RoomStatusLog{
			RoomID:     roomID,
			Type:       newStatus.Type,
			Topic:      newStatus.Topic,
			OperatorID: operatorID,
		}

		if err := gorm.G[model.RoomStatusLog](tx).Create(ctx, &log); err != nil {
			if errors.Is(err, gorm.ErrForeignKeyViolated) {
				if _, err := gorm.G[model.Room](
					tx,
				).Where("id = ?", log.RoomID).
					Take(ctx); err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return repository.ErrRoomNotFound
					}

					return err
				}

				return err
			}
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
		Order("updated_at DESC").
		Find(ctx)

	if err != nil {
		return nil, err
	}

	return logs, nil
}
