package migration

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type v3Message struct {
	SentAt *time.Time
}

func (v3Message) TableName() string {
	return "messages"
}

func v3() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "3",
		Migrate: func(db *gorm.DB) error {
			return db.Migrator().AddColumn(&v3Message{}, "sent_at")
		},
		Rollback: func(db *gorm.DB) error {
			return db.Migrator().DropColumn(&v3Message{}, "sent_at")
		},
	}
}
