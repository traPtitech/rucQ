package service

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository/mockrepository"
	"github.com/traPtitech/rucQ/service/mockservice"
	"github.com/traPtitech/rucQ/testutil/random"
)

type schedulerTestSetup struct {
	scheduler *schedulerServiceImpl
	mockRepo  *mockrepository.MockRepository
	mockTraq  *mockservice.MockTraqService
}

func setupSchedulerTest(t *testing.T) *schedulerTestSetup {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockRepo := mockrepository.NewMockRepository(ctrl)
	mockTraq := mockservice.NewMockTraqService(ctrl)
	scheduler := NewSchedulerService(mockRepo, mockTraq)

	return &schedulerTestSetup{
		scheduler: scheduler,
		mockRepo:  mockRepo,
		mockTraq:  mockTraq,
	}
}

func TestSchedulerServiceImpl_processReadyMessages(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		s := setupSchedulerTest(t)

		messages := []model.Message{
			{
				Model:        gorm.Model{ID: uint(random.PositiveInt(t))},
				TargetUserID: random.AlphaNumericString(t, 32),
				Content:      random.AlphaNumericString(t, 100),
				SendAt:       time.Now().Add(-time.Hour),
			},
		}

		// GetReadyToSendMessagesが呼ばれることを期待
		s.mockRepo.MockMessageRepository.EXPECT().
			GetReadyToSendMessages(gomock.Any()).
			Return(messages, nil).
			Times(1)

		// メッセージ送信が呼ばれることを期待
		s.mockTraq.EXPECT().PostDirectMessage(
			gomock.Any(),
			messages[0].TargetUserID,
			messages[0].Content,
		).Return(nil).Times(1)

		// メッセージ更新が呼ばれることを期待
		s.mockRepo.MockMessageRepository.EXPECT().
			UpdateMessage(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		s.scheduler.processReadyMessages(t.Context())
	})

	t.Run("Send Message Error", func(t *testing.T) {
		t.Parallel()

		s := setupSchedulerTest(t)

		messages := []model.Message{
			{
				Model:        gorm.Model{ID: uint(random.PositiveInt(t))},
				TargetUserID: random.AlphaNumericString(t, 32),
				Content:      random.AlphaNumericString(t, 100),
				SendAt:       time.Now().Add(-time.Hour),
			},
		}

		s.mockRepo.MockMessageRepository.EXPECT().
			GetReadyToSendMessages(gomock.Any()).
			Return(messages, nil).
			Times(1)
		s.mockTraq.EXPECT().PostDirectMessage(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(errors.New("send error")).Times(1)

		// 送信に失敗した場合、UpdateMessageは呼ばれない
		s.mockRepo.MockMessageRepository.EXPECT().
			UpdateMessage(gomock.Any(), gomock.Any()).
			Times(0)

		s.scheduler.processReadyMessages(t.Context())
	})

	t.Run("Get Messages Error", func(t *testing.T) {
		t.Parallel()

		setup := setupSchedulerTest(t)

		setup.mockRepo.MockMessageRepository.EXPECT().GetReadyToSendMessages(gomock.Any()).
			Return(nil, errors.New("database error")).Times(1)

		// エラーの場合、他のメソッドは呼ばれない
		setup.mockTraq.EXPECT().PostDirectMessage(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		setup.mockRepo.MockMessageRepository.EXPECT().
			UpdateMessage(gomock.Any(), gomock.Any()).
			Times(0)

		setup.scheduler.processReadyMessages(t.Context())
	})
}

func TestSchedulerServiceImpl_Start(t *testing.T) {
	t.Parallel()

	t.Run("Normal Operation", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			s := setupSchedulerTest(t)

			// processReadyMessagesが2回呼ばれることを期待（2回のtick）
			s.mockRepo.MockMessageRepository.EXPECT().
				GetReadyToSendMessages(gomock.Any()).
				Return([]model.Message{}, nil).
				Times(2)

			// Startを別のgoroutineで実行
			ctx, cancel := context.WithCancel(t.Context())
			go s.scheduler.Start(ctx)

			// 1分間進める（1回目のtick）
			time.Sleep(time.Minute)
			synctest.Wait()

			// さらに1分間進める（2回目のtick）
			time.Sleep(time.Minute)
			synctest.Wait()

			// コンテキストをキャンセルしてスケジューラーを停止
			cancel()
			synctest.Wait()
		})
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			s := setupSchedulerTest(t)

			// processReadyMessagesが呼ばれないことを期待（即座にキャンセル）
			s.mockRepo.MockMessageRepository.EXPECT().
				GetReadyToSendMessages(gomock.Any()).
				Return([]model.Message{}, nil).
				Times(0)

			// Startを別のgoroutineで実行
			ctx, cancel := context.WithCancel(t.Context())
			go s.scheduler.Start(ctx)

			// すぐにキャンセルする
			cancel()
			synctest.Wait()
		})
	})

	t.Run("Message Processing During Ticks", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			s := setupSchedulerTest(t)

			messages := []model.Message{
				{
					Model:        gorm.Model{ID: uint(random.PositiveInt(t))},
					TargetUserID: random.AlphaNumericString(t, 32),
					Content:      random.AlphaNumericString(t, 100),
					SendAt:       time.Now().Add(-time.Hour),
				},
			}

			// 最初のtickでメッセージを返す
			s.mockRepo.MockMessageRepository.EXPECT().
				GetReadyToSendMessages(gomock.Any()).
				Return(messages, nil).
				Times(1)

			// メッセージ送信が成功することを期待
			s.mockTraq.EXPECT().
				PostDirectMessage(gomock.Any(), messages[0].TargetUserID, messages[0].Content).
				Return(nil).
				Times(1)

			// メッセージ更新が呼ばれることを期待
			s.mockRepo.MockMessageRepository.EXPECT().
				UpdateMessage(gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)

			// 2回目のtickではメッセージなし
			s.mockRepo.MockMessageRepository.EXPECT().
				GetReadyToSendMessages(gomock.Any()).
				Return([]model.Message{}, nil).
				Times(1)

			// Startを別のgoroutineで実行
			ctx, cancel := context.WithCancel(t.Context())
			go s.scheduler.Start(ctx)

			// 1分間進める（1回目のtick - メッセージ処理）
			time.Sleep(time.Minute)
			synctest.Wait()

			// 1分間進める（2回目のtick - メッセージなし）
			time.Sleep(time.Minute)
			synctest.Wait()

			// コンテキストをキャンセルしてスケジューラーを停止
			cancel()
			synctest.Wait()
		})
	})
}
