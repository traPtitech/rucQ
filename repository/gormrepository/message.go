package gormrepository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
)

func (r *Repository) CreateMessage(ctx context.Context, message *model.Message) error {
	return gorm.G[model.Message](r.db).Create(ctx, message)
}

func (r *Repository) GetReadyToSendMessages(ctx context.Context) ([]model.Message, error) {
	messages, err := gorm.G[model.Message](r.db).
		Where("sent_at IS NULL").
		Where("send_at <= ?", time.Now()).
		Find(ctx)

	return messages, err
}

func (r *Repository) UpdateMessage(ctx context.Context, message *model.Message) error {
	_, err := gorm.G[*model.Message](r.db).Updates(ctx, message)

	return err
}
