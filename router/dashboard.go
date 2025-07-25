package router

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
)

func (s *Server) GetDashboard(
	e echo.Context,
	campID api.CampId,
	params api.GetDashboardParams,
) error {
	isParticipant, err := s.repo.IsCampParticipant(
		e.Request().Context(),
		uint(campID),
		*params.XForwardedUser,
	)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			slog.WarnContext(
				e.Request().Context(),
				"camp not found",
				slog.Int("campId", int(campID)),
			)

			return echo.NewHTTPError(
				http.StatusNotFound,
				"Camp not found",
			)
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to check camp participation",
			slog.String("error", err.Error()),
			slog.Int("campId", int(campID)),
		)

		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Internal server error",
		)
	}

	if !isParticipant {
		slog.WarnContext(
			e.Request().Context(),
			"user is not a participant of camp",
			slog.String("userId", *params.XForwardedUser),
			slog.Int("campId", int(campID)),
		)

		return echo.NewHTTPError(http.StatusNotFound, "User is not a participant of this camp")
	}

	// TODO: ユーザーのPaymentとRoomを取得してレスポンスに含める
	res := api.DashboardResponse{
		Id: *params.XForwardedUser,
	}

	return e.JSON(http.StatusOK, &res)
}
