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

func TestServer_AdminPostRoom(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		memberID1 := random.AlphaNumericString(t, 32)
		memberID2 := random.AlphaNumericString(t, 32)
		req := api.AdminPostRoomJSONRequestBody{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupId: random.PositiveInt(t),
			MemberIds: []string{
				memberID1,
				memberID2,
			},
		}
		username := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil).
			Times(1)

		roomID := uint(random.PositiveInt(t))

		h.repo.MockRoomRepository.EXPECT().
			CreateRoom(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ any, room *model.Room) error {
				room.ID = roomID
				return nil
			}).Times(1)
		h.repo.MockRoomRepository.EXPECT().
			GetRoomByID(roomID).
			Return(&model.Room{
				Model: gorm.Model{
					ID: roomID,
				},
				Name:        req.Name,
				RoomGroupID: uint(req.RoomGroupId),
				Members: []model.User{
					{
						ID:      memberID1,
						IsStaff: false,
					},
					{
						ID:      memberID2,
						IsStaff: true,
					},
				},
			}, nil).
			Times(1)

		res := h.expect.POST("/api/admin/rooms").
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusCreated).JSON().Object()

		res.Keys().ContainsOnly("id", "name", "members")
		res.Value("id").Number().IsEqual(roomID)
		res.Value("name").String().IsEqual(req.Name)
		res.Value("members").Array().Length().IsEqual(len(req.MemberIds))

		member1 := res.Value("members").Array().Value(0).Object()

		member1.Keys().ContainsOnly("id", "isStaff")
		member1.Value("id").String().IsEqual(memberID1)
		member1.Value("isStaff").Boolean().IsFalse()

		member2 := res.Value("members").Array().Value(1).Object()

		member2.Keys().ContainsOnly("id", "isStaff")
		member2.Value("id").String().IsEqual(memberID2)
		member2.Value("isStaff").Boolean().IsTrue()
	})

	t.Run("Success without members", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		req := api.AdminPostRoomJSONRequestBody{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupId: random.PositiveInt(t),
			MemberIds:   []string{},
		}
		username := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil).
			Times(1)

		roomID := uint(random.PositiveInt(t))

		h.repo.MockRoomRepository.EXPECT().
			CreateRoom(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ any, room *model.Room) error {
				room.ID = roomID
				return nil
			}).Times(1)
		h.repo.MockRoomRepository.EXPECT().
			GetRoomByID(roomID).
			Return(&model.Room{
				Model: gorm.Model{
					ID: roomID,
				},
				Name:        req.Name,
				RoomGroupID: uint(req.RoomGroupId),
				Members:     []model.User{},
			}, nil).
			Times(1)

		res := h.expect.POST("/api/admin/rooms").
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusCreated).JSON().Object()

		res.Keys().ContainsAll("id", "name", "members")
		res.Value("id").Number().IsEqual(roomID)
		res.Value("name").String().IsEqual(req.Name)
		res.Value("members").Array().Length().IsEqual(0)
	})

	t.Run("Forbidden - User is not staff", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		req := api.AdminPostRoomJSONRequestBody{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupId: random.PositiveInt(t),
			MemberIds:   []string{random.AlphaNumericString(t, 32)},
		}
		username := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: false}, nil).
			Times(1)

		h.expect.POST("/api/admin/rooms").
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusForbidden)
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		username := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil).
			Times(1)

		h.expect.POST("/api/admin/rooms").
			WithJSON("invalid json").
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusBadRequest)
	})

	t.Run("GetOrCreateUser Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		req := api.AdminPostRoomJSONRequestBody{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupId: random.PositiveInt(t),
			MemberIds:   []string{random.AlphaNumericString(t, 32)},
		}
		username := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(nil, errors.New("user error")).
			Times(1)

		h.expect.POST("/api/admin/rooms").
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("Invalid user or room group ID", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		req := api.AdminPostRoomJSONRequestBody{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupId: random.PositiveInt(t),
			MemberIds:   []string{random.AlphaNumericString(t, 32)},
		}
		username := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil).
			Times(1)
		h.repo.MockRoomRepository.EXPECT().
			CreateRoom(gomock.Any(), gomock.Any()).
			Return(repository.ErrUserOrRoomGroupNotFound).
			Times(1)

		h.expect.POST("/api/admin/rooms").
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusBadRequest).JSON().Object().
			Value("message").String().IsEqual("Invalid user or room group ID")
	})

	t.Run("CreateRoom Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		req := api.AdminPostRoomJSONRequestBody{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupId: random.PositiveInt(t),
			MemberIds:   []string{random.AlphaNumericString(t, 32)},
		}
		username := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil).
			Times(1)
		h.repo.MockRoomRepository.EXPECT().
			CreateRoom(gomock.Any(), gomock.Any()).
			Return(errors.New("database error")).
			Times(1)

		h.expect.POST("/api/admin/rooms").
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusInternalServerError)
	})
}

