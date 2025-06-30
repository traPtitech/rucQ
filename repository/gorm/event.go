package gorm

import (
	"context"

	"github.com/traPtitech/rucQ/model"
	"gorm.io/gorm"
)

func (r *Repository) GetEvents() ([]model.Event, error) {
	var events []model.Event

	if err := r.db.Find(&events).Error; err != nil {
		return nil, err
	}

	return events, nil
}

func (r *Repository) GetEventByID(id uint) (*model.Event, error) {
	var event model.Event

	if err := r.db.First(&event, id).Error; err != nil {
		return nil, err
	}

	return &event, nil
}

func (r *Repository) CreateEvent(event *model.Event) error {
	if err := r.db.Create(event).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) UpdateEvent(ctx context.Context, ID uint, event *model.Event) error {
	if _, err := gorm.G[*model.Event](r.db).Where(&model.Event{
		Model: gorm.Model{
			ID: ID,
		},
	}).Updates(ctx, event); err != nil {
		return err
	}

	return nil
}

func (r *Repository) DeleteEvent(ID uint) error {
	if err := r.db.
		Delete(&model.Event{}, ID).Error; err != nil {
		return err
	}

	return nil
}
