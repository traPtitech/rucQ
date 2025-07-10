package router

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/converter"
	"github.com/traPtitech/rucQ/model"
)

func (s *Server) AdminPostRoomGroup(
	e echo.Context,
	campID api.CampId,
	params api.AdminPostRoomGroupParams,
) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPostRoomGroupJSONRequestBody

	if err := e.Bind(&req); err != nil {
		e.Logger().Errorf("failed to bind request body: %v", err)

		return echo.NewHTTPError(http.StatusBadRequest, "Bad request")
	}

	roomGroup, err := converter.Convert[model.RoomGroup](req)

	if err != nil {
		e.Logger().Errorf("failed to convert request body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	roomGroup.CampID = uint(campID)

	if err := s.repo.CreateRoomGroup(e.Request().Context(), &roomGroup); err != nil {
		e.Logger().Errorf("failed to create room group: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[api.RoomGroupResponse](roomGroup)

	if err != nil {
		e.Logger().Errorf("failed to convert response body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, res)
}

func (s *Server) AdminPutRoomGroup(
	e echo.Context,
	roomGroupID api.RoomGroupId,
	params api.AdminPutRoomGroupParams,
) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)
	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPutRoomGroupJSONRequestBody

	if err := e.Bind(&req); err != nil {
		e.Logger().Errorf("failed to bind request body: %v", err)

		return echo.NewHTTPError(http.StatusBadRequest, "Bad request")
	}

	roomGroup, err := converter.Convert[model.RoomGroup](req)
	if err != nil {
		e.Logger().Errorf("failed to convert request body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := s.repo.UpdateRoomGroup(e.Request().Context(), uint(roomGroupID), &roomGroup); err != nil {
		e.Logger().Errorf("failed to update room group: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	// 更新後のデータを取得（存在確認も兼ねる）
	updatedRoomGroup, err := s.repo.GetRoomGroupByID(e.Request().Context(), uint(roomGroupID))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Room group not found")
		}
		e.Logger().Errorf("failed to get updated room group: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	res, err := converter.Convert[api.RoomGroupResponse](*updatedRoomGroup)
	if err != nil {
		e.Logger().Errorf("failed to convert response body: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}
