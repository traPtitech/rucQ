package router

import (
	"errors"
	"fmt"
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
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user (userId: %s): %w", *params.XForwardedUser, err))
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if err := s.repo.DeleteQuestionByID(uint(questionID)); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Question not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to delete question (questionId: %d): %w", questionID, err))
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
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user (userId: %s): %w", *params.XForwardedUser, err))
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
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert request to model: %w", err))
	}

	question.QuestionGroupID = uint(questionGroupID)

	if err := s.repo.CreateQuestion(&question); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to create question (questionGroupId: %d): %w", questionGroupID, err))
	}

	res, err := converter.Convert[api.QuestionResponse](question)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert model to response: %w", err))
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
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user (userId: %s): %w", *params.XForwardedUser, err))
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
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert request to model: %w", err))
	}

	// 既存の質問を取得
	existingQuestion, err := s.repo.GetQuestionByID(uint(questionID))

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Question not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get question (questionId: %d): %w", questionID, err))
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

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to update question (questionId: %d): %w", questionID, err))
	}

	res, err := converter.Convert[api.QuestionResponse](requestQuestion)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert model to response: %w", err))
	}

	return e.JSON(http.StatusOK, &res)
}
