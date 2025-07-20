package gormrepository

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}
