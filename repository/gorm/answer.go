package gorm

import "github.com/traP-jp/rucQ/backend/model"

func (r *Repository) CreateAnswer(answer *model.Answer) error {
	if err := r.db.Create(answer).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetAnswerByID(id uint) (*model.Answer, error) {
	var answer model.Answer

	if err := r.db.First(&answer, id).Error; err != nil {
		return nil, err
	}

	return &answer, nil
}

func (r *Repository) UpdateAnswer(answer *model.Answer) error {
	if err := r.db.Save(answer).Error; err != nil {
		return err
	}

	return nil
}
