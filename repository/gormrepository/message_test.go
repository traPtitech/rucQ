package gormrepository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestRepository_CreateMessage(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		// 先にユーザーを作成
		userID := random.AlphaNumericString(t, 32)
		_, err := r.GetOrCreateUser(t.Context(), userID)
		require.NoError(t, err)

		message := &model.Message{
			TargetUserID: userID,
			Content:      random.AlphaNumericString(t, 100),
			SendAt:       time.Now().Add(time.Hour),
		}

		err = r.CreateMessage(t.Context(), message)
		assert.NoError(t, err)
		assert.NotZero(t, message.ID)
	})

	t.Run("Foreign Key Violation - User Not Found", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		// 存在しないユーザーIDでメッセージを作成
		message := &model.Message{
			TargetUserID: random.AlphaNumericString(t, 32),
			Content:      random.AlphaNumericString(t, 100),
			SendAt:       time.Now().Add(time.Hour),
		}

		err := r.CreateMessage(t.Context(), message)
		assert.ErrorIs(t, err, repository.ErrUserNotFound)
		assert.Zero(t, message.ID)
	})

	t.Run("Context Cancelled", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		// 先にユーザーを作成
		userID := random.AlphaNumericString(t, 32)
		_, err := r.GetOrCreateUser(t.Context(), userID)
		require.NoError(t, err)

		message := &model.Message{
			TargetUserID: userID,
			Content:      random.AlphaNumericString(t, 100),
			SendAt:       time.Now().Add(time.Hour),
		}

		// キャンセルされたコンテキストを使用
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		err = r.CreateMessage(ctx, message)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestRepository_GetReadyToSendMessages(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		// 共通のユーザーIDを作成
		userID1 := random.AlphaNumericString(t, 32)
		userID2 := random.AlphaNumericString(t, 32)
		userID3 := random.AlphaNumericString(t, 32)
		_, err := r.GetOrCreateUser(t.Context(), userID1)
		require.NoError(t, err)
		_, err = r.GetOrCreateUser(t.Context(), userID2)
		require.NoError(t, err)
		_, err = r.GetOrCreateUser(t.Context(), userID3)
		require.NoError(t, err)

		// 過去の送信予定時刻のメッセージを作成
		pastMessage := &model.Message{
			TargetUserID: userID1,
			Content:      random.AlphaNumericString(t, 100),
			SendAt:       time.Now().Add(-time.Hour),
		}
		err = r.CreateMessage(t.Context(), pastMessage)
		require.NoError(t, err)

		// 未来の送信予定時刻のメッセージを作成
		futureMessage := &model.Message{
			TargetUserID: userID2,
			Content:      random.AlphaNumericString(t, 100),
			SendAt:       time.Now().Add(time.Hour),
		}
		err = r.CreateMessage(t.Context(), futureMessage)
		require.NoError(t, err)

		// 送信済みのメッセージを作成
		sentTime := time.Now()
		sentMessage := &model.Message{
			TargetUserID: userID3,
			Content:      random.AlphaNumericString(t, 100),
			SendAt:       time.Now().Add(-time.Hour),
			SentAt:       &sentTime,
		}
		err = r.CreateMessage(t.Context(), sentMessage)
		require.NoError(t, err)

		messages, err := r.GetReadyToSendMessages(t.Context())
		assert.NoError(t, err)

		// 過去の送信予定時刻で未送信のメッセージのみ取得されることを確認
		assert.Len(t, messages, 1)
		assert.Equal(t, pastMessage.ID, messages[0].ID)
	})

	t.Run("No Messages Ready", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		messages, err := r.GetReadyToSendMessages(t.Context())
		assert.NoError(t, err)
		assert.Empty(t, messages)
	})

	t.Run("Context Cancelled", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		// キャンセルされたコンテキストを使用
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		messages, err := r.GetReadyToSendMessages(ctx)
		assert.ErrorIs(t, err, context.Canceled)
		assert.Nil(t, messages)
	})
}

func TestRepository_UpdateMessage(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		// 先にユーザーを作成
		userID := random.AlphaNumericString(t, 32)
		_, err := r.GetOrCreateUser(t.Context(), userID)
		require.NoError(t, err)

		message := &model.Message{
			TargetUserID: userID,
			Content:      random.AlphaNumericString(t, 100),
			SendAt:       time.Now().Add(-time.Hour),
		}
		err = r.CreateMessage(t.Context(), message)
		require.NoError(t, err)

		// 送信時刻を設定
		sentTime := time.Now()
		message.SentAt = &sentTime

		err = r.UpdateMessage(t.Context(), message.ID, message)
		require.NoError(t, err)

		// 更新されたメッセージが送信済みとして扱われることを確認
		messages, err := r.GetReadyToSendMessages(t.Context())
		assert.NoError(t, err)
		assert.Len(t, messages, 0)
	})

	t.Run("Message Not Found", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		// 存在しないメッセージを更新しようとする
		message := &model.Message{}
		nonExistentID := uint(random.PositiveInt(t))

		err := r.UpdateMessage(t.Context(), nonExistentID, message)
		assert.Equal(t, repository.ErrMessageNotFound, err)
	})

	t.Run("Context Cancelled", func(t *testing.T) {
		t.Parallel()

		r := setup(t)

		// 先にユーザーを作成
		userID := random.AlphaNumericString(t, 32)
		_, err := r.GetOrCreateUser(t.Context(), userID)
		require.NoError(t, err)

		message := &model.Message{
			TargetUserID: userID,
			Content:      random.AlphaNumericString(t, 100),
			SendAt:       time.Now().Add(-time.Hour),
		}
		err = r.CreateMessage(t.Context(), message)
		require.NoError(t, err)

		// キャンセルされたコンテキストを使用
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		sentTime := time.Now()
		message.SentAt = &sentTime

		err = r.UpdateMessage(ctx, message.ID, message)
		assert.ErrorIs(t, err, context.Canceled)
	})
}
