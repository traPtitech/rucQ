package model

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	Content string
	SendAt  time.Time
}
