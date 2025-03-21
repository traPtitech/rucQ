package model

import "gorm.io/gorm"

type Room struct {
	gorm.Model
	Name    string
	CampID  uint
	Members []User `gorm:"many2many:room_members;ForeignKey:id;References:id"`
}