func TestServer_AdminPutRoom(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		req := api.AdminPutRoomJSONRequestBody{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupId: random.PositiveInt(t),
			MemberIds: []string{
				random.AlphaNumericString(t, 32),
				random.AlphaNumericString(t, 32),
			},
		}
		username := random.AlphaNumericString(t, 32)
		roomID := uint(random.PositiveInt(t))
		updatedRoom := &model.Room{
			Model:       gorm.Model{ID: roomID},
			Name:        req.Name,
			RoomGroupID: uint(req.RoomGroupId),
			Members: []model.User{
				{ID: req.MemberIds[0]},
				{ID: req.MemberIds[1]},
			},
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil).
			Times(1)
		h.repo.MockRoomRepository.EXPECT().
			UpdateRoom(gomock.Any(), roomID, gomock.Any()).
			DoAndReturn(func(_ any, _ uint, room *model.Room) error {
				*room = *updatedRoom
				return nil
			}).Times(1)
		h.repo.MockRoomRepository.EXPECT().
			GetRoomByID(roomID).
			Return(&model.Room{
				Model: gorm.Model{
					ID: roomID,
				},
				Name:        req.Name,
				RoomGroupID: uint(req.RoomGroupId),
				Members: []model.User{
					{
						ID:      req.MemberIds[0],
						IsStaff: true,
					},
					{
						ID:      req.MemberIds[1],
						IsStaff: false,
					},
				},
			}, nil).
			Times(1)

		res := h.expect.PUT("/api/admin/rooms/{roomId}", roomID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusOK).JSON().Object()

		res.Keys().ContainsOnly("id", "name", "members")
		res.Value("id").Number().IsEqual(roomID)
		res.Value("name").String().IsEqual(req.Name)
		res.Value("members").Array().Length().IsEqual(len(req.MemberIds))

		member1 := res.Value("members").Array().Value(0).Object()
		member1.Keys().ContainsOnly("id", "isStaff")
		member1.Value("id").String().IsEqual(req.MemberIds[0])
		member1.Value("isStaff").Boolean().IsTrue()

		member2 := res.Value("members").Array().Value(1).Object()
		member2.Keys().ContainsOnly("id", "isStaff")
		member2.Value("id").String().IsEqual(req.MemberIds[1])
		member2.Value("isStaff").Boolean().IsFalse()
	})

	t.Run("Success - Remove all members", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		req := api.AdminPutRoomJSONRequestBody{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupId: random.PositiveInt(t),
			MemberIds:   []string{},
		}
		username := random.AlphaNumericString(t, 32)
		roomID := uint(random.PositiveInt(t))

		updatedRoom := &model.Room{
			Model:       gorm.Model{ID: roomID},
			Name:        req.Name,
			RoomGroupID: uint(req.RoomGroupId),
			Members:     []model.User{},
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil).
			Times(1)
		h.repo.MockRoomRepository.EXPECT().
			UpdateRoom(gomock.Any(), uint(roomID), gomock.Any()).
			DoAndReturn(func(_ any, _ uint, room *model.Room) error {
				*room = *updatedRoom
				return nil
			}).Times(1)
		h.repo.MockRoomRepository.EXPECT().
			GetRoomByID(roomID).
			Return(&model.Room{
				Model: gorm.Model{
					ID: roomID,
				},
				Name:        req.Name,
				RoomGroupID: uint(req.RoomGroupId),
				Members:     []model.User{},
			}, nil).
			Times(1)

		res := h.expect.PUT("/api/admin/rooms/{roomId}", roomID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusOK).JSON().Object()

		res.Keys().ContainsAll("id", "name", "members")
		res.Value("id").Number().IsEqual(roomID)
		res.Value("name").String().IsEqual(req.Name)
		res.Value("members").Array().Length().IsEqual(0)
	})

	t.Run("Success - Change RoomGroup", func(t *testing.T) {
		t.Parallel()

		h := setup(t)

		req := api.AdminPutRoomJSONRequestBody{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupId: random.PositiveInt(t),
			MemberIds:   []string{random.AlphaNumericString(t, 32)},
		}
		username := random.AlphaNumericString(t, 32)
		roomID := uint(random.PositiveInt(t))

		updatedRoom := &model.Room{
			Model:       gorm.Model{ID: roomID},
			Name:        req.Name,
			RoomGroupID: uint(req.RoomGroupId),
			Members: []model.User{
				{ID: req.MemberIds[0]},
			},
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil).
			Times(1)
		h.repo.MockRoomRepository.EXPECT().
			UpdateRoom(gomock.Any(), uint(roomID), gomock.Any()).
			DoAndReturn(func(_ any, _ uint, room *model.Room) error {
				*room = *updatedRoom
				return nil
			}).Times(1)
		h.repo.MockRoomRepository.EXPECT().
			GetRoomByID(roomID).
			Return(&model.Room{
				Model: gorm.Model{
					ID: roomID,
				},
				Name:        req.Name,
				RoomGroupID: uint(req.RoomGroupId),
				Members: []model.User{
					{
						ID:      req.MemberIds[0],
						IsStaff: true,
					},
				},
			}, nil).
			Times(1)

		res := h.expect.PUT("/api/admin/rooms/{roomId}", roomID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusOK).JSON().Object()

		res.Keys().ContainsAll("id", "name", "members")
		res.Value("id").Number().IsEqual(roomID)
		res.Value("name").String().IsEqual(req.Name)
		res.Value("members").Array().Length().IsEqual(len(req.MemberIds))

		member := res.Value("members").Array().Value(0).Object()
		member.Keys().ContainsOnly("id", "isStaff")
		member.Value("id").String().IsEqual(req.MemberIds[0])
		member.Value("isStaff").Boolean().IsTrue()
	})

	t.Run("Forbidden - User is not staff", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		req := api.AdminPutRoomJSONRequestBody{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupId: random.PositiveInt(t),
			MemberIds:   []string{random.AlphaNumericString(t, 32)},
		}
		username := random.AlphaNumericString(t, 32)
		roomID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: false}, nil).
			Times(1)
		h.expect.PUT("/api/admin/rooms/{roomId}", roomID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusForbidden)
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		username := random.AlphaNumericString(t, 32)
		roomID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil).
			Times(1)
		h.expect.PUT("/api/admin/rooms/{roomId}", roomID).
			WithJSON("invalid json").
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusBadRequest)
	})

	t.Run("GetOrCreateUser Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		req := api.AdminPutRoomJSONRequestBody{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupId: random.PositiveInt(t),
			MemberIds:   []string{random.AlphaNumericString(t, 32)},
		}
		username := random.AlphaNumericString(t, 32)
		roomID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(nil, errors.New("user error")).
			Times(1)
		h.expect.PUT("/api/admin/rooms/{roomId}", roomID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("Room Not Found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		req := api.AdminPutRoomJSONRequestBody{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupId: random.PositiveInt(t),
			MemberIds:   []string{random.AlphaNumericString(t, 32)},
		}
		username := random.AlphaNumericString(t, 32)
		roomID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil).
			Times(1)
		h.repo.MockRoomRepository.EXPECT().
			UpdateRoom(gomock.Any(), uint(roomID), gomock.Any()).
			Return(repository.ErrRoomNotFound).
			Times(1)

		h.expect.PUT("/api/admin/rooms/{roomId}", roomID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusNotFound).JSON().Object().
			Value("message").String().IsEqual("Room not found")
	})

	t.Run("User Not Found in Members", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		req := api.AdminPutRoomJSONRequestBody{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupId: random.PositiveInt(t),
			MemberIds:   []string{random.AlphaNumericString(t, 32)},
		}
		username := random.AlphaNumericString(t, 32)
		roomID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil).
			Times(1)
		h.repo.MockRoomRepository.EXPECT().
			UpdateRoom(gomock.Any(), uint(roomID), gomock.Any()).
			Return(repository.ErrUserNotFound).
			Times(1)
		h.expect.PUT("/api/admin/rooms/{roomId}", roomID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusBadRequest).JSON().Object().
			Value("message").String().IsEqual("User not found")
	})

	t.Run("Room Group Not Found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		req := api.AdminPutRoomJSONRequestBody{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupId: random.PositiveInt(t),
			MemberIds:   []string{random.AlphaNumericString(t, 32)},
		}
		username := random.AlphaNumericString(t, 32)
		roomID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil).
			Times(1)
		h.repo.MockRoomRepository.EXPECT().
			UpdateRoom(gomock.Any(), uint(roomID), gomock.Any()).
			Return(repository.ErrRoomGroupNotFound).
			Times(1)
		h.expect.PUT("/api/admin/rooms/{roomId}", roomID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusBadRequest).JSON().Object().
			Value("message").String().IsEqual("Room group not found")
	})

	t.Run("UpdateRoom Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		req := api.AdminPutRoomJSONRequestBody{
			Name:        random.AlphaNumericString(t, 20),
			RoomGroupId: random.PositiveInt(t),
			MemberIds:   []string{random.AlphaNumericString(t, 32)},
		}
		username := random.AlphaNumericString(t, 32)
		roomID := random.PositiveInt(t)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil).
			Times(1)
		h.repo.MockRoomRepository.EXPECT().
			UpdateRoom(gomock.Any(), uint(roomID), gomock.Any()).
			Return(errors.New("database error")).
			Times(1)
		h.expect.PUT("/api/admin/rooms/{roomId}", roomID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusInternalServerError)
	})
}
