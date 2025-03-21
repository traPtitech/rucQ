package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/traP-jp/rucQ/backend/model"
	"gorm.io/gorm"
)

// roomについてのAPIを追加
func v4() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "4",
		Migrate: func(db *gorm.DB) error {
			return db.AutoMigrate(model.GetAllModels()...)
		},
	}
}
