package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        string `gorm:"primaryKey;size:32"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	IsStaff   bool           `gorm:"index"`
	Answers   []Answer
	TraqUuid  string

	Payments        []Payment
	OrganizedEvents []Event   `gorm:"foreignKey:OrganizerID"`
	TargetMessages  []Message `gorm:"foreignKey:TargetUserID"`
}
