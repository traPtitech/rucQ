package model

import "gorm.io/gorm"

type Camp struct {
	gorm.Model
	DisplayID          string
	Name               string
	Description        string
	IsDraft            bool `gorm:"index"`
	IsPaymentOpen      bool
	IsRegistrationOpen bool

	Payments       []Payment
	Events         []Event
	QuestionGroups []QuestionGroup
	Rooms          []Room
}
