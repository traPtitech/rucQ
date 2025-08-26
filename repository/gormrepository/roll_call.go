package gormrepository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
)

func (r *Repository) CreateRollCall(ctx context.Context, rollCall *model.RollCall) error {
	if err := r.db.WithContext(ctx).Omit("Subjects.*").Create(rollCall).Error; err != nil {
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			// 外部キーエラーが起きたときはCampかUserが存在しないので、
			// どちらが存在しないかを確認して適切なエラーを返す
			if _, err := gorm.G[*model.Camp](r.db).
				Where("id = ?", rollCall.CampID).
				First(ctx); errors.Is(err, gorm.ErrRecordNotFound) {
				return repository.ErrCampNotFound
			} else {
				return repository.ErrUserNotFound
			}
		}

		return err
	}

	return nil
}

func (r *Repository) GetRollCalls(ctx context.Context, campID uint) ([]model.RollCall, error) {
	rollCalls, err := gorm.G[model.RollCall](r.db).
		Preload("Reactions", nil).
		Preload("Subjects", nil).
		Where("camp_id = ?", campID).
		Find(ctx)

	if err != nil {
		return nil, err
	}

	// RollCallが見つからなかった場合、Campが存在しない可能性を考慮してCampの存在確認を行う
	if len(rollCalls) == 0 {
		if _, err := gorm.G[*model.Camp](r.db).
			Where("id = ?", campID).
			First(ctx); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, repository.ErrCampNotFound
			}

			return nil, err
		}
	}

	return rollCalls, nil
}
