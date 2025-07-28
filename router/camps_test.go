package router

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/oapi-codegen/runtime/types"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
	"github.com/traPtitech/rucQ/testutil/random"
)

func TestGetCamps(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		dateStart := random.Time(t)
		dateEnd := dateStart.Add(time.Duration(random.PositiveInt(t)))
		camp := model.Camp{
			Model: gorm.Model{
				ID: uint(random.PositiveInt(t)),
			},
			DisplayID:          random.AlphaNumericString(t, 10),
			Name:               random.AlphaNumericString(t, 20),
			Guidebook:          random.AlphaNumericString(t, 100),
			IsDraft:            random.Bool(t),
			IsPaymentOpen:      random.Bool(t),
			IsRegistrationOpen: random.Bool(t),
			DateStart:          dateStart,
			DateEnd:            dateEnd,
		}

		h.repo.MockCampRepository.EXPECT().GetCamps().Return([]model.Camp{camp}, nil)

		res := h.expect.GET("/api/camps").Expect().Status(http.StatusOK).JSON().Array()

		res.Length().IsEqual(1)

		val := res.Value(0).Object()

		val.Keys().ContainsOnly(
			"id", "displayId", "name", "guidebook", "isDraft", "isPaymentOpen",
			"isRegistrationOpen", "dateStart", "dateEnd")
		val.Value("id").Number().IsEqual(camp.ID)
		val.Value("displayId").String().IsEqual(camp.DisplayID)
		val.Value("name").String().IsEqual(camp.Name)
		val.Value("guidebook").String().IsEqual(camp.Guidebook)
		val.Value("isDraft").Boolean().IsEqual(camp.IsDraft)
		val.Value("isPaymentOpen").Boolean().IsEqual(camp.IsPaymentOpen)
		val.Value("isRegistrationOpen").Boolean().IsEqual(camp.IsRegistrationOpen)
		val.Value("dateStart").String().IsEqual(camp.DateStart.Format(time.DateOnly))
		val.Value("dateEnd").String().IsEqual(camp.DateEnd.Format(time.DateOnly))
	})
}

func TestAdminPostCamp(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		dateStart := random.Time(t)
		dateEnd := dateStart.Add(time.Duration(random.PositiveInt(t)))
		req := api.AdminPostCampJSONRequestBody{
			DisplayId:          random.AlphaNumericString(t, 10),
			Name:               random.AlphaNumericString(t, 20),
			Guidebook:          random.AlphaNumericString(t, 100),
			IsDraft:            random.Bool(t),
			IsPaymentOpen:      random.Bool(t),
			IsRegistrationOpen: random.Bool(t),
			DateStart:          types.Date{Time: dateStart},
			DateEnd:            types.Date{Time: dateEnd},
		}
		username := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil)
		h.repo.MockCampRepository.EXPECT().CreateCamp(gomock.Any()).Return(nil)

		res := h.expect.POST("/api/admin/camps").
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusCreated).JSON().Object()

		res.Keys().ContainsOnly(
			"id", "displayId", "name", "guidebook", "isDraft", "isPaymentOpen",
			"isRegistrationOpen", "dateStart", "dateEnd")
		res.Value("displayId").String().IsEqual(req.DisplayId)
		res.Value("name").String().IsEqual(req.Name)
		res.Value("guidebook").String().IsEqual(req.Guidebook)
		res.Value("isDraft").Boolean().IsEqual(req.IsDraft)
		res.Value("isPaymentOpen").Boolean().IsEqual(req.IsPaymentOpen)
		res.Value("isRegistrationOpen").Boolean().IsEqual(req.IsRegistrationOpen)
		res.Value("dateStart").String().IsEqual(req.DateStart.Format(time.DateOnly))
		res.Value("dateEnd").String().IsEqual(req.DateEnd.Format(time.DateOnly))
	})
}

