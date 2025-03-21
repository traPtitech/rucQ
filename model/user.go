package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	TraqID   string `gorm:"primaryKey;size:32"` // 主キー
	IsStaff  bool   `gorm:"index"`
	Answers  []Answer
	TraqUuid string

	Budgets []Budget
}
