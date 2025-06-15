package model

import (
	"time"

	"gorm.io/gorm"
)

type QuestionGroup struct {
	gorm.Model
	Name        string
	Description *string
	Due         time.Time
	Questions   []Question

	CampID uint
}