func TestAdminPutCamp(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := api.CampId(random.PositiveInt(t))
		dateStart := random.Time(t)
		dateEnd := dateStart.Add(time.Duration(random.PositiveInt(t)))
		req := api.AdminPutCampJSONRequestBody{
			DisplayId:          random.AlphaNumericString(t, 10),
			Name:               random.AlphaNumericString(t, 20),
			Guidebook:          random.AlphaNumericString(t, 100),
			IsDraft:            random.Bool(t),
			IsPaymentOpen:      random.Bool(t),
			IsRegistrationOpen: random.Bool(t),
			DateStart:          types.Date{Time: dateStart},
			DateEnd:            types.Date{Time: dateEnd},
		}
		username := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil)
		h.repo.MockCampRepository.EXPECT().
			UpdateCamp(gomock.Any(), uint(campID), gomock.Any()).
			Return(nil)

		res := h.expect.PUT("/api/admin/camps/{campId}", campID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusOK).JSON().Object()

		res.Keys().ContainsOnly(
			"id", "displayId", "name", "guidebook", "isDraft", "isPaymentOpen",
			"isRegistrationOpen", "dateStart", "dateEnd")
		res.Value("id").Number().IsEqual(campID)
		res.Value("displayId").String().IsEqual(req.DisplayId)
		res.Value("name").String().IsEqual(req.Name)
		res.Value("guidebook").String().IsEqual(req.Guidebook)
		res.Value("isDraft").Boolean().IsEqual(req.IsDraft)
		res.Value("isPaymentOpen").Boolean().IsEqual(req.IsPaymentOpen)
		res.Value("isRegistrationOpen").Boolean().IsEqual(req.IsRegistrationOpen)
		res.Value("dateStart").String().IsEqual(req.DateStart.Format(time.DateOnly))
		res.Value("dateEnd").String().IsEqual(req.DateEnd.Format(time.DateOnly))
	})

	t.Run("Unauthorized", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := api.CampId(random.PositiveInt(t))
		dateStart := random.Time(t)
		dateEnd := dateStart.Add(time.Duration(random.PositiveInt(t)))
		req := api.AdminPutCampJSONRequestBody{
			DisplayId:          random.AlphaNumericString(t, 10),
			Name:               random.AlphaNumericString(t, 20),
			Guidebook:          random.AlphaNumericString(t, 100),
			IsDraft:            random.Bool(t),
			IsPaymentOpen:      random.Bool(t),
			IsRegistrationOpen: random.Bool(t),
			DateStart:          types.Date{Time: dateStart},
			DateEnd:            types.Date{Time: dateEnd},
		}
		username := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: false}, nil)

		h.expect.PUT("/api/admin/camps/{campId}", campID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusForbidden)
	})

	t.Run("Bad Request", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := api.CampId(random.PositiveInt(t))
		username := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil)

		h.expect.PUT("/api/admin/camps/{campId}", campID).
			WithJSON("invalid json").
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusBadRequest)
	})

	t.Run("UpdateCamp Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := api.CampId(random.PositiveInt(t))
		dateStart := random.Time(t)
		dateEnd := dateStart.Add(time.Duration(random.PositiveInt(t)))
		req := api.AdminPutCampJSONRequestBody{
			DisplayId:          random.AlphaNumericString(t, 10),
			Name:               random.AlphaNumericString(t, 20),
			Guidebook:          random.AlphaNumericString(t, 100),
			IsDraft:            random.Bool(t),
			IsPaymentOpen:      random.Bool(t),
			IsRegistrationOpen: random.Bool(t),
			DateStart:          types.Date{Time: dateStart},
			DateEnd:            types.Date{Time: dateEnd},
		}
		username := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil)
		h.repo.MockCampRepository.EXPECT().
			UpdateCamp(gomock.Any(), uint(campID), gomock.Any()).
			Return(errors.New("update error"))

		h.expect.PUT("/api/admin/camps/{campId}", campID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("GetOrCreateUser Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := api.CampId(random.PositiveInt(t))
		dateStart := random.Time(t)
		dateEnd := dateStart.Add(time.Duration(random.PositiveInt(t)))
		req := api.AdminPutCampJSONRequestBody{
			DisplayId:          random.AlphaNumericString(t, 10),
			Name:               random.AlphaNumericString(t, 20),
			Guidebook:          random.AlphaNumericString(t, 100),
			IsDraft:            random.Bool(t),
			IsPaymentOpen:      random.Bool(t),
			IsRegistrationOpen: random.Bool(t),
			DateStart:          types.Date{Time: dateStart},
			DateEnd:            types.Date{Time: dateEnd},
		}
		username := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(nil, errors.New("user error"))

		h.expect.PUT("/api/admin/camps/{campId}", campID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("Camp Not Found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := api.CampId(random.PositiveInt(t))
		dateStart := random.Time(t)
		dateEnd := dateStart.Add(time.Duration(random.PositiveInt(t)))
		req := api.AdminPutCampJSONRequestBody{
			DisplayId:          random.AlphaNumericString(t, 10),
			Name:               random.AlphaNumericString(t, 20),
			Guidebook:          random.AlphaNumericString(t, 100),
			IsDraft:            random.Bool(t),
			IsPaymentOpen:      random.Bool(t),
			IsRegistrationOpen: random.Bool(t),
			DateStart:          types.Date{Time: dateStart},
			DateEnd:            types.Date{Time: dateEnd},
		}
		username := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(&model.User{IsStaff: true}, nil)
		h.repo.MockCampRepository.EXPECT().
			UpdateCamp(gomock.Any(), uint(campID), gomock.Any()).
			Return(model.ErrNotFound)

		h.expect.PUT("/api/admin/camps/{campId}", campID).
			WithJSON(req).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			HasValue("message", "Camp not found")
	})
}

func TestPostCampRegister(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := api.CampId(random.PositiveInt(t))
		username := random.AlphaNumericString(t, 32)
		user := &model.User{
			ID:      username,
			IsStaff: false,
		}
		camp := &model.Camp{
			Model: gorm.Model{
				ID: uint(campID),
			},
			IsRegistrationOpen: true,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(user, nil)
		h.repo.MockCampRepository.EXPECT().
			GetCampByID(uint(campID)).
			Return(camp, nil)
		h.repo.MockCampRepository.EXPECT().
			AddCampParticipant(gomock.Any(), uint(campID), user).
			Return(nil)

		h.expect.POST("/api/camps/{campId}/register", campID).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusNoContent)
	})

	t.Run("GetOrCreateUser Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := api.CampId(random.PositiveInt(t))
		username := random.AlphaNumericString(t, 32)

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(nil, errors.New("user error"))

		h.expect.POST("/api/camps/{campId}/register", campID).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("GetCampByID Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := api.CampId(random.PositiveInt(t))
		username := random.AlphaNumericString(t, 32)
		user := &model.User{
			ID:      username,
			IsStaff: false,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(user, nil)
		h.repo.MockCampRepository.EXPECT().
			GetCampByID(uint(campID)).
			Return(nil, errors.New("camp error"))

		h.expect.POST("/api/camps/{campId}/register", campID).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusInternalServerError)
	})

	t.Run("Camp Not Found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := api.CampId(random.PositiveInt(t))
		username := random.AlphaNumericString(t, 32)
		user := &model.User{
			ID:      username,
			IsStaff: false,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(user, nil)
		h.repo.MockCampRepository.EXPECT().
			GetCampByID(uint(campID)).
			Return(nil, model.ErrNotFound)

		h.expect.POST("/api/camps/{campId}/register", campID).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			HasValue("message", "Camp not found")
	})

	t.Run("Registration Closed", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		username := random.AlphaNumericString(t, 32)
		user := &model.User{
			ID:      username,
			IsStaff: false,
		}
		camp := &model.Camp{
			Model: gorm.Model{
				ID: uint(campID),
			},
			IsRegistrationOpen: false,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(user, nil)
		h.repo.MockCampRepository.EXPECT().
			GetCampByID(uint(campID)).
			Return(camp, nil)

		h.expect.POST("/api/camps/{campId}/register", campID).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusForbidden).
			JSON().
			Object().
			HasValue("message", "Registration for this camp is closed")
	})

	t.Run("AddCampParticipant Error", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := api.CampId(random.PositiveInt(t))
		username := random.AlphaNumericString(t, 32)
		user := &model.User{
			ID:      username,
			IsStaff: false,
		}
		camp := &model.Camp{
			Model: gorm.Model{
				ID: uint(campID),
			},
			IsRegistrationOpen: true,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(user, nil)
		h.repo.MockCampRepository.EXPECT().
			GetCampByID(uint(campID)).
			Return(camp, nil)
		h.repo.MockCampRepository.EXPECT().
			AddCampParticipant(gomock.Any(), uint(campID), user).
			Return(errors.New("participant error"))

		h.expect.POST("/api/camps/{campId}/register", campID).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusInternalServerError)
	})
}

