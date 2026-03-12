package migration

import (
	"fmt"

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
				// すでに同じDisplayIDを持つキャンプが存在する可能性があるのでチェック
				var duplicates []string
				if err := db.Model(&v7Camp{}).
					Select("display_id").
					Group("display_id").
					Having("COUNT(*) > 1").
					Scan(&duplicates).Error; err != nil {
					return err
				}

				if len(duplicates) > 0 {
					return fmt.Errorf("ユニークインデックス作成前に重複する display_id を解消してください: %v", duplicates)
				}

				return db.Migrator().CreateIndex(&v7Camp{}, "DisplayID")
			}
			return nil
		},
		Rollback: func(db *gorm.DB) error {
			return db.Migrator().DropIndex(&v7Camp{}, "idx_camps_display_id")
		},
	}
}
