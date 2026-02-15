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
	"github.com/traPtitech/rucQ/service/traq"
)

func (s *Server) GetCamps(e echo.Context) error {
	camps, err := s.repo.GetCamps()

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get camps: %w", err))
	}

	response, err := converter.Convert[[]api.CampResponse](camps)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert camps: %w", err))
	}

	return e.JSON(http.StatusOK, response)
}

// AdminPostCamp 新規キャンプ追加
// (POST /api/admin/camps)
func (s *Server) AdminPostCamp(e echo.Context, params api.AdminPostCampParams) error {
	var req api.AdminPostCampJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user: %w", err))
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	campModel, err := converter.Convert[model.Camp](req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert request to model: %w", err))
	}

	if err := s.repo.CreateCamp(&campModel); err != nil {
		if errors.Is(err, model.ErrAlreadyExists) {
			return echo.NewHTTPError(http.StatusConflict, "Camp already exists")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to create camp: %w", err))
	}

	response, err := converter.Convert[api.CampResponse](campModel)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert camp to response: %w", err))
	}

	return e.JSON(http.StatusCreated, &response)
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
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user: %w", err))
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPutCampJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	newCamp, err := converter.Convert[model.Camp](req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert request to model: %w", err))
	}

	newCamp.ID = uint(campID)

	if err := s.repo.UpdateCamp(e.Request().Context(), uint(campID), &newCamp); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to update camp: %w", err))
	}

	response, err := converter.Convert[api.CampResponse](newCamp)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert camp to response: %w", err))
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
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user: %w", err))
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	// TODO: Not found エラーのハンドリングを追加
	if err := s.repo.DeleteCamp(e.Request().Context(), uint(campID)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to delete camp: %w", err))
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
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user: %w", err))
	}

	camp, err := s.repo.GetCampByID(e.Request().Context(), uint(campID))
	if err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get camp: %w", err))
	}

	if !camp.IsRegistrationOpen {
		return echo.NewHTTPError(http.StatusForbidden, "Registration for this camp is closed")
	}

	if err := s.repo.AddCampParticipant(e.Request().Context(), uint(campID), user); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to add camp participant: %w", err))
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
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user: %w", err))
	}

	camp, err := s.repo.GetCampByID(e.Request().Context(), uint(campID))

	if err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get camp: %w", err))
	}

	if !camp.IsRegistrationOpen {
		return echo.NewHTTPError(http.StatusForbidden, "Registration for this camp is closed")
	}

	if err := s.repo.RemoveCampParticipant(e.Request().Context(), uint(campID), user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to remove camp participant: %w", err))
	}

	return e.NoContent(http.StatusNoContent)
}

// GetCampParticipants 合宿の参加者一覧を取得
func (s *Server) GetCampParticipants(e echo.Context, campID api.CampId) error {
	participants, err := s.repo.GetCampParticipants(e.Request().Context(), uint(campID))

	// TODO: Not found エラーのハンドリングを追加（APIスキーマにも追加する）
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get camp participants: %w", err))
	}

	var response []api.UserResponse

	if err := copier.Copy(&response, &participants); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to copy participants: %w", err))
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
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create operator: %w", err))
	}

	if !operator.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminAddCampParticipantJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	targetUserName, err := s.traqService.GetCanonicalUserName(e.Request().Context(), req.UserId)

	if err != nil {
		if errors.Is(err, traq.ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "User not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get canonical user name: %w", err))
	}

	targetUser, err := s.repo.GetOrCreateUser(e.Request().Context(), targetUserName)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create target user: %w", err))
	}

	camp, err := s.repo.GetCampByID(e.Request().Context(), uint(campID))

	if err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get camp: %w", err))
	}

	if err := s.repo.AddCampParticipant(
		e.Request().Context(),
		uint(campID),
		targetUser,
	); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to add camp participant: %w", err))
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
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create operator: %w", err))
	}

	if !operator.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	targetUser, err := s.repo.GetOrCreateUser(e.Request().Context(), string(userID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create target user: %w", err))
	}

	camp, err := s.repo.GetCampByID(e.Request().Context(), uint(campID))

	if err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get camp: %w", err))
	}

	if err := s.repo.RemoveCampParticipant(
		e.Request().Context(),
		uint(campID),
		targetUser,
	); err != nil {
		switch {
		case errors.Is(err, repository.ErrCampNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")

		case errors.Is(err, repository.ErrParticipantNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "Participant not found")

		default:
			return echo.NewHTTPError(http.StatusInternalServerError).
				SetInternal(fmt.Errorf("failed to remove camp participant: %w", err))
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
