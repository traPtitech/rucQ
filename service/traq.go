//go:generate go tool mockgen -source=$GOFILE -destination=mockservice/$GOFILE -package=mockservice
package service

import (
	"context"
)

// TraqService はtraQ APIとの連携を担当するサービスです。
type TraqService interface {
	// PostDirectMessage は指定したユーザーにダイレクトメッセージを送信します。
	PostDirectMessage(ctx context.Context, userID string, content string) error
}
