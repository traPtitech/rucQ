package gormrepository

import (
	"context"

	"gorm.io/gorm"

	"github.com/traPtitech/rucQ/model"
)

func (r *Repository) GetEvents(ctx context.Context, campID uint) ([]model.Event, error) {
	events, err := gorm.G[model.Event](r.db).
		Where(&model.Event{CampID: campID}).
		Find(ctx)

	if err != nil {
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
