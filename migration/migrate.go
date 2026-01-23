package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
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
		v1(), // questionsテーブルにis_requiredカラムを追加
		v2(), // ゼロ値で上書きされてしまっていたcreated_atを修正
		v3(), // messagesテーブルにsent_atカラムを追加
		v4(), // roll_callsテーブルにcamp_idカラムを追加
		v5(), // 同一displayIdのキャンプが存在しないことを確認
	}
}
