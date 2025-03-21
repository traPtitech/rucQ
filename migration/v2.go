package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/traP-jp/rucQ/backend/model"
	"gorm.io/gorm"
)

func v2() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "2",
		Migrate: func(db *gorm.DB) error {
			return db.AutoMigrate(model.GetAllModels()...)
		},
	}
}
