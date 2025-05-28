package model

import "gorm.io/gorm"

type Dashboard struct {
	gorm.Model
	PaymentID uint
	RoomID    uint
	Payment   Payment
	Room      Room
}
