//go:generate go tool mockgen -source=$GOFILE -destination=mocktraq/traq.go -package=mocktraq
package traq

import (
	"context"
	"errors"
)

var ErrUserNotFound = errors.New("user not found")

// TraqService はtraQ APIとの連携を担当するサービスです。
type TraqService interface {
	GetCanonicalUserName(ctx context.Context, userID string) (string, error)
	// PostDirectMessage は指定したユーザーにダイレクトメッセージを送信します。
	PostDirectMessage(ctx context.Context, userID string, content string) error
}
