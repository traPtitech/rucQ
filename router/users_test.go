package router

import (
	"errors"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/service/traq"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestServer_AdminGetUser(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		canonicalUserID := random.AlphaNumericString(t, 32)

		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}
		targetUser := &model.User{
			ID:      canonicalUserID,
			IsStaff: random.Bool(t),
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)
		h.traqService.EXPECT().
			GetCanonicalUserName(gomock.Any(), targetUserID).
			Return(canonicalUserID, nil).
			Times(1)
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), canonicalUserID).
			Return(targetUser, nil).
			Times(1)

		res := h.expect.GET("/api/admin/users/{userId}", targetUserID).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		res.Keys().ContainsOnly("id", "isStaff")
		res.Value("id").String().IsEqual(targetUser.ID)
		res.Value("isStaff").Boolean().IsEqual(targetUser.IsStaff)
	})

	t.Run("Forbidden", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		nonAdminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)

		nonAdminUser := &model.User{
			ID:      nonAdminUserID,
			IsStaff: false,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), nonAdminUserID).
			Return(nonAdminUser, nil).
			Times(1)

		h.expect.GET("/api/admin/users/{userId}", targetUserID).
			WithHeader("X-Forwarded-User", nonAdminUserID).
			Expect().
			Status(http.StatusForbidden)
	})

	t.Run("User Not Found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)

		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)
		h.traqService.EXPECT().
			GetCanonicalUserName(gomock.Any(), targetUserID).
			Return("", traq.ErrUserNotFound).
			Times(1)

		h.expect.GET("/api/admin/users/{userId}", targetUserID).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().
			Status(http.StatusNotFound)
	})

	t.Run("Internal Server Error on GetOrCreateUser for operator", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(nil, errors.New("db error")).
			Times(1)

		h.expect.GET("/api/admin/users/{userId}", targetUserID).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("Internal Server Error on GetCanonicalUserName", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)

		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)
		h.traqService.EXPECT().
			GetCanonicalUserName(gomock.Any(), targetUserID).
			Return("", errors.New("traq api error")).
			Times(1)

		h.expect.GET("/api/admin/users/{userId}", targetUserID).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("Internal Server Error on GetOrCreateUser for target", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		canonicalUserID := random.AlphaNumericString(t, 32)

		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)
		h.traqService.EXPECT().
			GetCanonicalUserName(gomock.Any(), targetUserID).
			Return(canonicalUserID, nil).
			Times(1)
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), canonicalUserID).
			Return(nil, errors.New("db error")).
			Times(1)

		h.expect.GET("/api/admin/users/{userId}", targetUserID).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().
			Status(http.StatusInternalServerError)
	})
}

func TestServer_AdminPutUser(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)

		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}
		targetUser := &model.User{
			ID:      targetUserID,
			IsStaff: false,
		}

		req := api.UserRequest{
			IsStaff: true,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)
		h.traqService.EXPECT().
			GetCanonicalUserName(gomock.Any(), targetUserID).
			Return(targetUserID, nil).
			Times(1)
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), targetUserID).
			Return(targetUser, nil).
			Times(1)
		h.repo.MockUserRepository.EXPECT().
			UpdateUser(gomock.Any(), targetUser).
			Return(nil).
			Times(1)

		res := h.expect.PUT("/api/admin/users/{userId}", targetUserID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		res.Keys().ContainsOnly("id", "isStaff")
		res.Value("id").String().IsEqual(targetUser.ID)
		res.Value("isStaff").Boolean().IsTrue()
	})

	t.Run("Forbidden", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		nonAdminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)

		nonAdminUser := &model.User{
			ID:      nonAdminUserID,
			IsStaff: false,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), nonAdminUserID).
			Return(nonAdminUser, nil).
			Times(1)

		h.expect.PUT("/api/admin/users/{userId}", targetUserID).
			WithJSON(api.UserRequest{IsStaff: true}).
			WithHeader("X-Forwarded-User", nonAdminUserID).
			Expect().
			Status(http.StatusForbidden)
	})

	t.Run("User Not Found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)

		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)
		h.traqService.EXPECT().
			GetCanonicalUserName(gomock.Any(), targetUserID).
			Return("", traq.ErrUserNotFound).
			Times(1)

		h.expect.PUT("/api/admin/users/{userId}", targetUserID).
			WithJSON(api.UserRequest{IsStaff: true}).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().
			Status(http.StatusNotFound)
	})

	t.Run("Internal Server Error on GetOrCreateUser for operator", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(nil, errors.New("db error")).
			Times(1)

		h.expect.PUT("/api/admin/users/{userId}", targetUserID).
			WithJSON(api.UserRequest{IsStaff: true}).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("Internal Server Error on GetCanonicalUserName", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)

		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)
		h.traqService.EXPECT().
			GetCanonicalUserName(gomock.Any(), targetUserID).
			Return("", errors.New("traq api error")).
			Times(1)

		h.expect.PUT("/api/admin/users/{userId}", targetUserID).
			WithJSON(api.UserRequest{IsStaff: true}).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("Internal Server Error on GetOrCreateUser for target", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)

		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)
		h.traqService.EXPECT().
			GetCanonicalUserName(gomock.Any(), targetUserID).
			Return(targetUserID, nil).
			Times(1)
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), targetUserID).
			Return(nil, errors.New("db error")).
			Times(1)

		h.expect.PUT("/api/admin/users/{userId}", targetUserID).
			WithJSON(api.UserRequest{IsStaff: true}).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("Internal Server Error on UpdateUser", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		adminUserID := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)

		adminUser := &model.User{
			ID:      adminUserID,
			IsStaff: true,
		}
		targetUser := &model.User{
			ID:      targetUserID,
			IsStaff: false,
		}

		req := api.UserRequest{IsStaff: true}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUserID).
			Return(adminUser, nil).
			Times(1)
		h.traqService.EXPECT().
			GetCanonicalUserName(gomock.Any(), targetUserID).
			Return(targetUserID, nil).
			Times(1)
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), targetUserID).
			Return(targetUser, nil).
			Times(1)
		h.repo.MockUserRepository.EXPECT().
			UpdateUser(gomock.Any(), targetUser).
			Return(errors.New("db error")).
			Times(1)

		h.expect.PUT("/api/admin/users/{userId}", targetUserID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", adminUserID).
			Expect().
			Status(http.StatusInternalServerError)
	})
}
