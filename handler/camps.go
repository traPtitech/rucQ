package handler

import (
	"errors"
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
	"github.com/traP-jp/rucQ/backend/model"
)

func (s *Server) GetCamps(e echo.Context) error {
	camps, err := s.repo.GetCamps()

	if err != nil {
		e.Logger().Errorf("failed to get camps: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response []Camp

	if err := copier.Copy(&response, &camps); err != nil {
		e.Logger().Errorf("failed to copy camps: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, response)
}

// AdminPostCamp 新規キャンプ追加
// (POST /api/admin/camps)
func (s *Server) AdminPostCamp(e echo.Context, params AdminPostCampParams) error {
	var req AdminPostCampJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, err)
	}

	user, err := s.repo.GetOrCreateUser(*params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var campModel model.Camp

	if err := copier.Copy(&campModel, &req); err != nil {
		e.Logger().Errorf("failed to copy request to model: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.CreateCamp(&campModel); err != nil {
		if errors.Is(err, model.ErrAlreadyExists) {
			return echo.NewHTTPError(http.StatusConflict, "Camp already exists")
		}

		e.Logger().Errorf("failed to create camp: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response Camp

	if err := copier.Copy(&response, &campModel); err != nil {
		e.Logger().Errorf("failed to copy camp: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, response)
}

func (s *Server) GetDefaultCamp(e echo.Context) error {
	camp, err := s.repo.GetDefaultCamp()

	if err != nil {
		e.Logger().Errorf("failed to get default camp: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response Camp

	if err := copier.Copy(&response, camp); err != nil {
		e.Logger().Errorf("failed to copy camp: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, response)
}

func (s *Server) GetCamp(e echo.Context, campID CampId) error {
	camp, err := s.repo.GetCampByID(uint(campID))

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Not found")
		}

		e.Logger().Errorf("failed to get camp: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response Camp

	if err := copier.Copy(&response, camp); err != nil {
		e.Logger().Errorf("failed to copy camp: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, response)
}

// AdminPutCamp キャンプ情報編集
// (PUT /api/admin/camps/{campId})
func (s *Server) AdminPutCamp(e echo.Context, campId CampId, params AdminPutCampParams) error {
	user, err := s.repo.GetOrCreateUser(*params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req AdminPutCampJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, err)
	}

	camp, err := s.repo.GetCampByID(uint(campId))

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Not found")
		}

		e.Logger().Errorf("failed to get camp: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := copier.Copy(camp, &req); err != nil {
		e.Logger().Errorf("failed to copy request to model: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.UpdateCamp(uint(campId), camp); err != nil {
		e.Logger().Errorf("failed to update camp: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var response Camp

	if err := copier.Copy(&response, camp); err != nil {
		e.Logger().Errorf("failed to copy camp: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, response)
}

// AdminDeleteCamp キャンプ削除
// (DELETE /api/admin/camps/{campId})
func (s *Server) AdminDeleteCamp(e echo.Context, campId CampId, params AdminDeleteCampParams) error {
	user, err := s.repo.GetOrCreateUser(*params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	// TODO: キャンプ削除機能は実装されていません
	return echo.NewHTTPError(http.StatusNotImplemented, "Camp deletion not implemented")
}
