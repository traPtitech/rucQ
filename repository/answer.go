package repository

import "github.com/traP-jp/rucQ/backend/model"

type AnswerRepository interface {
	CreateAnswer(answer *model.Answer) error
	GetAnswerByID(id uint) (*model.Answer, error)
	UpdateAnswer(answer *model.Answer) error
}
