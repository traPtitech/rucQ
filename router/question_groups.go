package router

import (
	"errors"
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
)

func (s *Server) GetQuestionGroups(e echo.Context, _ api.CampId) error {
	questionGroups, err := s.repo.GetQuestionGroups()

	if err != nil {
		e.Logger().Errorf("failed to get question groups: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var res []api.QuestionGroupResponse

	if err := copier.Copy(&res, &questionGroups); err != nil {
		e.Logger().Errorf("failed to copy response body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}

func (s *Server) GetQuestionGroup(e echo.Context, questionGroupID api.QuestionGroupId) error {
	questionGroup, err := s.repo.GetQuestionGroup(uint(questionGroupID))

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Not found")
		}

		e.Logger().Errorf("failed to get question group: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var res api.QuestionGroupResponse

	if err := copier.Copy(&res, questionGroup); err != nil {
		e.Logger().Errorf("failed to copy response body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}

func (s *Server) AdminPostQuestionGroup(e echo.Context, _ api.CampId, params api.AdminPostQuestionGroupParams) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPostQuestionGroupJSONRequestBody

	if err := e.Bind(&req); err != nil {
		e.Logger().Errorf("failed to bind request body: %v", err)

		return echo.NewHTTPError(http.StatusBadRequest, "Bad request")
	}

	var questionGroup model.QuestionGroup

	if err := copier.Copy(&questionGroup, &req); err != nil {
		e.Logger().Errorf("failed to copy request body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.CreateQuestionGroup(&questionGroup); err != nil {
		e.Logger().Errorf("failed to create question group: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var res api.QuestionGroupResponse

	if err := copier.Copy(&res, &questionGroup); err != nil {
		e.Logger().Errorf("failed to copy response body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, res)
}

func (s *Server) AdminPutQuestionGroup(e echo.Context, questionGroupId api.QuestionGroupId, params api.AdminPutQuestionGroupParams) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPutQuestionGroupJSONRequestBody

	if err := e.Bind(&req); err != nil {
		e.Logger().Errorf("failed to bind request body: %v", err)

		return echo.NewHTTPError(http.StatusBadRequest, "Bad request")
	}

	updateQuestionGroup, err := s.repo.GetQuestionGroup(uint(questionGroupId))

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Not found")
		}
		e.Logger().Errorf("failed to get question group: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := copier.Copy(&updateQuestionGroup, req); err != nil {
		e.Logger().Errorf("failed to copy request body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.UpdateQuestionGroup(uint(questionGroupId), updateQuestionGroup); err != nil {
		e.Logger().Errorf("failed to update question group: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var res api.QuestionGroupResponse

	if err := copier.Copy(&res, updateQuestionGroup); err != nil {
		e.Logger().Errorf("failed to copy response body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}

func (s *Server) AdminDeleteQuestionGroup(e echo.Context, questionGroupId api.QuestionGroupId, params api.AdminDeleteQuestionGroupParams) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if err := s.repo.DeleteQuestionGroup(uint(questionGroupId)); err != nil {
		e.Logger().Errorf("failed to delete question group: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.NoContent(http.StatusNoContent)
}
