package router

import (
	"errors"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestAdminPostRoomGroup(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		req := api.AdminPostRoomGroupJSONRequestBody{
			Name: random.AlphaNumericString(t, 20),
		}
		username := random.AlphaNumericString(t, 32)
		campID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil)
		h.repo.MockRoomGroupRepository.EXPECT().
			CreateRoomGroup(gomock.Any(), gomock.Any()).
			Return(nil)

		res := h.expect.POST("/api/admin/camps/{campId}/room-groups", campID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusCreated).JSON().Object()

		res.Keys().ContainsAll("id", "name", "rooms")
		res.Value("name").String().IsEqual(req.Name)
		res.Value("rooms").Array().Length().IsEqual(0)
	})

	t.Run("Forbidden", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		req := api.AdminPostRoomGroupJSONRequestBody{
			Name: random.AlphaNumericString(t, 20),
		}
		username := random.AlphaNumericString(t, 32)
		campID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: false}, nil)

		h.expect.POST("/api/admin/camps/{campId}/room-groups", campID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusForbidden)
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		username := random.AlphaNumericString(t, 32)
		campID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil)

		h.expect.POST("/api/admin/camps/{campId}/room-groups", campID).
			WithJSON("invalid json").
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusBadRequest)
	})

	t.Run("CreateRoomGroup Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		req := api.AdminPostRoomGroupJSONRequestBody{
			Name: random.AlphaNumericString(t, 20),
		}
		username := random.AlphaNumericString(t, 32)
		campID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil)
		h.repo.MockRoomGroupRepository.EXPECT().
			CreateRoomGroup(gomock.Any(), gomock.Any()).
			Return(errors.New("database error"))

		h.expect.POST("/api/admin/camps/{campId}/room-groups", campID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("GetOrCreateUser Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		req := api.AdminPostRoomGroupJSONRequestBody{
			Name: random.AlphaNumericString(t, 20),
		}
		username := random.AlphaNumericString(t, 32)
		campID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(nil, errors.New("user error"))

		h.expect.POST("/api/admin/camps/{campId}/room-groups", campID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("Camp Not Found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		req := api.AdminPostRoomGroupJSONRequestBody{
			Name: random.AlphaNumericString(t, 20),
		}
		username := random.AlphaNumericString(t, 32)
		campID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil)
		h.repo.MockRoomGroupRepository.EXPECT().
			CreateRoomGroup(gomock.Any(), gomock.Any()).
			Return(repository.ErrCampNotFound)

		h.expect.POST("/api/admin/camps/{campId}/room-groups", campID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusNotFound)
	})
}

func TestAdminPutRoomGroup(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		req := api.AdminPutRoomGroupJSONRequestBody{
			Name: random.AlphaNumericString(t, 20),
		}
		username := random.AlphaNumericString(t, 32)
		roomGroupID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil)
		h.repo.MockRoomGroupRepository.EXPECT().
			UpdateRoomGroup(gomock.Any(), uint(roomGroupID), gomock.Any()).
			Return(nil)
		h.repo.MockRoomGroupRepository.EXPECT().
			GetRoomGroupByID(gomock.Any(), uint(roomGroupID)).
			Return(&model.RoomGroup{
				Model: gorm.Model{
					ID: uint(roomGroupID),
				},
				Name:  req.Name,
				Rooms: []model.Room{},
			}, nil)

		res := h.expect.PUT("/api/admin/room-groups/{roomGroupId}", roomGroupID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusOK).JSON().Object()

		res.Keys().ContainsAll("id", "name", "rooms")
		res.Value("id").Number().IsEqual(roomGroupID)
		res.Value("name").String().IsEqual(req.Name)
	})

	t.Run("Forbidden", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		req := api.AdminPutRoomGroupJSONRequestBody{
			Name: random.AlphaNumericString(t, 20),
		}
		username := random.AlphaNumericString(t, 32)
		roomGroupID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: false}, nil)

		h.expect.PUT("/api/admin/room-groups/{roomGroupId}", roomGroupID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusForbidden)
	})

	t.Run("Room Group Not Found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		req := api.AdminPutRoomGroupJSONRequestBody{
			Name: random.AlphaNumericString(t, 20),
		}
		username := random.AlphaNumericString(t, 32)
		roomGroupID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil)
		h.repo.MockRoomGroupRepository.EXPECT().
			UpdateRoomGroup(gomock.Any(), uint(roomGroupID), gomock.Any()).
			Return(repository.ErrRoomGroupNotFound)

		h.expect.PUT("/api/admin/room-groups/{roomGroupId}", roomGroupID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusNotFound)
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		username := random.AlphaNumericString(t, 32)
		roomGroupID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil)

		h.expect.PUT("/api/admin/room-groups/{roomGroupId}", roomGroupID).
			WithJSON("invalid json").
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusBadRequest)
	})

	t.Run("UpdateRoomGroup Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		req := api.AdminPutRoomGroupJSONRequestBody{
			Name: random.AlphaNumericString(t, 20),
		}
		username := random.AlphaNumericString(t, 32)
		roomGroupID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil)
		h.repo.MockRoomGroupRepository.EXPECT().
			UpdateRoomGroup(gomock.Any(), uint(roomGroupID), gomock.Any()).
			Return(errors.New("database error"))

		h.expect.PUT("/api/admin/room-groups/{roomGroupId}", roomGroupID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("GetOrCreateUser Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		req := api.AdminPutRoomGroupJSONRequestBody{
			Name: random.AlphaNumericString(t, 20),
		}
		username := random.AlphaNumericString(t, 32)
		roomGroupID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(nil, errors.New("user error"))

		h.expect.PUT("/api/admin/room-groups/{roomGroupId}", roomGroupID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusInternalServerError)
	})
}
