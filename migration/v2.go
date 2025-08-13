package migration

import (
	"context"
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type v2Camp struct {
	gorm.Model
}

func (v2Camp) TableName() string {
	return "camps"
}

type v2Question struct {
	gorm.Model
}

func (v2Question) TableName() string {
	return "questions"
}

func v2() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "2",
		Migrate: func(db *gorm.DB) error {
			ctx := context.Background()

			if _, err := gorm.G[v2Camp](db).
				Where("created_at = ?", time.Time{}).
				Update(ctx, "created_at", gorm.Expr("updated_at")); err != nil {
				return fmt.Errorf("failed to update created_at in camps: %w", err)
			}

			if _, err := gorm.G[v2Question](db).
				Where("created_at = ?", time.Time{}).
				Update(ctx, "created_at", gorm.Expr("updated_at")); err != nil {
				return fmt.Errorf("failed to update created_at in questions: %w", err)
			}

			return nil
		},
	}
}
