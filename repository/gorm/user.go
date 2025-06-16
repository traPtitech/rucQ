package gorm

import (
	"context"
	"fmt"
	"os"

	traq "github.com/traPtitech/go-traq"
	"gorm.io/gorm"

	"github.com/traP-jp/rucQ/backend/model"
)

func (r *Repository) GetOrCreateUser(ctx context.Context, traqID string) (*model.User, error) {
	// まずデータベースを検索
	users, err := gorm.G[*model.User](r.db).Limit(1).Where(&model.User{ID: traqID}).Find(ctx)

	if err != nil {
		return nil, err
	}

	var user model.User

	if len(users) > 0 {
		user = *users[0]
	}

	// if user.TraqUUID != "" {
	// 	return &user, nil
	// }

	configuration := traq.NewConfiguration()
	apiClient := traq.NewAPIClient(configuration)
	configuration.AddDefaultHeader("Authorization", "Bearer "+os.Getenv("BOT_ACCESS_TOKEN"))
	usersUuid, httpResp, err := apiClient.UserApi.GetUsers(context.Background()).Name(traqID).Execute()
	if err != nil {
		return nil, fmt.Errorf("error when calling UserApi.GetUsers: %w\nfull HTTP response: %v", err, httpResp)
	}

	// traQ API のレスポンスをチェック
	if len(usersUuid) != 1 {
		return nil, fmt.Errorf("no users found with name %s", traqID)
	}

	// 追加、更新するユーザーを作成
	// user.TraqUUID = usersUuid[0].Id
	user.ID = traqID

	if err := r.db.Save(&user).Error; err != nil {
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
