package repository

import (
	"github.com/traP-jp/rucQ/backend/model"
	"gorm.io/gorm"
)

func (r *Repository) GetParticipants(eventID uint) ([]model.User, error) {
	event, err := r.GetEventByID(eventID)

	if err != nil {
		return nil, err
	}

	return event.Participants, nil
}

func (r *Repository) RegisterEvent(eventID uint, participant *model.User) error {
	event, err := r.GetEventByID(eventID)

	if err != nil {
		return err
	}

	count := r.db.
		Model(event).
		Where(&model.User{TraqID: participant.TraqID}).
		Association("Participants").
		Count()

	if count == 1 {
		// 同じデータが見つかった場合は作成せずに終了
		return nil
	}

	// TODO: 1より大きい場合はエラーを返す

	if err := r.db.
		Model(event).
		Association("Participants").
		Append(participant); err != nil {
		return err
	}

	return nil
}

func (r *Repository) UnregisterEvent(eventID uint, user *model.User) error {
	if err := r.db.
		Where(&model.Event{Model: gorm.Model{ID: eventID}}).
		Association("Participants").
		Delete(user); err != nil {
		return err
	}

	return nil
}
