package router

import (
	"errors"
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
	"github.com/traP-jp/rucQ/backend/model"
)

func (s *Server) GetQuestions(e echo.Context) error {
	questions, err := s.repo.GetQuestions()

	if err != nil {
		e.Logger().Errorf("failed to get questions: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var questionsResponse []QuestionResponse

	if err := copier.Copy(&questionsResponse, &questions); err != nil {
		e.Logger().Errorf("failed to copy questions: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, questionsResponse)
}

func (s *Server) AdminPostQuestion(e echo.Context, params AdminPostQuestionParams) error {
	user, err := s.repo.GetOrCreateUser(*params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req AdminPostQuestionJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, err)
	}

	var questionModel model.Question

	if err := copier.Copy(&questionModel, &req); err != nil {
		e.Logger().Errorf("failed to copy request to model: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.CreateQuestion(&questionModel); err != nil {
		e.Logger().Errorf("failed to create question: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var questionResponse QuestionResponse

	if err := copier.Copy(&questionResponse, &questionModel); err != nil {
		e.Logger().Errorf("failed to copy model to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, &questionResponse)
}

func (s *Server) AdminDeleteQuestion(e echo.Context, questionId QuestionId, params AdminDeleteQuestionParams) error {
	user, err := s.repo.GetOrCreateUser(*params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if err := s.repo.DeleteQuestionByID(uint(questionId)); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Not found")
		}

		e.Logger().Errorf("failed to delete question: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.NoContent(http.StatusNoContent)
}

func (s *Server) GetQuestion(e echo.Context, questionID QuestionId) error {
	question, err := s.repo.GetQuestionByID(uint(questionID))

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Not found")
		}

		e.Logger().Errorf("failed to get question: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var questionResponse QuestionResponse

	if err := copier.Copy(&questionResponse, &question); err != nil {
		e.Logger().Errorf("failed to copy question: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, &questionResponse)
}

func (s *Server) AdminPutQuestion(e echo.Context, questionId QuestionId, params AdminPutQuestionParams) error {
	user, err := s.repo.GetOrCreateUser(*params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req AdminPutQuestionJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, err)
	}

	var questionModel model.Question

	if err := copier.Copy(&questionModel, &req); err != nil {
		e.Logger().Errorf("failed to copy request to model: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.UpdateQuestion(uint(questionId), &questionModel); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Not found")
		}

		e.Logger().Errorf("failed to update question: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var questionResponse QuestionResponse

	if err := copier.Copy(&questionResponse, &questionModel); err != nil {
		e.Logger().Errorf("failed to copy model to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// TODO: questionResponse.Id is not available in new Question structure
	// questionResponse.Id = questionId

	return e.JSON(http.StatusOK, &questionResponse)
}
