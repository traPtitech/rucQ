package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type v5RoomStatus struct {
	gorm.Model
	RoomID uint   `gorm:"not null;uniqueIndex"`
	Type   string `gorm:"not null;size:8"`
	Topic  string `gorm:"not null;size:64"`
}

func (v5RoomStatus) TableName() string {
	return "room_statuses"
}

type v5RoomStatusLog struct {
	gorm.Model
	RoomID     uint   `gorm:"not null;index"`
	Type       string `gorm:"not null;size:8"`
	Topic      string `gorm:"not null;size:64"`
	OperatorID string `gorm:"not null;size:32"`
}

func (v5RoomStatusLog) TableName() string {
	return "room_status_logs"
}

func v5() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "5",
		Migrate: func(db *gorm.DB) error {
			return db.AutoMigrate(&v5RoomStatus{}, &v5RoomStatusLog{})
		},
		Rollback: func(db *gorm.DB) error {
			if err := db.Migrator().DropTable(&v5RoomStatusLog{}); err != nil {
				return err
			}

			return db.Migrator().DropTable(&v5RoomStatus{})
		},
	}
}
