package model

import "gorm.io/gorm"

type RoomStatus struct {
	gorm.Model
	RoomID uint   `gorm:"not null;uniqueIndex"`
	Type   string `gorm:"not null;size:8"`
	Topic  string `gorm:"not null;size:64"`
}

type RoomStatusLog struct {
	gorm.Model
	RoomID     uint   `gorm:"not null;index"`
	Type       string `gorm:"not null;size:8"`
	Topic      string `gorm:"not null;size:64"`
	OperatorID string `gorm:"not null;size:32"`
}
