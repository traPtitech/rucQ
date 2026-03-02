package gormrepository

import (
	"context"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
)

func (r *Repository) CreateActivity(ctx context.Context, activity *model.Activity) error {
	return gorm.G[model.Activity](r.db).Create(ctx, activity)
}

func (r *Repository) GetActivitiesByCampID(
	ctx context.Context,
	campID uint,
) ([]model.Activity, error) {
	return gorm.G[model.Activity](r.db).
		Where("camp_id = ?", campID).
		Order("created_at DESC").
		Find(ctx)
}
