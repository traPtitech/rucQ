package model

import "gorm.io/gorm"

type Payment struct {
	gorm.Model
	Amount     int
	AmountPaid int
	CampID     uint
	UserID     string
}
