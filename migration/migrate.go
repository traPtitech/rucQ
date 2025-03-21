package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/traP-jp/rucQ/backend/model"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, getAllMigrations())

	m.InitSchema(func(db *gorm.DB) error {
		return db.AutoMigrate(model.GetAllModels()...)
	})

	return m.Migrate()
}

func getAllMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		v1(),
		v2(),
		v3(),
		v4(),
		v5(), // イベント参加者の追加
	}
}
