package model

import (
	"time"

	"gorm.io/gorm"
)

type EventType string

const (
	EventTypeDuration EventType = "duration"
	EventTypeMoment   EventType = "moment"
	EventTypeOfficial EventType = "official"
)

type Event struct {
	gorm.Model
	Type         EventType `gorm:"type:enum('duration', 'moment', 'official')"`
	Name         string
	Description  string
	Location     string
	TimeStart    time.Time // MomentEventのtimeもTimeStartとして扱う
	TimeEnd      *time.Time
	OrganizerID  *string
	DisplayColor *string

	CampID uint
}
