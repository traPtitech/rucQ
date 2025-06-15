package model

import "gorm.io/gorm"

type RoomGroup struct {
	gorm.Model
	Name  string
	Rooms []Room
}
