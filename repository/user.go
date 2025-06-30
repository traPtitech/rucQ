//go:generate go tool mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package repository

import (
	"context"

	"github.com/traPtitech/rucQ/model"
)

type UserRepository interface {
	GetOrCreateUser(ctx context.Context, traqID string) (*model.User, error)
	GetUserTraqID(ID uint) (string, error)
	GetStaffs() ([]model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
}
