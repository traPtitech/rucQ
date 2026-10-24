package model

import "gorm.io/gorm"

type RollCallReaction struct {
	gorm.Model
	Content string
	UserID  string `gorm:"uniqueIndex:idx_user_id_roll_call_id"`

	RollCallID uint `gorm:"index;uniqueIndex:idx_user_id_roll_call_id"`
}
