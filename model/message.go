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
	SentAt       *time.Time // 送信時刻。nilの場合は未送信
}
