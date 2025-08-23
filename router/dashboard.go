package router

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/converter"
	"github.com/traPtitech/rucQ/model"
	"github.com/traPtitech/rucQ/repository"
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

	res := api.DashboardResponse{
		Id: *params.XForwardedUser,
	}

	payment, err := s.repo.GetPaymentByUserID(
		e.Request().Context(),
		uint(campID),
		*params.XForwardedUser,
	)

	if err != nil && !errors.Is(err, repository.ErrPaymentNotFound) {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get payment",
			slog.String("userId", *params.XForwardedUser),
			slog.Int("campId", campID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if payment != nil {
		apiPayment, err := converter.Convert[api.PaymentResponse](payment)

		if err != nil {
			slog.ErrorContext(
				e.Request().Context(),
				"failed to convert payment",
				slog.String("userId", *params.XForwardedUser),
				slog.Int("campId", campID),
			)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		res.Payment = &apiPayment
	}

	room, err := s.repo.GetRoomByUserID(e.Request().Context(), uint(campID), *params.XForwardedUser)

	if err != nil && !errors.Is(err, repository.ErrRoomNotFound) {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get room",
			slog.String("userId", *params.XForwardedUser),
			slog.Int("campId", campID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if room != nil {
		apiRoom, err := converter.Convert[api.RoomResponse](room)

		if err != nil {
			slog.ErrorContext(
				e.Request().Context(),
				"failed to convert room",
				slog.String("userId", *params.XForwardedUser),
				slog.Int("campId", campID),
			)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		res.Room = &apiRoom
	}

	return e.JSON(http.StatusOK, &res)
}
