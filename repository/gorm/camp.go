package gorm

import (
	"context"
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
)

func (r *Repository) CreateCamp(camp *model.Camp) error {
	if err := r.db.Create(camp).Error; err != nil {
		var mysqlErr *mysql.MySQLError

		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return model.ErrAlreadyExists
		}

		return err
	}

	return nil
}

func (r *Repository) GetCamps() ([]model.Camp, error) {
	var camps []model.Camp

	if err := r.db.Find(&camps).Error; err != nil {
		return nil, err
	}

	return camps, nil
}

func (r *Repository) GetCampByID(id uint) (*model.Camp, error) {
	var camp model.Camp

	if err := r.db.Where(&model.Camp{
		Model: gorm.Model{
			ID: id,
		},
	}).First(&camp).Error; err != nil {
		return nil, err
	}

	return &camp, nil
}

func (r *Repository) UpdateCamp(ctx context.Context, campID uint, camp *model.Camp) error {
	if _, err := gorm.G[*model.Camp](r.db).
		Where("id = ?", campID).
		Select("*").
		Updates(ctx, camp); err != nil {
		return err
	}

	return nil
}

func (r *Repository) DeleteCamp(ctx context.Context, campID uint) error {
	_, err := gorm.G[*model.Camp](r.db).Where(&model.Camp{
		Model: gorm.Model{
			ID: campID,
		},
	}).Delete(ctx)

	return err
}

func (r *Repository) AddCampParticipant(ctx context.Context, campID uint, user *model.User) error {
	camp, err := gorm.G[*model.Camp](r.db).Where(&model.Camp{
		Model: gorm.Model{
			ID: campID,
		},
	}).First(ctx)

	if err != nil {
		return err
	}

	if !camp.IsRegistrationOpen {
		return model.ErrForbidden
	}

	// Generics APIではまだAssociationが使えないため従来の書き方を使用
	// https://github.com/go-gorm/gorm/pull/7424#issuecomment-2918449411
	if err := r.db.Model(camp).Association("Participants").Append(user); err != nil {
		return err
	}

	return nil
}

func (r *Repository) RemoveCampParticipant(
	ctx context.Context,
	campID uint,
	user *model.User,
) error {
	camp, err := gorm.G[*model.Camp](r.db).Where(&model.Camp{
		Model: gorm.Model{
			ID: campID,
		},
	}).First(ctx)

	if err != nil {
		return err
	}

	if err := r.db.Model(camp).Association("Participants").Delete(user); err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetCampParticipants(ctx context.Context, campID uint) ([]model.User, error) {
	camp, err := gorm.G[*model.Camp](r.db).Preload("Participants", nil).Where(&model.Camp{
		Model: gorm.Model{
			ID: campID,
		},
	}).First(ctx)

	if err != nil {
		return nil, err
	}

	return camp.Participants, nil
}

func (r *Repository) IsCampParticipant(
	ctx context.Context,
	campID uint,
	userID string,
) (bool, error) {
	// まず、合宿が存在するかを確認
	_, err := gorm.G[*model.Camp](r.db).Where(&model.Camp{
		Model: gorm.Model{
			ID: campID,
		},
	}).First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, model.ErrNotFound
		}
		return false, err
	}

	// 合宿が存在する場合、参加者かどうかを確認
	var count int64
	err = r.db.WithContext(ctx).
		Table("camp_participants").
		Where("camp_id = ? AND user_id = ?", campID, userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
