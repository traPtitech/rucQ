package repository

import "github.com/traP-jp/rucQ/backend/model"

type UserRepository interface {
	GetOrCreateUser(traqID string) (*model.User, error)
	GetUserTraqID(ID uint) (string, error)
	GetStaffs() ([]model.User, error)
	SetUserIsStaff(user *model.User, isStaff bool) error
}
