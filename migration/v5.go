package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func v5() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "5",
		Migrate: func(db *gorm.DB) error {
			return db.Exec(
				"ALTER TABLE camps ADD UNIQUE INDEX idx_camps_display_id (display_id)",
			).Error
		},
		Rollback: func(db *gorm.DB) error {
			return db.Exec("ALTER TABLE camps DROP INDEX idx_camps_display_id").Error
		},
	}
}
