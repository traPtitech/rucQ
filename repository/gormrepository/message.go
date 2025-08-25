package gormrepository

import (
	"context"
	"time"

	"github.com/traPtitech/rucQ/model"
)

func (r *Repository) CreateMessage(ctx context.Context, message *model.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *Repository) GetReadyToSendMessages(ctx context.Context) ([]model.Message, error) {
	var messages []model.Message
	err := r.db.WithContext(ctx).
		Where("sent_at IS NULL").
		Where("send_at <= ?", time.Now()).
		Find(&messages).Error
	return messages, err
}

func (r *Repository) UpdateMessage(ctx context.Context, message *model.Message) error {
	return r.db.WithContext(ctx).Save(message).Error
}
