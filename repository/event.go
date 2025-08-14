//go:generate go tool mockgen -source=$GOFILE -destination=mockrepository/$GOFILE -package=mockrepository
package repository

import (
	"context"

	"github.com/traPtitech/rucQ/model"
)

type EventRepository interface {
	GetEvents(ctx context.Context, campID uint) ([]model.Event, error)
	GetEventByID(id uint) (*model.Event, error)
	CreateEvent(event *model.Event) error
	UpdateEvent(ctx context.Context, ID uint, event *model.Event) error
	DeleteEvent(ID uint) error
}
