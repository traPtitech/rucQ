package gormrepository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestRepository_CreateMessage(t *testing.T) {
	t.Parallel()

	r := setup(t)

	// 先にユーザーを作成
	userID := random.AlphaNumericString(t, 20)
	_, err := r.GetOrCreateUser(t.Context(), userID)
	require.NoError(t, err)

	message := &model.Message{
		TargetUserID: userID,
		Content:      random.AlphaNumericString(t, 100),
		SendAt:       time.Now().Add(time.Hour),
	}

	err = r.CreateMessage(t.Context(), message)
	require.NoError(t, err)
	assert.NotZero(t, message.ID)
}

func TestRepository_GetReadyToSendMessages(t *testing.T) {
	t.Parallel()

	r := setup(t)

	// 共通のユーザーIDを作成
	userID1 := random.AlphaNumericString(t, 20)
	userID2 := random.AlphaNumericString(t, 20)
	userID3 := random.AlphaNumericString(t, 20)
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
	require.NoError(t, err)

	// 過去の送信予定時刻で未送信のメッセージのみ取得されることを確認
	assert.Len(t, messages, 1)
	assert.Equal(t, pastMessage.ID, messages[0].ID)
}

func TestRepository_UpdateMessage(t *testing.T) {
	t.Parallel()

	r := setup(t)

	// 先にユーザーを作成
	userID := random.AlphaNumericString(t, 20)
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

	err = r.UpdateMessage(t.Context(), message)
	require.NoError(t, err)

	// 更新されたメッセージが送信済みとして扱われることを確認
	messages, err := r.GetReadyToSendMessages(t.Context())
	require.NoError(t, err)
	assert.Len(t, messages, 0)
}
