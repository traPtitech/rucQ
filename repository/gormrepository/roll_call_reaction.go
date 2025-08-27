package gormrepository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
)

func (r *Repository) CreateRollCallReaction(
	ctx context.Context,
	reaction *model.RollCallReaction,
) error {
	if err := r.db.WithContext(ctx).Create(reaction).Error; err != nil {
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			// 外部キーエラーが起きたときはRollCallが存在しないかどうかを確認
			rollCallExists, err := r.rollCallExists(ctx, reaction.RollCallID)

			if err != nil {
				return err
			}

			if !rollCallExists {
				return repository.ErrRollCallNotFound
			}

			return repository.ErrUserNotFound
		}

		return err
	}

	return nil
}

func (r *Repository) GetRollCallReactions(
	ctx context.Context,
	rollCallID uint,
) ([]model.RollCallReaction, error) {
	var reactions []model.RollCallReaction

	if err := r.db.WithContext(ctx).Where("roll_call_id = ?", rollCallID).Find(&reactions).Error; err != nil {
		return nil, err
	}

	// リアクションが見つからなかった場合、RollCallが存在しない可能性を考慮してRollCallの存在確認を行う
	if len(reactions) == 0 {
		rollCallExists, err := r.rollCallExists(ctx, rollCallID)

		if err != nil {
			return nil, err
		}

		if !rollCallExists {
			return nil, repository.ErrRollCallNotFound
		}
	}

	return reactions, nil
}

func (r *Repository) GetRollCallReactionByID(
	ctx context.Context,
	id uint,
) (*model.RollCallReaction, error) {
	var reaction model.RollCallReaction

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&reaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrRollCallReactionNotFound
		}

		return nil, err
	}

	return &reaction, nil
}

func (r *Repository) UpdateRollCallReaction(
	ctx context.Context,
	id uint,
	reaction *model.RollCallReaction,
) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Updates(reaction)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrRollCallReactionNotFound
	}

	return nil
}

func (r *Repository) DeleteRollCallReaction(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.RollCallReaction{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrRollCallReactionNotFound
	}

	return nil
}
