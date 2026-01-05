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

func TestServer_GetRoomGroups(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		roomGroups := []model.RoomGroup{
			{
				Model: gorm.Model{ID: uint(random.PositiveInt(t))},
				Name:  random.AlphaNumericString(t, 20),
				Rooms: []model.Room{
					{
						Model: gorm.Model{ID: uint(random.PositiveInt(t))},
						Name:  random.AlphaNumericString(t, 15),
						Members: []model.User{
							{
								ID:      random.AlphaNumericString(t, 32),
								IsStaff: random.Bool(t),
							},
						},
					},
				},
				CampID: uint(campID),
			},
			{
				Model:  gorm.Model{ID: uint(random.PositiveInt(t))},
				Name:   random.AlphaNumericString(t, 20),
				Rooms:  []model.Room{},
				CampID: uint(campID),
			},
		}

		h.repo.MockRoomGroupRepository.EXPECT().
			GetRoomGroups(gomock.Any(), uint(campID)).
			Return(roomGroups, nil)

		res := h.expect.GET("/api/camps/{campId}/room-groups", campID).
			Expect().
			Status(http.StatusOK).JSON().Array()

		res.Length().IsEqual(2)

		firstGroup := res.Value(0).Object()
		firstGroup.Keys().ContainsAll("id", "name", "rooms")
		firstGroup.Value("id").Number().IsEqual(roomGroups[0].ID)
		firstGroup.Value("name").String().IsEqual(roomGroups[0].Name)
		firstGroup.Value("rooms").Array().Length().IsEqual(1)

		room := firstGroup.Value("rooms").Array().Value(0).Object()
		room.Keys().ContainsAll("id", "name", "members")
		room.Value("id").Number().IsEqual(roomGroups[0].Rooms[0].ID)
		room.Value("name").String().IsEqual(roomGroups[0].Rooms[0].Name)
		room.Value("members").Array().Length().IsEqual(1)

		member := room.Value("members").Array().Value(0).Object()
		member.Keys().ContainsAll("id", "isStaff")
		member.Value("id").String().IsEqual(roomGroups[0].Rooms[0].Members[0].ID)
		member.Value("isStaff").Boolean().IsEqual(roomGroups[0].Rooms[0].Members[0].IsStaff)

		secondGroup := res.Value(1).Object()
		secondGroup.Keys().ContainsAll("id", "name", "rooms")
		secondGroup.Value("id").Number().IsEqual(roomGroups[1].ID)
		secondGroup.Value("name").String().IsEqual(roomGroups[1].Name)
		secondGroup.Value("rooms").Array().Length().IsEqual(0)
	})

	t.Run("Camp Not Found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		campID := random.PositiveInt(t)

		h.repo.MockRoomGroupRepository.EXPECT().
			GetRoomGroups(gomock.Any(), uint(campID)).
			Return(nil, repository.ErrCampNotFound)

		h.expect.GET("/api/camps/{campId}/room-groups", campID).
			Expect().
			Status(http.StatusNotFound).JSON().Object().
			Value("message").String().IsEqual("Camp not found")
	})

	t.Run("Repository Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		campID := random.PositiveInt(t)

		h.repo.MockRoomGroupRepository.EXPECT().
			GetRoomGroups(gomock.Any(), uint(campID)).
			Return(nil, errors.New("database error"))

		h.expect.GET("/api/camps/{campId}/room-groups", campID).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("Empty Room Groups", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		campID := random.PositiveInt(t)
		roomGroups := []model.RoomGroup{}

		h.repo.MockRoomGroupRepository.EXPECT().
			GetRoomGroups(gomock.Any(), uint(campID)).
			Return(roomGroups, nil)

		res := h.expect.GET("/api/camps/{campId}/room-groups", campID).
			Expect().
			Status(http.StatusOK).JSON().Array()

		res.Length().IsEqual(0)
	})
}

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

	t.Run("UserAlreadyAssigned", func(t *testing.T) {
		t.Parallel()
		h := setup(t)
		username := random.AlphaNumericString(t, 32)
		campID := random.PositiveInt(t)

		req := api.AdminPostRoomGroupJSONRequestBody{
			Name: "Duplicate Group",
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil).Times(1)

		h.repo.MockRoomGroupRepository.EXPECT().
			CreateRoomGroup(gomock.Any(), gomock.Any()).
			Return(repository.ErrUserAlreadyAssigned).Times(1)

		h.expect.POST("/api/admin/camps/{campId}/room-groups", campID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object().
			Value("message").String().
			IsEqual("Some users are already assigned to another room in this camp")
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

func TestServer_AdminDeleteRoomGroup(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		username := random.AlphaNumericString(t, 32)
		roomGroupID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil).
			Times(1)
		h.repo.MockRoomGroupRepository.EXPECT().
			DeleteRoomGroup(gomock.Any(), uint(roomGroupID)).
			Return(nil).
			Times(1)

		h.expect.DELETE("/api/admin/room-groups/{roomGroupId}", roomGroupID).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusNoContent)
	})

	t.Run("Non-staff user", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		username := random.AlphaNumericString(t, 32)
		roomGroupID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: false}, nil).
			Times(1)

		h.expect.DELETE("/api/admin/room-groups/{roomGroupId}", roomGroupID).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusForbidden)
	})

	t.Run("Room group not found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		username := random.AlphaNumericString(t, 32)
		roomGroupID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil).
			Times(1)
		h.repo.MockRoomGroupRepository.EXPECT().
			DeleteRoomGroup(gomock.Any(), uint(roomGroupID)).
			Return(repository.ErrRoomGroupNotFound).
			Times(1)

		h.expect.DELETE("/api/admin/room-groups/{roomGroupId}", roomGroupID).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusNotFound)
	})

	t.Run("Delete error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		username := random.AlphaNumericString(t, 32)
		roomGroupID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil).
			Times(1)
		h.repo.MockRoomGroupRepository.EXPECT().
			DeleteRoomGroup(gomock.Any(), uint(roomGroupID)).
			Return(errors.New("database error")).
			Times(1)

		h.expect.DELETE("/api/admin/room-groups/{roomGroupId}", roomGroupID).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("GetOrCreateUser Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		username := random.AlphaNumericString(t, 32)
		roomGroupID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(nil, errors.New("user error")).
			Times(1)

		h.expect.DELETE("/api/admin/room-groups/{roomGroupId}", roomGroupID).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusInternalServerError)
	})
}
