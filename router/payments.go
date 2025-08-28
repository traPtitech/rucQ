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

// AdminGetPayments 支払い情報の一覧を取得（管理者用）
func (s *Server) AdminGetPayments(
	e echo.Context,
	campID api.CampId,
	params api.AdminGetPaymentsParams,
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

	payments, err := s.repo.GetPayments(e.Request().Context(), uint(campID))

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to get payments",
			slog.String("error", err.Error()),
			slog.Int("campId", int(campID)),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := converter.Convert[[]api.PaymentResponse](payments)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert payments to response",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, response)
}

// AdminPostPayment 支払い情報を作成（管理者用）
func (s *Server) AdminPostPayment(
	e echo.Context,
	campID api.CampId,
	params api.AdminPostPaymentParams,
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

	var req api.AdminPostPaymentJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	payment, err := converter.Convert[model.Payment](req)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert request to model",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	payment.CampID = uint(campID)

	if err := s.repo.CreatePayment(e.Request().Context(), &payment); err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to create payment",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[api.PaymentResponse](payment)

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

// AdminPutPayment 支払い情報を更新（管理者用）
func (s *Server) AdminPutPayment(
	e echo.Context,
	paymentID api.PaymentId,
	params api.AdminPutPaymentParams,
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

	var req api.AdminPutPaymentJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	payment, err := converter.Convert[model.Payment](req)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert request to model",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.UpdatePayment(e.Request().Context(), uint(paymentID), &payment); err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Payment not found")
		}

		slog.ErrorContext(
			e.Request().Context(),
			"failed to update payment",
			slog.String("error", err.Error()),
			slog.Int("paymentId", paymentID),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// 更新後のPaymentオブジェクトにIDを設定してレスポンスを作成
	payment.ID = uint(paymentID)

	res, err := converter.Convert[api.PaymentResponse](payment)

	if err != nil {
		slog.ErrorContext(
			e.Request().Context(),
			"failed to convert model to response",
			slog.String("error", err.Error()),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}
