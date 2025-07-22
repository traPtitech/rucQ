package gormrepository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
)

func (r *Repository) GetOrCreateUser(ctx context.Context, userID string) (*model.User, error) {
	// 先に作成を試す（取得を先に行うと、同時に複数のリクエストが来た場合に競合が発生する可能性があるため）
	user := model.User{
		ID: userID,
	}

	err := gorm.G[model.User](r.db).Create(ctx, &user)

	if err == nil {
		// ユーザーが正常に作成された場合は、そのユーザーを返す
		return &user, nil
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		// 重複キーエラーが発生した場合は、すでに存在するユーザーを取得する
		users, err := gorm.G[*model.User](r.db).Limit(1).Where(&model.User{ID: userID}).Find(ctx)

		if err != nil {
			return nil, err
		}

		if len(users) > 0 {
			return users[0], nil
		}

		return nil, fmt.Errorf("user not found after duplicate key error: %s", userID)
	}

	// その他の予期せぬエラー
	return nil, err
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
	_, err := gorm.G[*model.User](r.db).Select("is_staff").Updates(ctx, user)

	return err
}
