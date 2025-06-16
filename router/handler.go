package router

import (
	"gorm.io/gorm"

	"github.com/traP-jp/rucQ/backend/repository"
	gormRepository "github.com/traP-jp/rucQ/backend/repository/gorm"
)

type Server struct {
	repo repository.Repository
}

func NewServer(db *gorm.DB) *Server {
	return &Server{repo: gormRepository.NewGormRepository(db)}
}
