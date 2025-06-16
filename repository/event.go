package repository

import "github.com/traP-jp/rucQ/backend/model"

type EventRepository interface {
	GetEvents() ([]model.Event, error)
	GetEventByID(id uint) (*model.Event, error)
	CreateEvent(event *model.Event) error
	UpdateEvent(ID uint, event *model.Event) error
	DeleteEvent(ID uint) error
}
