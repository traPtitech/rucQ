package model

import (
	"time"

	"gorm.io/gorm"
)

type QuestionGroup struct {
	gorm.Model
	CampID      uint
	Name        string
	Description string
	Due         time.Time
	Questions   []Question
}
