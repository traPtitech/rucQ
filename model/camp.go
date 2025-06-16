package model

import "gorm.io/gorm"

type Camp struct {
	gorm.Model
	DisplayID          string
	Name               string
	Description        string
	IsDraft            bool
	IsPaymentOpen      bool
	IsRegistrationOpen bool

	Participants   []User `gorm:"many2many:camp_participants;"`
	Payments       []Payment
	Events         []Event
	QuestionGroups []QuestionGroup
	RoomGroups     []RoomGroup
	Images         []Image
}
