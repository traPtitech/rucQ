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
	TimeStart       time.Time
	TimeEnd         time.Time
	CampID          uint
	OrganizerTraqID string
	ByStaff         bool
	DisplayColor    string
	Participants    []User `gorm:"many2many:event_participants;ForeignKey:id;References:id"`
}
