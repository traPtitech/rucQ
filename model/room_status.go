package model

import "gorm.io/gorm"

type RoomStatus struct {
	gorm.Model
	RoomID uint   `gorm:"not null;uniqueIndex"`
	Room   *Room  `gorm:"foreignKey:RoomID;references:ID;constraint:OnDelete:CASCADE"`
	Type   string `gorm:"not null;size:8"`
	Topic  string `gorm:"not null;size:64"`
}

type RoomStatusLog struct {
	gorm.Model
	RoomID     uint   `gorm:"not null"`
	Room       *Room  `gorm:"foreignKey:RoomID;references:ID;constraint:OnDelete:CASCADE"`
	Type       string `gorm:"not null;size:8"`
	Topic      string `gorm:"not null;size:64"`
	OperatorID string `gorm:"not null;size:32"`
	Operator   *User  `gorm:"foreignKey:OperatorID;references:ID;constraint:OnDelete:RESTRICT"`
}