func TestAdminAddCampParticipant(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		adminUsername := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		admin := &model.User{
			ID:      adminUsername,
			IsStaff: true,
		}
		targetUser := &model.User{
			ID:      targetUserID,
			IsStaff: random.Bool(t),
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUsername).
			Return(admin, nil)
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), targetUserID).
			Return(targetUser, nil)
		h.repo.MockCampRepository.EXPECT().
			AddCampParticipant(gomock.Any(), uint(campID), targetUser).
			Return(nil)

		req := api.AdminAddCampParticipantJSONRequestBody{
			UserId: targetUserID,
		}

		h.expect.POST("/api/admin/camps/{campId}/participants", campID).
			WithHeader("X-Forwarded-User", adminUsername).
			WithJSON(req).
			Expect().
			Status(http.StatusNoContent)
	})

	t.Run("Non-Staff User", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		username := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		user := &model.User{
			ID:      username,
			IsStaff: false,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(user, nil)

		req := api.AdminAddCampParticipantJSONRequestBody{
			UserId: targetUserID,
		}

		h.expect.POST("/api/admin/camps/{campId}/participants", campID).
			WithHeader("X-Forwarded-User", username).
			WithJSON(req).
			Expect().
			Status(http.StatusForbidden).
			JSON().
			Object().
			HasValue("message", "Forbidden")
	})

	t.Run("Camp Not Found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		adminUsername := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		admin := &model.User{
			ID:      adminUsername,
			IsStaff: true,
		}
		targetUser := &model.User{
			ID:      targetUserID,
			IsStaff: random.Bool(t),
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUsername).
			Return(admin, nil)
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), targetUserID).
			Return(targetUser, nil)
		h.repo.MockCampRepository.EXPECT().
			AddCampParticipant(gomock.Any(), uint(campID), targetUser).
			Return(model.ErrNotFound)

		req := api.AdminAddCampParticipantJSONRequestBody{
			UserId: targetUserID,
		}

		h.expect.POST("/api/admin/camps/{campId}/participants", campID).
			WithHeader("X-Forwarded-User", adminUsername).
			WithJSON(req).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			HasValue("message", "Camp not found")
	})
}

