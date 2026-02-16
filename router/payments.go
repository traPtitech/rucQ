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

	ctx := e.Request().Context()

	if err := s.repo.Transaction(ctx, func(tx repository.Repository) error {
		if err := tx.CreatePayment(ctx, &payment); err != nil {
			return fmt.Errorf("failed to create payment: %w", err)
		}

		return s.activityService.RecordPaymentCreated(ctx, tx, payment)
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(err)
	}

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

	beforePayment, err := s.repo.GetPaymentByID(e.Request().Context(), uint(paymentID))

	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return echo.ErrNotFound
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get payment: %w", err))
	}

	ctx := e.Request().Context()

	var updatedPayment *model.Payment

	if err := s.repo.Transaction(ctx, func(tx repository.Repository) error {
		if err := tx.UpdatePayment(ctx, uint(paymentID), &payment); err != nil {
			// もしErrPaymentNotFoundがここで返ってきた場合、それは異常なので500エラーとする
			return fmt.Errorf("failed to update payment: %w", err)
		}

		var err error
		updatedPayment, err = tx.GetPaymentByID(ctx, uint(paymentID))
		if err != nil {
			return fmt.Errorf("failed to get payment: %w", err)
		}

		// Amount/AmountPaid が変更された場合にアクティビティを記録
		if updatedPayment.Amount != beforePayment.Amount {
			if err := s.activityService.RecordPaymentAmountChanged(
				ctx,
				tx,
				*updatedPayment,
			); err != nil {
				return err
			}
		}

		if updatedPayment.AmountPaid != beforePayment.AmountPaid {
			if err := s.activityService.RecordPaymentPaidChanged(
				ctx,
				tx,
				*updatedPayment,
			); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(err)
	}

	res, err := converter.Convert[api.PaymentResponse](updatedPayment)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert model to response: %w", err))
	}

	return e.JSON(http.StatusOK, res)
}
