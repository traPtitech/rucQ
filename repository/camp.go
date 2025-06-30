//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package repository

import (
	"context"

	"github.com/traPtitech/rucQ/model"
)

type CampRepository interface {
	CreateCamp(camp *model.Camp) error
	GetCamps() ([]model.Camp, error)
	GetCampByID(id uint) (*model.Camp, error)
	UpdateCamp(campID uint, camp *model.Camp) error
	DeleteCamp(ctx context.Context, campID uint) error
	AddCampParticipant(ctx context.Context, campID uint, user *model.User) error
	RemoveCampParticipant(ctx context.Context, campID uint, user *model.User) error
	GetCampParticipants(ctx context.Context, campID uint) ([]model.User, error)
}