func TestAdminRemoveCampParticipant(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		adminUsername := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		admin := &model.User{
			ID:      adminUsername,
			IsStaff: true,
		}
		targetUser := &model.User{
			ID:      targetUserID,
			IsStaff: random.Bool(t),
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUsername).
			Return(admin, nil)
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), targetUserID).
			Return(targetUser, nil)
		h.repo.MockCampRepository.EXPECT().
			RemoveCampParticipant(gomock.Any(), uint(campID), targetUser).
			Return(nil)

		h.expect.DELETE("/api/admin/camps/{campId}/participants/{userId}", campID, targetUserID).
			WithHeader("X-Forwarded-User", adminUsername).
			Expect().
			Status(http.StatusNoContent)
	})

	t.Run("Non-Staff User", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		username := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		user := &model.User{
			ID:      username,
			IsStaff: false,
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), username).
			Return(user, nil)

		h.expect.DELETE("/api/admin/camps/{campId}/participants/{userId}", campID, targetUserID).
			WithHeader("X-Forwarded-User", username).
			Expect().
			Status(http.StatusForbidden).
			JSON().
			Object().
			HasValue("message", "Forbidden")
	})

	t.Run("Camp Not Found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		adminUsername := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		admin := &model.User{
			ID:      adminUsername,
			IsStaff: true,
		}
		targetUser := &model.User{
			ID:      targetUserID,
			IsStaff: random.Bool(t),
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUsername).
			Return(admin, nil)
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), targetUserID).
			Return(targetUser, nil)
		h.repo.MockCampRepository.EXPECT().
			RemoveCampParticipant(gomock.Any(), uint(campID), targetUser).
			Return(repository.ErrCampNotFound)

		h.expect.DELETE("/api/admin/camps/{campId}/participants/{userId}", campID, targetUserID).
			WithHeader("X-Forwarded-User", adminUsername).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			HasValue("message", "Camp not found")
	})

	t.Run("Participant Not Found", func(t *testing.T) {
		t.Parallel()

		h := setup(t)
		campID := random.PositiveInt(t)
		adminUsername := random.AlphaNumericString(t, 32)
		targetUserID := random.AlphaNumericString(t, 32)
		admin := &model.User{
			ID:      adminUsername,
			IsStaff: true,
		}
		targetUser := &model.User{
			ID:      targetUserID,
			IsStaff: random.Bool(t),
		}

		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), adminUsername).
			Return(admin, nil)
		h.repo.MockUserRepository.EXPECT().
			GetOrCreateUser(gomock.Any(), targetUserID).
			Return(targetUser, nil)
		h.repo.MockCampRepository.EXPECT().
			RemoveCampParticipant(gomock.Any(), uint(campID), targetUser).
			Return(repository.ErrParticipantNotFound)

		h.expect.DELETE("/api/admin/camps/{campId}/participants/{userId}", campID, targetUserID).
			WithHeader("X-Forwarded-User", adminUsername).
			Expect().
			Status(http.StatusNotFound).
			JSON().
			Object().
			HasValue("message", "Participant not found")
	})
}
