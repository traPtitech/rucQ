package repository

import (
	"context"

	"github.com/traP-jp/rucQ/backend/model"
)

type EventRepository interface {
	GetEvents() ([]model.Event, error)
	GetEventByID(id uint) (*model.Event, error)
	CreateEvent(event *model.Event) error
	UpdateEvent(ctx context.Context, ID uint, event *model.Event) error
	DeleteEvent(ID uint) error
}
