package router

import (
	"errors"
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/converter"
	"github.com/traPtitech/rucQ/model"
)

func (s *Server) GetQuestionGroups(e echo.Context, campID api.CampId) error {
	questionGroups, err := s.repo.GetQuestionGroups(e.Request().Context(), uint(campID))

	if err != nil {
		e.Logger().Errorf("failed to get question groups: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[[]api.QuestionGroupResponse](questionGroups)

	if err != nil {
		e.Logger().Errorf("failed to convert response body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}

func (s *Server) AdminPostQuestionGroup(e echo.Context, campID api.CampId, params api.AdminPostQuestionGroupParams) error {
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

	questionGroup, err := converter.Convert[model.QuestionGroup](req)

	if err != nil {
		e.Logger().Errorf("failed to convert request body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	questionGroup.CampID = uint(campID)

	if err := s.repo.CreateQuestionGroup(&questionGroup); err != nil {
		e.Logger().Errorf("failed to create question group: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[api.QuestionGroupResponse](questionGroup)

	if err != nil {
		e.Logger().Errorf("failed to convert response body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, res)
}

func (s *Server) AdminPutQuestionGroupMetadata(e echo.Context, questionGroupId api.QuestionGroupId, params api.AdminPutQuestionGroupMetadataParams) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPutQuestionGroupMetadataJSONRequestBody

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

	if err := s.repo.UpdateQuestionGroup(e.Request().Context(), uint(questionGroupId), *updateQuestionGroup); err != nil {
		e.Logger().Errorf("failed to update question group: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[api.QuestionGroupResponse](*updateQuestionGroup)

	if err != nil {
		e.Logger().Errorf("failed to convert response body: %v", err)

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
