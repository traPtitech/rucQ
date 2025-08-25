//go:generate go tool mockgen -source=$GOFILE -destination=mockrepository/$GOFILE -package=mockrepository
package repository

import (
	"context"
	"errors"

	"github.com/traPtitech/rucQ/model"
)

var ErrMessageNotFound = errors.New("message not found")

type MessageRepository interface {
	// CreateMessage メッセージをデータベースに作成します
	CreateMessage(ctx context.Context, message *model.Message) error
	// GetReadyToSendMessages 送信予定時刻を過ぎた未送信のメッセージを取得します
	GetReadyToSendMessages(ctx context.Context) ([]model.Message, error)
	// UpdateMessage メッセージの情報を更新します
	UpdateMessage(ctx context.Context, messageID uint, message *model.Message) error
}
