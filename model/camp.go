package model

import (
	"time"

	"gorm.io/gorm"
)

type Camp struct {
	gorm.Model
	DisplayID          string
	Name               string
	Guidebook          string
	IsDraft            bool
	IsPaymentOpen      bool
	IsRegistrationOpen bool
	DateStart          time.Time
	DateEnd            time.Time

	Participants   []User `gorm:"many2many:camp_participants;"`
	Payments       []Payment
	Events         []Event
	QuestionGroups []QuestionGroup
	RoomGroups     []RoomGroup
	Images         []Image
}
