package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository/mockrepository"
	"github.com/traPtitech/rucQ/service/mockservice"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestSchedulerService_processReadyMessages(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mockrepository.NewMockRepository(ctrl)
		mockTraq := mockservice.NewMockTraqService(ctrl)

		messages := []model.Message{
			{
				TargetUserID: random.AlphaNumericString(t, 20),
				Content:      random.AlphaNumericString(t, 100),
				SendAt:       time.Now().Add(-time.Hour),
			},
		}
		messages[0].ID = 1

		// GetReadyToSendMessagesが呼ばれることを期待
		mockRepo.MockMessageRepository.EXPECT().
			GetReadyToSendMessages(gomock.Any()).
			Return(messages, nil).
			Times(1)

		// メッセージ送信が呼ばれることを期待
		mockTraq.EXPECT().PostDirectMessage(
			gomock.Any(),
			messages[0].TargetUserID,
			messages[0].Content,
		).Return(nil).Times(1)

		// メッセージ更新が呼ばれることを期待
		mockRepo.MockMessageRepository.EXPECT().
			UpdateMessage(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		scheduler := NewSchedulerService(mockRepo, mockTraq)
		scheduler.processReadyMessages(context.Background())
	})

	t.Run("Send Message Error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mockrepository.NewMockRepository(ctrl)
		mockTraq := mockservice.NewMockTraqService(ctrl)

		messages := []model.Message{
			{
				TargetUserID: random.AlphaNumericString(t, 20),
				Content:      random.AlphaNumericString(t, 100),
				SendAt:       time.Now().Add(-time.Hour),
			},
		}
		messages[0].ID = 1

		mockRepo.MockMessageRepository.EXPECT().
			GetReadyToSendMessages(gomock.Any()).
			Return(messages, nil).
			Times(1)
		mockTraq.EXPECT().PostDirectMessage(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(errors.New("send error")).Times(1)

		// 送信に失敗した場合、UpdateMessageは呼ばれない
		mockRepo.MockMessageRepository.EXPECT().UpdateMessage(gomock.Any(), gomock.Any()).Times(0)

		scheduler := NewSchedulerService(mockRepo, mockTraq)
		scheduler.processReadyMessages(context.Background())
	})

	t.Run("Get Messages Error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mockrepository.NewMockRepository(ctrl)
		mockTraq := mockservice.NewMockTraqService(ctrl)

		mockRepo.MockMessageRepository.EXPECT().GetReadyToSendMessages(gomock.Any()).
			Return(nil, errors.New("database error")).Times(1)

		// エラーの場合、他のメソッドは呼ばれない
		mockTraq.EXPECT().PostDirectMessage(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		mockRepo.MockMessageRepository.EXPECT().UpdateMessage(gomock.Any(), gomock.Any()).Times(0)

		scheduler := NewSchedulerService(mockRepo, mockTraq)
		scheduler.processReadyMessages(context.Background())
	})
}

func TestNewSchedulerService(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepository.NewMockRepository(ctrl)
	mockTraq := mockservice.NewMockTraqService(ctrl)

	scheduler := NewSchedulerService(mockRepo, mockTraq)
	assert.NotNil(t, scheduler)

	// 型アサーションでintervalが正しく設定されていることを確認
	schedulerImpl := scheduler
	assert.Equal(t, time.Minute, schedulerImpl.interval)
}
