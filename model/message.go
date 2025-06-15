package model

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	TargetUserID string
	Content      string
	SendAt       time.Time
}
