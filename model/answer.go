package model

import "gorm.io/gorm"

type Answer struct {
	gorm.Model
	QuestionID uint      `gorm:"uniqueIndex:idx_question_id_user_id"`
	UserID     uint      `gorm:"uniqueIndex:idx_question_id_user_id"`
	Content    *[]string `gorm:"serializer:json"`
}
