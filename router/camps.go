package router

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/converter"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
	"github.com/traPtitech/rucQ/service"
)

func (s *Server) GetCamps(e echo.Context) error {
	camps, err := s.repo.GetCamps()

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get camps",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := converter.Convert[[]api.CampResponse](camps)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert camps",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, response)
}

// AdminPostCamp 新規キャンプ追加
// (POST /api/admin/camps)
func (s *Server) AdminPostCamp(e echo.Context, params api.AdminPostCampParams) error {
	var req api.AdminPostCampJSONRequestBody

	if err := e.Bind(&req); err != nil {
		slog.WarnContext(
			e.Request().Context(),
			"failed to bind request body",
			slog.String("error", err.Error()),
		)

		return err
	}

	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create user",
			slog.String("error", err.Error()),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		slog.WarnContext(
			e.Request().Context(),
			"user is not a staff member",
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	campModel, err := converter.Convert[model.Camp](req)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert request to model",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.CreateCamp(&campModel); err != nil {
		if errors.Is(err, model.ErrAlreadyExists) {
			return echo.NewHTTPError(http.StatusConflict, "Camp already exists")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to create camp",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := converter.Convert[api.CampResponse](campModel)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert camp to response",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, &response)
}

func (s *Server) GetCamp(e echo.Context, campID api.CampId) error {
	camp, err := s.repo.GetCampByID(e.Request().Context(), uint(campID))

	if err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"camp not found",
				slog.Int("campId", int(campID)),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to get camp",
			slog.String("error", err.Error()),
			slog.Int("campId", int(campID)),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := converter.Convert[api.CampResponse](camp)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert camp to response",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, &response)
}

// AdminPutCamp キャンプ情報編集
// (PUT /api/admin/camps/{campId})
func (s *Server) AdminPutCamp(
	e echo.Context,
	campID api.CampId,
	params api.AdminPutCampParams,
) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create user",
			slog.String("error", err.Error()),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		slog.WarnContext(
			e.Request().Context(),
			"user is not a staff member",
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPutCampJSONRequestBody

	if err := e.Bind(&req); err != nil {
		slog.WarnContext(
			e.Request().Context(),
			"failed to bind request body",
			slog.String("error", err.Error()),
		)

		return err
	}

	newCamp, err := converter.Convert[model.Camp](req)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert request to model",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	newCamp.ID = uint(campID)

	if err := s.repo.UpdateCamp(e.Request().Context(), uint(campID), &newCamp); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"camp not found",
				slog.Int("campId", int(campID)),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to update camp",
			slog.String("error", err.Error()),
			slog.Int("campId", int(campID)),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := converter.Convert[api.CampResponse](newCamp)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert camp to response",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, &response)
}

// AdminDeleteCamp キャンプ削除
// (DELETE /api/admin/camps/{campId})
func (s *Server) AdminDeleteCamp(
	e echo.Context,
	campID api.CampId,
	params api.AdminDeleteCampParams,
) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create user",
			slog.String("error", err.Error()),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		slog.WarnContext(
			e.Request().Context(),
			"user is not a staff member",
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	// TODO: Not found エラーのハンドリングを追加
	if err := s.repo.DeleteCamp(e.Request().Context(), uint(campID)); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to delete camp",
			slog.String("error", err.Error()),
			slog.Int("campId", int(campID)),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.NoContent(http.StatusNoContent)
}

// PostCampRegister 合宿に登録
func (s *Server) PostCampRegister(
	e echo.Context,
	campID api.CampId,
	params api.PostCampRegisterParams,
) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create user",
			slog.String("error", err.Error()),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	camp, err := s.repo.GetCampByID(e.Request().Context(), uint(campID))
	if err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"camp not found",
				slog.Int("campId", int(campID)),
			)
			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get camp",
			slog.String("error", err.Error()),
			slog.Int("campId", int(campID)),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !camp.IsRegistrationOpen {
		slog.WarnContext(
			e.Request().Context(),
			"registration for camp is closed",
			slog.Int("campId", int(campID)),
			slog.String("userId", *params.XForwardedUser),
		)
		return echo.NewHTTPError(http.StatusForbidden, "Registration for this camp is closed")
	}

	if err := s.repo.AddCampParticipant(e.Request().Context(), uint(campID), user); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"camp not found",
				slog.Int("campId", int(campID)),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to add camp participant",
			slog.String("error", err.Error()),
			slog.Int("campId", int(campID)),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.NoContent(http.StatusNoContent)
}

// DeleteCampRegister 合宿登録を取り消し
func (s *Server) DeleteCampRegister(
	e echo.Context,
	campID api.CampId,
	params api.DeleteCampRegisterParams,
) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create user",
			slog.String("error", err.Error()),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	camp, err := s.repo.GetCampByID(e.Request().Context(), uint(campID))

	if err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"camp not found",
				slog.Int("campId", int(campID)),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to get camp",
			slog.String("error", err.Error()),
			slog.Int("campId", int(campID)),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !camp.IsRegistrationOpen {
		slog.WarnContext(
			e.Request().Context(),
			"registration for camp is closed",
			slog.Int("campId", int(campID)),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusForbidden, "Registration for this camp is closed")
	}

	if err := s.repo.RemoveCampParticipant(e.Request().Context(), uint(campID), user); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to remove camp participant",
			slog.String("error", err.Error()),
			slog.Int("campId", int(campID)),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.NoContent(http.StatusNoContent)
}

