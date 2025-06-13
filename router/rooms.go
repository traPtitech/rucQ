package router

import (
	"errors"
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
	"github.com/traP-jp/rucQ/backend/model"
)

func (s *Server) GetRooms(e echo.Context) error {
	rooms, err := s.repo.GetRooms()

	if err != nil {
		e.Logger().Errorf("failed to get rooms: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	var res []RoomResponse

	if err := copier.Copy(&res, &rooms); err != nil {
		e.Logger().Errorf("failed to copy models to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}

func (s *Server) AdminPostRoom(e echo.Context, params AdminPostRoomParams) error {
	operator, err := s.repo.GetOrCreateUser(*params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !operator.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req AdminPostRoomJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, err)
	}

	var roomModel model.Room

	if err := copier.Copy(&roomModel, &req); err != nil {
		e.Logger().Errorf("failed to copy request to model: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	roomModel.Members = make([]model.User, len(req.MemberIds))

	for i := range req.MemberIds {
		member, err := s.repo.GetOrCreateUser(req.MemberIds[i])

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

	var res RoomResponse

	if err := copier.Copy(&res, &roomModel); err != nil {
		e.Logger().Errorf("failed to copy model to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusCreated, res)
}

func (s *Server) AdminPutRoom(e echo.Context, roomId RoomId, params AdminPutRoomParams) error {
	operator, err := s.repo.GetOrCreateUser(*params.XForwardedUser)

	if err != nil {
		e.Logger().Errorf("failed to get or create user: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !operator.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req AdminPutRoomJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return e.JSON(http.StatusBadRequest, err)
	}

	roomModel, err := s.repo.GetRoomByID(uint(roomId))

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Not found")
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
		member, err := s.repo.GetOrCreateUser(req.MemberIds[i])

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

	var res RoomResponse

	if err := copier.Copy(&res, roomModel); err != nil {
		e.Logger().Errorf("failed to copy model to response: %v", err)

		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return e.JSON(http.StatusOK, res)
}
