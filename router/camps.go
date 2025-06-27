package router

import (
	"errors"
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"

	"github.com/traP-jp/rucQ/backend/api"
	"github.com/traP-jp/rucQ/backend/converter"
	"github.com/traP-jp/rucQ/backend/model"
)

func (s *Server) GetCamps(e echo.Context) error {
	camps, err := s.repo.GetCamps()

	if err != nil {
		e.Logger().Errorf("failed to get camps: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := converter.Convert[[]api.CampResponse](camps)

	if err != nil {
		e.Logger().Errorf("failed to convert camps: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, response)
}

// AdminPostCamp 新規キャンプ追加
// (POST /api/admin/camps)
func (s *Server) AdminPostCamp(e echo.Context, params api.AdminPostCampParams) error {
	var req api.AdminPostCampJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, err)
	}

	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	campModel, err := converter.Convert[model.Camp](req)

	if err != nil {
		e.Logger().Errorf("failed to convert request to model: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.CreateCamp(&campModel); err != nil {
		if errors.Is(err, model.ErrAlreadyExists) {
			return echo.NewHTTPError(http.StatusConflict, "Camp already exists")
		}

		e.Logger().Errorf("failed to create camp: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := converter.Convert[api.CampResponse](campModel)

	if err != nil {
		e.Logger().Errorf("failed to convert camp to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, &response)
}

func (s *Server) GetCamp(e echo.Context, campID api.CampId) error {
	camp, err := s.repo.GetCampByID(uint(campID))

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Not found")
		}

		e.Logger().Errorf("failed to get camp: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := converter.Convert[api.CampResponse](camp)

	if err != nil {
		e.Logger().Errorf("failed to convert camp to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, &response)
}

// AdminPutCamp キャンプ情報編集
// (PUT /api/admin/camps/{campId})
func (s *Server) AdminPutCamp(e echo.Context, campID api.CampId, params api.AdminPutCampParams) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPutCampJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, err)
	}

	newCamp, err := converter.Convert[model.Camp](req)

	if err != nil {
		e.Logger().Errorf("failed to get camp: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	newCamp.ID = uint(campID)

	// TODO: Not foundエラーのハンドリングを追加
	if err := s.repo.UpdateCamp(uint(campID), &newCamp); err != nil {
		e.Logger().Errorf("failed to update camp: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	response, err := converter.Convert[api.CampResponse](newCamp)

	if err != nil {
		e.Logger().Errorf("failed to convert camp to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, &response)
}

// AdminDeleteCamp キャンプ削除
// (DELETE /api/admin/camps/{campId})
func (s *Server) AdminDeleteCamp(e echo.Context, campID api.CampId, params api.AdminDeleteCampParams) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	// TODO: Not found エラーのハンドリングを追加
	if err := s.repo.DeleteCamp(e.Request().Context(), uint(campID)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.NoContent(http.StatusNoContent)
}

// PostCampRegister 合宿に登録
func (s *Server) PostCampRegister(e echo.Context, campID api.CampId, params api.PostCampRegisterParams) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// TODO: Forbidden、 Not found エラーのハンドリングを追加
	if err := s.repo.AddCampParticipant(e.Request().Context(), uint(campID), user); err != nil {
		e.Logger().Errorf("failed to add camp participant: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.NoContent(http.StatusNoContent)
}

// DeleteCampRegister 合宿登録を取り消し
func (s *Server) DeleteCampRegister(e echo.Context, campID api.CampId, params api.DeleteCampRegisterParams) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// TODO: Forbidden、 Not found エラーのハンドリングを追加
	if err := s.repo.RemoveCampParticipant(e.Request().Context(), uint(campID), user); err != nil {
		e.Logger().Errorf("failed to remove camp participant: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.NoContent(http.StatusNoContent)
}

// GetCampParticipants 合宿の参加者一覧を取得
func (s *Server) GetCampParticipants(e echo.Context, campID api.CampId) error {
	participants, err := s.repo.GetCampParticipants(e.Request().Context(), uint(campID))

	// TODO: Not found エラーのハンドリングを追加（APIスキーマにも追加する）
	if err != nil {
		e.Logger().Errorf("failed to get camp participants: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response []api.UserResponse

	if err := copier.Copy(&response, &participants); err != nil {
		e.Logger().Errorf("failed to copy participants: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, response)
}
