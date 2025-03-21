package repository

import "github.com/traP-jp/rucQ/backend/model"

func (r *Repository) CreateAnswer(answer *model.Answer) error {
	if err := r.db.Create(answer).Error; err != nil {
		return err
	}

	return nil
}

type GetAnswerQuery struct {
	QuestionID uint
	UserID     uint
}

func (r *Repository) GetOrCreateAnswer(query *GetAnswerQuery) (*model.Answer, error) {
	var answer model.Answer

	if err := r.db.
		Where(&model.Answer{
			QuestionID: query.QuestionID,
			UserID:     query.UserID,
		}).
		FirstOrCreate(&answer).
		Error; err != nil {
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
