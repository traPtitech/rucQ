package router

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/converter"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
)

func (s *Server) GetQuestionGroups(e echo.Context, campID api.CampId) error {
	questionGroups, err := s.repo.GetQuestionGroups(e.Request().Context(), uint(campID))

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get question groups: %w", err))
	}

	res, err := converter.Convert[[]api.QuestionGroupResponse](questionGroups)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert response body: %w", err))
	}

	return e.JSON(http.StatusOK, res)
}

func (s *Server) AdminPostQuestionGroup(
	e echo.Context,
	campID api.CampId,
	params api.AdminPostQuestionGroupParams,
) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user: %w", err))
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPostQuestionGroupJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	questionGroup, err := converter.Convert[model.QuestionGroup](req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert request body: %w", err))
	}

	questionGroup.CampID = uint(campID)

	ctx := e.Request().Context()

	if err := s.repo.Transaction(ctx, func(tx repository.Repository) error {
		if err := tx.CreateQuestionGroup(&questionGroup); err != nil {
			return fmt.Errorf("failed to create question group: %w", err)
		}

		return s.activityService.RecordQuestionCreated(ctx, tx, questionGroup)
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(err)
	}

	res, err := converter.Convert[api.QuestionGroupResponse](questionGroup)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert response body: %w", err))
	}

	return e.JSON(http.StatusCreated, res)
}

func (s *Server) AdminPutQuestionGroupMetadata(
	e echo.Context,
	questionGroupID api.QuestionGroupId,
	params api.AdminPutQuestionGroupMetadataParams,
) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user: %w", err))
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPutQuestionGroupMetadataJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	questionGroup, err := converter.Convert[model.QuestionGroup](req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert request body: %w", err))
	}

	if err := s.repo.UpdateQuestionGroup(
		e.Request().Context(),
		uint(questionGroupID),
		questionGroup,
	); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to update question group: %w", err))
	}

	updatedQuestionGroup, err := s.repo.GetQuestionGroup(
		e.Request().Context(),
		uint(questionGroupID),
	)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get updated question group: %w", err))
	}

	res, err := converter.Convert[api.QuestionGroupResponse](updatedQuestionGroup)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert response body: %w", err))
	}

	return e.JSON(http.StatusOK, res)
}

func (s *Server) AdminDeleteQuestionGroup(
	e echo.Context,
	questionGroupID api.QuestionGroupId,
	params api.AdminDeleteQuestionGroupParams,
) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user: %w", err))
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if err := s.repo.DeleteQuestionGroup(uint(questionGroupID)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to delete question group: %w", err))
	}

	return e.NoContent(http.StatusNoContent)
}
