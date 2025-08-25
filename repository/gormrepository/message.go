package gormrepository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
)

func (r *Repository) CreateMessage(ctx context.Context, message *model.Message) error {
	err := gorm.G[model.Message](r.db).Create(ctx, message)

	if err != nil {
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			return repository.ErrUserNotFound
		}

		return err
	}

	return nil
}

func (r *Repository) GetReadyToSendMessages(ctx context.Context) ([]model.Message, error) {
	messages, err := gorm.G[model.Message](r.db).
		Where("sent_at IS NULL").
		Where("send_at <= ?", time.Now()).
		Find(ctx)

	return messages, err
}

func (r *Repository) UpdateMessage(
	ctx context.Context,
	messageID uint,
	message *model.Message,
) error {
	if messageID == 0 {
		return repository.ErrMessageNotFound
	}

	rowsAffected, err := gorm.G[*model.Message](r.db).
		Where("id = ?", messageID).Updates(ctx, message)

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrMessageNotFound
	}

	return nil
}
