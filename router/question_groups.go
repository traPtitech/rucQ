package router

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/converter"
	"github.com/traPtitech/rucQ/model"
)

func (s *Server) GetQuestionGroups(e echo.Context, campID api.CampId) error {
	questionGroups, err := s.repo.GetQuestionGroups(e.Request().Context(), uint(campID))

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get question groups",
			slog.String("error", err.Error()),
			slog.Int("campId", int(campID)),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[[]api.QuestionGroupResponse](questionGroups)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert response body",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
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

	var req api.AdminPostQuestionGroupJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	questionGroup, err := converter.Convert[model.QuestionGroup](req)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert request body",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	questionGroup.CampID = uint(campID)

	if err := s.repo.CreateQuestionGroup(&questionGroup); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to create question group",
			slog.String("error", err.Error()),
			slog.Int("campId", campID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[api.QuestionGroupResponse](questionGroup)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert response body",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
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

	var req api.AdminPutQuestionGroupMetadataJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	questionGroup, err := converter.Convert[model.QuestionGroup](req)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert request body",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.UpdateQuestionGroup(e.Request().Context(), uint(questionGroupID), questionGroup); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to update question group",
			slog.String("error", err.Error()),
			slog.Int("questionGroupId", questionGroupID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	updatedQuestionGroup, err := s.repo.GetQuestionGroup(
		e.Request().Context(),
		uint(questionGroupID),
	)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get updated question group",
			slog.String("error", err.Error()),
			slog.Int("questionGroupId", questionGroupID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[api.QuestionGroupResponse](updatedQuestionGroup)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert response body",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
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

	if err := s.repo.DeleteQuestionGroup(uint(questionGroupID)); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to delete question group",
			slog.String("error", err.Error()),
			slog.Int("questionGroupId", questionGroupID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.NoContent(http.StatusNoContent)
}
