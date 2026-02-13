package model

import "gorm.io/gorm"

type Room struct {
	gorm.Model
	Name    string
	Members []User     `gorm:"many2many:room_members"`
	Status  RoomStatus `gorm:"constraint:OnDelete:CASCADE"`

	RoomGroupID uint
}
