package repository

import "github.com/traP-jp/rucQ/backend/model"

func (r *Repository) CreateQuestionGroup(questionGroup *model.QuestionGroup) error {
	if err := r.db.Create(questionGroup).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetQuestionGroups() ([]model.QuestionGroup, error) {
	var questionGroups []model.QuestionGroup

	if err := r.db.
		Preload("Questions").
		Preload("Questions.Options").
		Find(&questionGroups).
		Error; err != nil {
		return nil, err
	}

	return questionGroups, nil
}

func (r *Repository) GetQuestionGroup(ID uint) (*model.QuestionGroup, error) {
	var questionGroup model.QuestionGroup

	if err := r.db.
		Preload("Questions").
		Preload("Questions.Options").
		Where("id = ?", ID).
		First(&questionGroup).
		Error; err != nil {
		return nil, err
	}

	return &questionGroup, nil
}

func (r *Repository) UpdateQuestionGroup(ID uint, questionGroup *model.QuestionGroup) error {
	if err := r.db.
		Where("id = ?", ID).
		Updates(questionGroup).
		Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) DeleteQuestionGroup(ID uint) error {
	if err := r.db.
		Delete(&model.QuestionGroup{}, ID).
		Error; err != nil {
		return err
	}

	return nil
}
