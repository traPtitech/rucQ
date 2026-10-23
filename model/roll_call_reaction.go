package model

import "gorm.io/gorm"

type RollCallReaction struct {
	gorm.Model
	Content string
	UserID  string`gorm:"uniqueIndex:roll_call_reaction"` 

	RollCallID uint `gorm:"index;uniqueIndex:roll_call_reaction"` 
}
