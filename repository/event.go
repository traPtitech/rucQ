package repository

import "github.com/traP-jp/rucQ/backend/model"

func (r *Repository) GetEvents() ([]model.Event, error) {
	var events []model.Event

	if err := r.db.Preload("Participants").Find(&events).Error; err != nil {
		return nil, err
	}

	return events, nil
}

func (r *Repository) GetEventByID(id uint) (*model.Event, error) {
	var event model.Event

	if err := r.db.Preload("Participants").First(&event, id).Error; err != nil {
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

func (r *Repository) UpdateEvent(ID uint, event *model.Event) error {
	if err := r.db.
		Where("id = ?", ID).
		Save(event).Error; err != nil {
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
