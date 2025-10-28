package gormrepository

import (
	"gorm.io/gorm"
	"github.com/traPtitech/rucQ/migration"
)

type Repository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) (*Repository, error) {
	repo := &Repository{db: db}
	err := migration.Migrate(db)
	
	return repo, err
}
