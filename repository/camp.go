package repository

import "github.com/traP-jp/rucQ/backend/model"

type CampRepository interface {
	CreateCamp(camp *model.Camp) error
	GetCamps() ([]model.Camp, error)
	GetCampByID(id uint) (*model.Camp, error)
	UpdateCamp(campID uint, camp *model.Camp) error
}
