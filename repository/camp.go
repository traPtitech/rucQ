//go:generate go tool mockgen -source=$GOFILE -destination=mockrepository/$GOFILE -package=mockrepository
package repository

import (
	"context"
	"errors"

	"github.com/traPtitech/rucQ/model"
)

var (
	ErrCampNotFound        = errors.New("camp not found")
	ErrParticipantNotFound = errors.New("participant not found")
)

type CampRepository interface {
	CreateCamp(camp *model.Camp) error
	GetCamps() ([]model.Camp, error)
	GetCampByID(id uint) (*model.Camp, error)
	UpdateCamp(ctx context.Context, campID uint, camp *model.Camp) error
	DeleteCamp(ctx context.Context, campID uint) error
	AddCampParticipant(ctx context.Context, campID uint, user *model.User) error
	RemoveCampParticipant(ctx context.Context, campID uint, user *model.User) error
	GetCampParticipants(ctx context.Context, campID uint) ([]model.User, error)
	IsCampParticipant(ctx context.Context, campID uint, userID string) (bool, error)
}
