package gorm

import (
	"context"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
)

func (r *Repository) GetOrCreateUser(ctx context.Context, userID string) (*model.User, error) {
	users, err := gorm.G[*model.User](r.db).Limit(1).Where(&model.User{ID: userID}).Find(ctx)

	if err != nil {
		return nil, err
	}

	if len(users) > 0 {
		return users[0], nil
	}

	user := model.User{
		ID: userID,
	}

	if err := gorm.G[model.User](r.db).Create(ctx, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetUserTraqID(ID uint) (string, error) {
	var user model.User

	if err := r.db.Where("id = ?", ID).Find(&user).Error; err != nil {
		return "", err
	}

	return user.ID, nil
}

func (r *Repository) GetStaffs() ([]model.User, error) {
	var staffs []model.User

	if err := r.db.Where(&model.User{IsStaff: true}).Find(&staffs).Error; err != nil {
		return nil, err
	}

	return staffs, nil
}

func (r *Repository) UpdateUser(ctx context.Context, user *model.User) error {
	_, err := gorm.G[*model.User](r.db).Updates(ctx, user)

	return err
}
