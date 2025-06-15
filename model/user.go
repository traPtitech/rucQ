package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        string `gorm:"primaryKey;size:32"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	IsStaff   bool           `gorm:"index"`
	TraqUUID  uuid.UUID

	Answers         []Answer
	Payments        []Payment
	OrganizedEvents []Event   `gorm:"foreignKey:OrganizerID"`
	TargetMessages  []Message `gorm:"foreignKey:TargetUserID"`
}
