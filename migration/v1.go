package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type V1Questions struct {
	IsRequired bool `gorm:"not null;default:false"`
}

func (V1Questions) TableName() string {
	return "questions"
}

func v1() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "1",
		Migrate: func(db *gorm.DB) error {
			return db.Migrator().AddColumn(&V1Questions{}, "is_required")
		},
		Rollback: func(db *gorm.DB) error {
			return db.Migrator().DropColumn(&V1Questions{}, "is_required")
		},
	}
}
