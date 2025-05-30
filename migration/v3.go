package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/traP-jp/rucQ/backend/model"
	"gorm.io/gorm"
)

// UserIDを使う形式に戻す
func v3() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "3",
		Migrate: func(db *gorm.DB) error {
			return db.AutoMigrate(model.GetAllModels()...)
		},
	}
}
