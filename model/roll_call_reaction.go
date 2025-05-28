package model

import "gorm.io/gorm"

type RollCallReaction struct {
	gorm.Model
	Content string
	UserID  string
}
