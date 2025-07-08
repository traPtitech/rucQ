package router

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/converter"
	"github.com/traPtitech/rucQ/model"
)

// AdminGetPayments 支払い情報の一覧を取得（管理者用）
func (s *Server) AdminGetPayments(
	e echo.Context,
	campID api.CampId,
	params api.AdminGetPaymentsParams,
) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	payments, err := s.repo.GetPayments(e.Request().Context(), uint(campID))

	if err != nil {
		e.Logger().Errorf("failed to get payments: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := converter.Convert[[]api.PaymentResponse](payments)

	if err != nil {
		e.Logger().Errorf("failed to convert payments to response: %v", err)

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
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPostPaymentJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	payment, err := converter.Convert[model.Payment](req)

	if err != nil {
		e.Logger().Errorf("failed to convert request to model: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	payment.CampID = uint(campID)

	if err := s.repo.CreatePayment(e.Request().Context(), &payment); err != nil {
		e.Logger().Errorf("failed to create payment: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[api.PaymentResponse](payment)

	if err != nil {
		e.Logger().Errorf("failed to convert model to response: %v", err)

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
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPutPaymentJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	payment, err := converter.Convert[model.Payment](req)

	if err != nil {
		e.Logger().Errorf("failed to convert request to model: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.UpdatePayment(e.Request().Context(), uint(paymentID), &payment); err != nil {
		e.Logger().Errorf("failed to update payment: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// 更新後のPaymentオブジェクトにIDを設定してレスポンスを作成
	payment.ID = uint(paymentID)

	res, err := converter.Convert[api.PaymentResponse](payment)

	if err != nil {
		e.Logger().Errorf("failed to convert model to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}
