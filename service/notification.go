//go:generate go tool mockgen -source=$GOFILE -destination=mockservice/$GOFILE -package=mockservice
package service

import (
	"context"

	"github.com/traPtitech/rucQ/model"
)

type NotificationService interface {
	// 未回答だった場合oldAnswerはnil
	SendAnswerChangeMessage(
		ctx context.Context,
		editorUserID string,
		oldAnswer *model.Answer,
		newAnswer model.Answer,
	) error
}
