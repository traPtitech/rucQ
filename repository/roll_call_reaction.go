//go:generate go tool mockgen -source=$GOFILE -destination=mockrepository/$GOFILE -package=mockrepository
package repository

import (
	"context"
	"errors"

	"github.com/traPtitech/rucQ/model"
)

var ErrRollCallReactionNotFound = errors.New("roll call reaction not found")

type RollCallReactionRepository interface {
	CreateRollCallReaction(ctx context.Context, reaction *model.RollCallReaction) error
	GetRollCallReactions(ctx context.Context, rollCallID uint) ([]model.RollCallReaction, error)
	GetRollCallReactionByID(ctx context.Context, id uint) (*model.RollCallReaction, error)
	UpdateRollCallReaction(ctx context.Context, id uint, reaction *model.RollCallReaction) error
	DeleteRollCallReaction(ctx context.Context, id uint) error
}
