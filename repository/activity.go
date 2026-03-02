//go:generate go tool mockgen -source=$GOFILE -destination=mockrepository/$GOFILE -package=mockrepository
package repository

import (
	"context"

	"github.com/traPtitech/rucQ/model"
)

type ActivityRepository interface {
	CreateActivity(ctx context.Context, activity *model.Activity) error
	GetActivitiesByCampID(ctx context.Context, campID uint) ([]model.Activity, error)
}
