package router

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/converter"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
)

func (s *Server) GetRollCalls(e echo.Context, campID api.CampId) error {
	rollCalls, err := s.repo.GetRollCalls(e.Request().Context(), uint(campID))

	if err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get roll calls: %w", err))
	}

	res, err := converter.Convert[[]api.RollCallResponse](rollCalls)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert roll calls: %w", err))
	}

	return e.JSON(http.StatusOK, res)
}

func (s *Server) AdminPostRollCall(
	e echo.Context,
	campID api.CampId,
	params api.AdminPostRollCallParams,
) error {
	if params.XForwardedUser == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "X-Forwarded-User header is required")
	}

	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user: %w", err))
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPostRollCallJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	rollCall, err := converter.Convert[model.RollCall](req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert request body: %w", err))
	}

	rollCall.CampID = uint(campID)

	if err := s.repo.CreateRollCall(e.Request().Context(), &rollCall); err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		if errors.Is(err, repository.ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "One or more subject users not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to create roll call: %w", err))
	}

	res, err := converter.Convert[api.RollCallResponse](rollCall)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert roll call: %w", err))
	}

	return e.JSON(http.StatusCreated, res)
}

func (s *Server) GetRollCallReactions(e echo.Context, rollCallID api.RollCallId) error {
	reactions, err := s.repo.GetRollCallReactions(e.Request().Context(), uint(rollCallID))

	if err != nil {
		if errors.Is(err, repository.ErrRollCallNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"roll call not found when getting reactions",
				slog.Int("rollCallId", rollCallID),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Roll call not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to get roll call reactions",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[[]api.RollCallReactionResponse](reactions)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert roll call reactions",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}

func (s *Server) PostRollCallReaction(
	e echo.Context,
	rollCallID api.RollCallId,
	params api.PostRollCallReactionParams,
) error {
	if params.XForwardedUser == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "X-Forwarded-User header is required")
	}

	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get or create user",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var req api.PostRollCallReactionJSONRequestBody

	if err := e.Bind(&req); err != nil {
		slog.WarnContext(
			e.Request().Context(),
			"failed to bind request body",
			slog.String("error", err.Error()),
		)

		return err
	}

	reaction := model.RollCallReaction{
		Content:    req.Content,
		UserID:     user.ID,
		RollCallID: uint(rollCallID),
	}

	if err := s.repo.CreateRollCallReaction(e.Request().Context(), &reaction); err != nil {
		if errors.Is(err, repository.ErrRollCallNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"roll call not found",
				slog.Int("rollCallId", rollCallID),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Roll call not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to create roll call reaction",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[api.RollCallReactionResponse](reaction)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert roll call reaction",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, res)
}

func (s *Server) PutReaction(e echo.Context, reactionID api.ReactionId, params api.PutReactionParams) error {
	if params.XForwardedUser == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "X-Forwarded-User header is required")
	}

	// リアクションの存在確認と所有者チェック
	existingReaction, err := s.repo.GetRollCallReactionByID(e.Request().Context(), uint(reactionID))

	if err != nil {
		if errors.Is(err, repository.ErrRollCallReactionNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"roll call reaction not found",
				slog.Int("reactionId", reactionID),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Reaction not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to get roll call reaction",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// ユーザーの所有者チェック
	if existingReaction.UserID != *params.XForwardedUser {
		return echo.NewHTTPError(http.StatusForbidden, "You can only edit your own reactions")
	}

	var req api.PutReactionJSONRequestBody

	if err := e.Bind(&req); err != nil {
		slog.WarnContext(
			e.Request().Context(),
			"failed to bind request body",
			slog.String("error", err.Error()),
		)

		return err
	}

	updateData := model.RollCallReaction{
		Content: req.Content,
	}

	if err := s.repo.UpdateRollCallReaction(e.Request().Context(), uint(reactionID), &updateData); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to update roll call reaction",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// 更新後のデータを取得
	updatedReaction, err := s.repo.GetRollCallReactionByID(e.Request().Context(), uint(reactionID))

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get updated roll call reaction",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[api.RollCallReactionResponse](*updatedReaction)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert roll call reaction",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}

func (s *Server) DeleteReaction(
	e echo.Context,
	reactionID api.ReactionId,
	params api.DeleteReactionParams,
) error {
	if params.XForwardedUser == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "X-Forwarded-User header is required")
	}

	// リアクションの存在確認と所有者チェック
	existingReaction, err := s.repo.GetRollCallReactionByID(e.Request().Context(), uint(reactionID))

	if err != nil {
		if errors.Is(err, repository.ErrRollCallReactionNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"roll call reaction not found",
				slog.Int("reactionId", reactionID),
			)

			return echo.NewHTTPError(http.StatusNotFound, "Reaction not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to get roll call reaction",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// ユーザーの所有者チェック
	if existingReaction.UserID != *params.XForwardedUser {
		return echo.NewHTTPError(http.StatusForbidden, "You can only delete your own reactions")
	}

	if err := s.repo.DeleteRollCallReaction(e.Request().Context(), uint(reactionID)); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to delete roll call reaction",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.NoContent(http.StatusNoContent)
}
