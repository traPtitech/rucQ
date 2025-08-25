package router

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestServer_AdminPostMessage(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		userId := random.AlphaNumericString(t, 32)
		targetUserId := random.AlphaNumericString(t, 32)
		content := random.AlphaNumericString(t, 100)
		sendAt := random.Time(t)

		staffUser := model.User{
			ID:      userId,
			IsStaff: true,
		}

		// スタッフユーザーの取得をモック
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userId).
			Return(&staffUser, nil).
			Times(1)

		// メッセージ作成をモック
		h.repo.MockMessageRepository.EXPECT().
			CreateMessage(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ any, message *model.Message) error {
				// 引数をチェック
				assert.Equal(t, targetUserId, message.TargetUserID)
				assert.Equal(t, content, message.Content)
				assert.True(t, sendAt.Equal(message.SendAt))
				assert.Nil(t, message.SentAt)

				return nil
			}).
			Times(1)

		req := api.AdminPostMessageJSONRequestBody{
			Content: content,
			SendAt:  sendAt,
		}

		h.expect.POST("/api/admin/users/{userId}/messages", targetUserId).
			WithHeader("X-Forwarded-User", userId).
			WithJSON(req).
			Expect().
			Status(http.StatusAccepted)
	})

	t.Run("Non-staff user", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		userId := random.AlphaNumericString(t, 32)
		targetUserId := random.AlphaNumericString(t, 32)

		nonStaffUser := model.User{
			ID:      userId,
			IsStaff: false,
		}

		// 非スタッフユーザーの取得をモック
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userId).
			Return(&nonStaffUser, nil).
			Times(1)

		req := api.AdminPostMessageJSONRequestBody{
			Content: random.AlphaNumericString(t, 100),
			SendAt:  random.Time(t),
		}

		h.expect.POST("/api/admin/users/{userId}/messages", targetUserId).
			WithHeader("X-Forwarded-User", userId).
			WithJSON(req).
			Expect().
			Status(http.StatusForbidden)
	})

	t.Run("User repository error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		userId := random.AlphaNumericString(t, 32)
		targetUserId := random.AlphaNumericString(t, 32)

		// ユーザー取得エラーをモック
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userId).
			Return(nil, errors.New("database error")).
			Times(1)

		req := api.AdminPostMessageJSONRequestBody{
			Content: random.AlphaNumericString(t, 100),
			SendAt:  random.Time(t),
		}

		h.expect.POST("/api/admin/users/{userId}/messages", targetUserId).
			WithHeader("X-Forwarded-User", userId).
			WithJSON(req).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("Target user not found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		userId := random.AlphaNumericString(t, 32)
		targetUserId := random.AlphaNumericString(t, 32)

		staffUser := model.User{
			ID:      userId,
			IsStaff: true,
		}

		// スタッフユーザーの取得をモック
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userId).
			Return(&staffUser, nil).
			Times(1)

		// メッセージ作成でターゲットユーザーが見つからないエラーをモック
		h.repo.MockMessageRepository.EXPECT().
			CreateMessage(gomock.Any(), gomock.Any()).
			Return(repository.ErrUserNotFound).
			Times(1)

		req := api.AdminPostMessageJSONRequestBody{
			Content: random.AlphaNumericString(t, 100),
			SendAt:  random.Time(t),
		}

		h.expect.POST("/api/admin/users/{userId}/messages", targetUserId).
			WithHeader("X-Forwarded-User", userId).
			WithJSON(req).
			Expect().
			Status(http.StatusNotFound)
	})

	t.Run("Database error during message creation", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		userId := random.AlphaNumericString(t, 32)
		targetUserId := random.AlphaNumericString(t, 32)

		staffUser := model.User{
			ID:      userId,
			IsStaff: true,
		}

		// スタッフユーザーの取得をモック
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), userId).
			Return(&staffUser, nil).
			Times(1)

		// メッセージ作成でデータベースエラーをモック
		h.repo.MockMessageRepository.EXPECT().
			CreateMessage(gomock.Any(), gomock.Any()).
			Return(errors.New("database error")).
			Times(1)

		req := api.AdminPostMessageJSONRequestBody{
			Content: random.AlphaNumericString(t, 100),
			SendAt:  random.Time(t),
		}

		h.expect.POST("/api/admin/users/{userId}/messages", targetUserId).
			WithHeader("X-Forwarded-User", userId).
			WithJSON(req).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("Invalid JSON request", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		userId := random.AlphaNumericString(t, 32)
		targetUserId := random.AlphaNumericString(t, 32)

		// JSONパース段階で失敗するため、モックは呼ばれない
		// 不正なJSONを送信
		h.expect.POST("/api/admin/users/{userId}/messages", targetUserId).
			WithHeader("X-Forwarded-User", userId).
			WithText("invalid json").
			WithHeader("Content-Type", "application/json").
			Expect().
			Status(http.StatusBadRequest)
	})
}
