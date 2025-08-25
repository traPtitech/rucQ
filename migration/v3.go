package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
)

// v3 adds messages table for scheduled message delivery
func v3() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202508250001",
		Migrate: func(db *gorm.DB) error {
			return db.AutoMigrate(&model.Message{})
		},
		Rollback: func(db *gorm.DB) error {
			return db.Migrator().DropTable(&model.Message{})
		},
	}
}
