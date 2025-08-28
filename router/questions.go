package router

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/converter"
	"github.com/traPtitech/rucQ/model"
)

func (s *Server) AdminDeleteQuestion(
	e echo.Context,
	questionID api.QuestionId,
	params api.AdminDeleteQuestionParams,
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
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if err := s.repo.DeleteQuestionByID(uint(questionID)); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Question not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to delete question",
			slog.String("error", err.Error()),
			slog.Int("questionId", questionID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.NoContent(http.StatusNoContent)
}

func (s *Server) AdminPostQuestion(
	e echo.Context,
	questionGroupID api.QuestionGroupId,
	params api.AdminPostQuestionParams,
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
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPostQuestionJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	question, err := converter.Convert[model.Question](req)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert request to model",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	question.QuestionGroupID = uint(questionGroupID)

	if err := s.repo.CreateQuestion(&question); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to create question",
			slog.String("error", err.Error()),
			slog.Int("questionGroupId", questionGroupID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[api.QuestionResponse](question)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert model to response",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, res)
}

func (s *Server) AdminPutQuestion(
	e echo.Context,
	questionID api.QuestionId,
	params api.AdminPutQuestionParams,
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
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPutQuestionJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	requestQuestion, err := converter.Convert[model.Question](req)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert request to model",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// 既存の質問を取得
	existingQuestion, err := s.repo.GetQuestionByID(uint(questionID))

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Question not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to get question",
			slog.String("error", err.Error()),
			slog.Int("questionId", questionID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// typeは変更できない
	if existingQuestion.Type != requestQuestion.Type {
		return echo.NewHTTPError(http.StatusBadRequest, "question type cannot be changed")
	}

	requestQuestion.ID = uint(questionID)

	if err := s.repo.UpdateQuestion(e.Request().Context(), uint(questionID), &requestQuestion); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Question not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to update question",
			slog.String("error", err.Error()),
			slog.Int("questionId", questionID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[api.QuestionResponse](requestQuestion)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert model to response",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, &res)
}
