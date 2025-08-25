//go:generate go tool mockgen -source=$GOFILE -destination=mockservice/$GOFILE -package=mockservice
package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
)

// SchedulerService はメッセージ送信スケジューリングを管理するサービスです
type SchedulerService interface {
	// Start はスケジューラーを開始します
	Start(ctx context.Context)
}

type schedulerServiceImpl struct {
	repo        repository.Repository
	traqService TraqService
	interval    time.Duration
}

// NewSchedulerService はSchedulerServiceの新しいインスタンスを作成します
func NewSchedulerService(
	repo repository.Repository,
	traqService TraqService,
) *schedulerServiceImpl {
	return &schedulerServiceImpl{
		repo:        repo,
		traqService: traqService,
		interval:    time.Minute, // 1分間隔でチェック
	}
}

// Start はメッセージ送信スケジューラーを開始します
func (s *schedulerServiceImpl) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	slog.InfoContext(ctx, "message scheduler started")

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "message scheduler stopped")
			return
		case <-ticker.C:
			s.processReadyMessages(ctx)
		}
	}
}

// processReadyMessages は送信準備が整ったメッセージを処理します
func (s *schedulerServiceImpl) processReadyMessages(ctx context.Context) {
	messages, err := s.repo.GetReadyToSendMessages(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get ready messages", slog.String("error", err.Error()))
		return
	}

	for _, message := range messages {
		if err := s.sendMessage(ctx, &message); err != nil {
			slog.ErrorContext(
				ctx,
				"failed to send message",
				slog.String("error", err.Error()),
				slog.Int("messageId", int(message.ID)),
				slog.String("targetUserId", message.TargetUserID),
			)
			continue
		}

		// 送信成功時刻を記録
		now := time.Now()
		message.SentAt = &now
		if err := s.repo.UpdateMessage(ctx, &message); err != nil {
			slog.ErrorContext(
				ctx,
				"failed to update message sent status",
				slog.String("error", err.Error()),
				slog.Int("messageId", int(message.ID)),
			)
		}
	}
}

// sendMessage は個別のメッセージを送信します
func (s *schedulerServiceImpl) sendMessage(ctx context.Context, message *model.Message) error {
	return s.traqService.PostDirectMessage(ctx, message.TargetUserID, message.Content)
}
