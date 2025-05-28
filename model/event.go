package model

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	gorm.Model
	Name            string
	Description     string
	Location        string
	TimeStart       *time.Time
	TimeEnd         *time.Time
	Time            *time.Time // For moment events
	CampID          uint
	OrganizerTraqID string
	Type            string // "duration", "moment", "official"
	DisplayColor    string
}