// GetCampParticipants 合宿の参加者一覧を取得
func (s *Server) GetCampParticipants(e echo.Context, campID api.CampId) error {
	participants, err := s.repo.GetCampParticipants(e.Request().Context(), uint(campID))

	// TODO: Not found エラーのハンドリングを追加（APIスキーマにも追加する）
	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get camp participants",
			slog.String("error", err.Error()),
			slog.Int("campId", int(campID)),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response []api.UserResponse

	if err := copier.Copy(&response, &participants); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to copy participants",
			slog.String("error", err.Error()),
			slog.Int("campId", int(campID)),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, response)
}

// AdminAddCampParticipant ユーザーをキャンプに参加させる（管理者用）
func (s *Server) AdminAddCampParticipant(
	e echo.Context,
	campID api.CampId,
	params api.AdminAddCampParticipantParams,
) error {
	operator, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create operator",
			slog.String("error", err.Error()),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !operator.IsStaff {
		slog.WarnContext(
			e.Request().Context(),
			"user is not a staff member",
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminAddCampParticipantJSONRequestBody

	if err := e.Bind(&req); err != nil {
		slog.WarnContext(
			e.Request().Context(),
			"failed to bind request body",
			slog.String("error", err.Error()),
		)

		return err
	}

	targetUserName, err := s.traqService.GetCanonicalUserName(e.Request().Context(), req.UserId)

	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"user not found",
				slog.String("userId", req.UserId),
			)

			return echo.NewHTTPError(http.StatusNotFound, "User not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to get canonical user name",
			slog.String("error", err.Error()),
			slog.String("userId", req.UserId),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	targetUser, err := s.repo.GetOrCreateUser(e.Request().Context(), targetUserName)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create target user",
			slog.String("error", err.Error()),
			slog.String("userId", req.UserId),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	camp, err := s.repo.GetCampByID(e.Request().Context(), uint(campID))

	if err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"camp not found",
				slog.Int("campId", campID),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to get camp",
			slog.String("error", err.Error()),
			slog.Int("campId", campID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.AddCampParticipant(e.Request().Context(), uint(campID), targetUser); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"camp not found",
				slog.Int("campId", campID),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to add camp participant",
			slog.String("error", err.Error()),
			slog.Int("campId", campID),
			slog.String("userId", req.UserId),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	go func() {
		ctx := context.WithoutCancel(e.Request().Context())
		message := fmt.Sprintf("@%sがあなたを%sに追加しました", operator.ID, camp.Name)

		if err := s.traqService.PostDirectMessage(ctx, targetUser.ID, message); err != nil {
			slog.ErrorContext(
				ctx,
				"failed to post direct message",
				slog.String("error", err.Error()),
				slog.String("userId", targetUser.ID),
			)
		}
	}()

	return e.NoContent(http.StatusNoContent)
}

// AdminRemoveCampParticipant ユーザーをキャンプから削除する（管理者用）
func (s *Server) AdminRemoveCampParticipant(
	e echo.Context,
	campID api.CampId,
	userID api.UserId,
	params api.AdminRemoveCampParticipantParams,
) error {
	operator, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create operator",
			slog.String("error", err.Error()),
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !operator.IsStaff {
		slog.WarnContext(
			e.Request().Context(),
			"user is not a staff member",
			slog.String("userId", *params.XForwardedUser),
		)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	targetUser, err := s.repo.GetOrCreateUser(e.Request().Context(), string(userID))
	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create target user",
			slog.String("error", err.Error()),
			slog.String("userId", userID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	camp, err := s.repo.GetCampByID(e.Request().Context(), uint(campID))

	if err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"camp not found",
				slog.Int("campId", campID),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to get camp",
			slog.String("error", err.Error()),
			slog.Int("campId", campID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.RemoveCampParticipant(e.Request().Context(), uint(campID), targetUser); err != nil {
		switch {
		case errors.Is(err, repository.ErrCampNotFound):
			slog.WarnContext(
				e.Request().Context(),
				"camp not found",
				slog.Int("campId", campID),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")

		case errors.Is(err, repository.ErrParticipantNotFound):
			slog.WarnContext(
				e.Request().Context(),
				"participant not found",
				slog.Int("campId", campID),
				slog.String("userId", userID),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Participant not found")

		default:
			slog.ErrorContext(
				e.Request().Context(),
				"failed to remove camp participant",
				slog.String("error", err.Error()),
				slog.Int("campId", campID),
				slog.String("userId", userID),
			)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}
	}

	go func() {
		ctx := context.WithoutCancel(e.Request().Context())
		message := fmt.Sprintf("@%sがあなたを%sから削除しました", operator.ID, camp.Name)

		if err := s.traqService.PostDirectMessage(ctx, targetUser.ID, message); err != nil {
			slog.ErrorContext(
				ctx,
				"failed to post direct message",
				slog.String("error", err.Error()),
				slog.String("userId", targetUser.ID),
			)
		}
	}()

	return e.NoContent(http.StatusNoContent)
}
