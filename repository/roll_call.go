//go:generate go tool mockgen -source=$GOFILE -destination=mockrepository/$GOFILE -package=mockrepository
package repository

import (
	"context"

	"github.com/traPtitech/rucQ/model"
)

type RollCallRepository interface {
	CreateRollCall(ctx context.Context, rollCall *model.RollCall) error
	GetRollCalls(ctx context.Context, campID uint) ([]model.RollCall, error)
}
