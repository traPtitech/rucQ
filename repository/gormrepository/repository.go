package gormrepository

import (
	"github.com/traPtitech/rucQ/migration"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) (*Repository, error) {
	repo := &Repository{db: db}
	err := migration.Migrate(db)
	
	return repo, err
}
