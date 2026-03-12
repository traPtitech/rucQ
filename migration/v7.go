package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type v7Camp struct {
	DisplayID string `gorm:"uniqueIndex:idx_camps_display_id"`
}

func (v7Camp) TableName() string {
	return "camps"
}

func v7() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "7",
		Migrate: func(db *gorm.DB) error {
			if !db.Migrator().HasIndex(&v7Camp{}, "idx_camps_display_id") {
       			return db.Migrator().CreateIndex(&v7Camp{}, "DisplayID")
    		}
		},
		Rollback: func(db *gorm.DB) error {
			return db.Migrator().DropIndex(&v7Camp{}, "idx_camps_display_id")
		},
	}
}
