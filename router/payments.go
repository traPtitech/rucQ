package router

import (
	"errors"
	"fmt"
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
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user: %w", err))
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	payments, err := s.repo.GetPayments(e.Request().Context(), uint(campID))

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get payments: %w", err))
	}

	response, err := converter.Convert[[]api.PaymentResponse](payments)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert payments to response: %w", err))
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
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user: %w", err))
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
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert request to model: %w", err))
	}

	payment.CampID = uint(campID)

	if err := s.repo.CreatePayment(e.Request().Context(), &payment); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to create payment: %w", err))
	}

	_ = s.activityService.RecordPaymentCreated(e.Request().Context(), payment)

	res, err := converter.Convert[api.PaymentResponse](payment)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert model to response: %w", err))
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
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user: %w", err))
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
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert request to model: %w", err))
	}

	if err := s.repo.UpdatePayment(e.Request().Context(), uint(paymentID), &payment); err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Payment not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to update payment: %w", err))
	}

	updatedPayment, err := s.repo.GetPaymentByID(e.Request().Context(), uint(paymentID))
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Payment not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get payment: %w", err))
	}

	// Amount/AmountPaid が変更された場合にアクティビティを記録
	if payment.Amount != 0 {
		_ = s.activityService.RecordPaymentAmountChanged(e.Request().Context(), *updatedPayment)
	}

	if payment.AmountPaid != 0 {
		_ = s.activityService.RecordPaymentPaidChanged(e.Request().Context(), *updatedPayment)
	}

	res, err := converter.Convert[api.PaymentResponse](updatedPayment)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert model to response: %w", err))
	}

	return e.JSON(http.StatusOK, res)
}
