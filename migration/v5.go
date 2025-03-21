package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type v5EventParticipants struct {
	EventID uint `gorm:"primaryKey;column:event_id"`
	UserID  uint `gorm:"primaryKey;column:user_id"`
}

func (v5EventParticipants) TableName() string {
	return "event_participants"
}

func v5() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "5",
		Migrate: func(tx *gorm.DB) error {
			if !tx.Migrator().HasTable(&v5EventParticipants{}) {
				return tx.Migrator().CreateTable(&v5EventParticipants{})
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			if tx.Migrator().HasTable(&v5EventParticipants{}) {
				return tx.Migrator().DropTable(&v5EventParticipants{})
			}

			return nil
		},
	}
}
