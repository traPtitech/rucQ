package model

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrForbidden     = errors.New("forbidden")
	ErrNotFound      = gorm.ErrRecordNotFound
)

// 全モデルを書いておく
func GetAllModels() []any {
	return []any{
		&Camp{},
		&Event{},
		&User{},
		&Payment{},
		&QuestionGroup{},
		&Question{},
		&Option{},
		&Answer{},
		&Room{},
		&RoomGroup{},
		&Image{},
		&Message{},
		&RollCall{},
		&RollCallReaction{},
	}
}
