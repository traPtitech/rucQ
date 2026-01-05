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

func (s *Server) GetRoomGroups(e echo.Context, campID api.CampId) error {
	roomGroups, err := s.repo.GetRoomGroups(e.Request().Context(), uint(campID))

	if err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get room groups (campId: %d): %w", campID, err))
	}

	res, err := converter.Convert[[]api.RoomGroupResponse](roomGroups)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert response body: %w", err))
	}

	return e.JSON(http.StatusOK, res)
}

func (s *Server) AdminPostRoomGroup(
	e echo.Context,
	campID api.CampId,
	params api.AdminPostRoomGroupParams,
) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user (userId: %s): %w", *params.XForwardedUser, err))
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPostRoomGroupJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	roomGroup, err := converter.Convert[model.RoomGroup](req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert request body: %w", err))
	}

	roomGroup.CampID = uint(campID)

	if err := s.repo.CreateRoomGroup(e.Request().Context(), &roomGroup); err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Camp not found")
		}

		if errors.Is(err, repository.ErrUserAlreadyAssigned) {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				"Some users are already assigned to another room in this camp",
			)
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to create room group (campID: %d): %w", campID, err))
	}

	res, err := converter.Convert[api.RoomGroupResponse](roomGroup)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert response body: %w", err))
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
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user (userId: %s): %w", *params.XForwardedUser, err))
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req api.AdminPutRoomGroupJSONRequestBody

	if err := e.Bind(&req); err != nil {
		return err
	}

	roomGroup, err := converter.Convert[model.RoomGroup](req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert request body: %w", err))
	}

	if err := s.repo.UpdateRoomGroup(e.Request().Context(), uint(roomGroupID), &roomGroup); err != nil {
		if errors.Is(err, repository.ErrRoomGroupNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Room group not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to update room group (roomGroupID: %d): %w", roomGroupID, err))
	}

	updatedRoomGroup, err := s.repo.GetRoomGroupByID(e.Request().Context(), uint(roomGroupID))

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get room group by ID: %w", err))
	}

	res, err := converter.Convert[api.RoomGroupResponse](updatedRoomGroup)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to convert response body: %w", err))
	}

	return e.JSON(http.StatusOK, res)
}

func (s *Server) AdminDeleteRoomGroup(
	e echo.Context,
	roomGroupID api.RoomGroupId,
	params api.AdminDeleteRoomGroupParams,
) error {
	user, err := s.repo.GetOrCreateUser(e.Request().Context(), *params.XForwardedUser)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to get or create user (userId: %s): %w", *params.XForwardedUser, err))
	}

	if !user.IsStaff {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if err := s.repo.DeleteRoomGroup(e.Request().Context(), uint(roomGroupID)); err != nil {
		if errors.Is(err, repository.ErrRoomGroupNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Room group not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to delete room group (roomGroupID: %d): %w", roomGroupID, err))
	}

	return e.NoContent(http.StatusNoContent)
}
