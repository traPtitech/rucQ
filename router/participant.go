package router

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/traPtitech/rucQ/repository"
)

// requireCampParticipant ユーザーが合宿の参加者でなければ403を返す
func (s *Server) requireCampParticipant(ctx context.Context, campID uint, userID string) error {
	isParticipant, err := s.repo.IsCampParticipant(ctx, campID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrCampNotFound) {
			return echo.ErrNotFound
		}

		return echo.NewHTTPError(http.StatusInternalServerError).
			SetInternal(fmt.Errorf("failed to check camp participation: %w", err))
	}

	if !isParticipant {
		return echo.ErrForbidden
	}

	return nil
}
