//go:generate go tool mockgen -source=$GOFILE -destination=mockrepository/$GOFILE -package=mockrepository
package repository

import (
	"context"
	"errors"

	"github.com/traPtitech/rucQ/model"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	GetOrCreateUser(ctx context.Context, traqID string) (*model.User, error)
	GetUserTraqID(ID uint) (string, error)
	GetStaffs() ([]model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
}
