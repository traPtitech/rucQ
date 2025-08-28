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

func (s *Server) AdminPostRoom(e echo.Context, params api.AdminPostRoomParams) error {
	operator, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user (userId: %s): %w", *params.XForwardedUser, err))
	}

	if !operator.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPostRoomJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	roomModel, err := converter.Convert[model.Room](req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert request to model: %w", err))
	}

	if err := s.repo.CreateRoom(e.Request().Context(), &roomModel); err != nil {
		if errors.Is(err, repository.ErrUserOrRoomGroupNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid user or room group ID")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to create room: %w", err))
	}

	// MemberのisStaffなどを正しく返すために取得
	updatedRoom, err := s.repo.GetRoomByID(roomModel.ID)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get room by ID: %w", err))
	}

	res, err := converter.Convert[api.RoomResponse](updatedRoom)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert model to response (roomId: %d): %w", updatedRoom.ID, err))
	}

	return e.JSON(http.StatusCreated, res)
}

func (s *Server) AdminPutRoom(
	e echo.Context,
	roomID api.RoomId,
	params api.AdminPutRoomParams,
) error {
	operator, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user (userId: %s): %w", *params.XForwardedUser, err))
	}

	if !operator.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPutRoomJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	roomModel, err := converter.Convert[model.Room](req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert request to model: %w", err))
	}

	if err := s.repo.UpdateRoom(e.Request().Context(), uint(roomID), &roomModel); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "User not found")
		}

		if errors.Is(err, repository.ErrRoomGroupNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "Room group not found")
		}

		if errors.Is(err, repository.ErrRoomNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Room not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to update room (roomId: %d): %w", roomID, err))
	}

	// MemberのisStaffなどを正しく返すために取得
	updatedRoom, err := s.repo.GetRoomByID(uint(roomID))

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get room by ID: %w", err))
	}

	res, err := converter.Convert[api.RoomResponse](updatedRoom)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert model to response (roomId: %d): %w", updatedRoom.ID, err))
	}

	return e.JSON(http.StatusOK, res)
}
