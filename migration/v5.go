package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type v5RoomStatus struct {
	gorm.Model
	RoomID uint    `gorm:"not null;uniqueIndex"`
	Room   *v5Room `gorm:"foreignKey:RoomID;references:ID;constraint:OnDelete:CASCADE"`
	Type   *string `gorm:"size:8"`
	Topic  string  `gorm:"not null;size:64"`
}

func (v5RoomStatus) TableName() string {
	return "room_statuses"
}

type v5RoomStatusLog struct {
	gorm.Model
	RoomID     uint    `gorm:"not null"`
	Room       *v5Room `gorm:"foreignKey:RoomID;references:ID;constraint:OnDelete:CASCADE"`
	Type       *string `gorm:"size:8"`
	Topic      string  `gorm:"not null;size:64"`
	OperatorID string  `gorm:"not null;size:32"`
	Operator   *v5User `gorm:"foreignKey:OperatorID;references:ID;constraint:OnDelete:RESTRICT"`
}

func (v5RoomStatusLog) TableName() string {
	return "room_status_logs"
}

type v5Room struct {
	gorm.Model
}

func (v5Room) TableName() string {
	return "rooms"
}

type v5User struct {
	ID string `gorm:"primaryKey;size:32"`
}

func (v5User) TableName() string {
	return "users"
}

func v5() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "5",
		Migrate: func(db *gorm.DB) error {
			if err := db.Migrator().CreateTable(&v5RoomStatus{}); err != nil {
				return err
			}
			if err := db.Migrator().CreateTable(&v5RoomStatusLog{}); err != nil {
				return err
			}

			return nil
		},
		Rollback: func(db *gorm.DB) error {
			if err := db.Migrator().DropTable(&v5RoomStatusLog{}); err != nil {
				return err
			}

			return db.Migrator().DropTable(&v5RoomStatus{})
		},
	}
}
