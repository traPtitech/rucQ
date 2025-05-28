package model

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = gorm.ErrRecordNotFound
)

// 全モデルを書いておく
func GetAllModels() []any {
	return []any{
		&Camp{},
		&Event{},
		&User{},
		&Budget{},
		&Payment{},
		&QuestionGroup{},
		&Question{},
		&Option{},
		&Answer{},
		&Room{},
		&RoomGroup{},
		&Dashboard{},
		&Image{},
		&Message{},
		&RollCall{},
		&RollCallReaction{},
	}
}
