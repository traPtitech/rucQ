package router

import (
	"errors"
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/api"
	"github.com/traPtitech/rucQ/model"
)

func (s *Server) AdminPostRoom(e echo.Context, params api.AdminPostRoomParams) error {
	operator, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !operator.IsStaff {
		e.Logger().Warnf("user %s is not a staff member", *params.XForwardedUser)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPostRoomJSONRequestBody

	if err := e.Bind(&req); err != nil {
		e.Logger().Warnf("failed to bind request body: %v", err)

		return err
	}

	var roomModel model.Room

	if err := copier.Copy(&roomModel, &req); err != nil {
		e.Logger().Errorf("failed to copy request to model: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	roomModel.Members = make([]model.User, len(req.MemberIds))

	for i := range req.MemberIds {
		member, err := s.repo.GetOrCreateUser(e.Request().Context(), req.MemberIds[i])

		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "User not found")
			}

			e.Logger().Errorf("failed to get user: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		roomModel.Members[i] = *member
	}

	if err := s.repo.CreateRoom(&roomModel); err != nil {
		if errors.Is(err, model.ErrAlreadyExists) {
			return echo.NewHTTPError(http.StatusConflict, "Room already exists")
		}

		e.Logger().Errorf("failed to create room: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var res api.RoomResponse

	if err := copier.Copy(&res, &roomModel); err != nil {
		e.Logger().Errorf("failed to copy model to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, res)
}

func (s *Server) AdminPutRoom(
	e echo.Context,
	roomId api.RoomId,
	params api.AdminPutRoomParams,
) error {
	operator, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !operator.IsStaff {
		e.Logger().Warnf("user %s is not a staff member", *params.XForwardedUser)

		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPutRoomJSONRequestBody

	if err := e.Bind(&req); err != nil {
		e.Logger().Warnf("failed to bind request body: %v", err)

		return err
	}

	roomModel, err := s.repo.GetRoomByID(uint(roomId))

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			e.Logger().Warnf("room with ID %d not found", roomId)

			return echo.NewHTTPError(http.StatusNotFound, "Room not found")
		}

		e.Logger().Errorf("failed to get room: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if err := copier.Copy(roomModel, &req); err != nil {
		e.Logger().Errorf("failed to copy request to model: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	roomModel.Members = make([]model.User, len(req.MemberIds))

	for i := range req.MemberIds {
		member, err := s.repo.GetOrCreateUser(e.Request().Context(), req.MemberIds[i])

		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "User not found")
			}

			e.Logger().Errorf("failed to get user: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		roomModel.Members[i] = *member
	}

	if err := s.repo.UpdateRoom(roomModel); err != nil {
		e.Logger().Errorf("failed to update room: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var res api.RoomResponse

	if err := copier.Copy(&res, roomModel); err != nil {
		e.Logger().Errorf("failed to copy model to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}
