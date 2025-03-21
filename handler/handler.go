package handler

import (
	"github.com/traP-jp/rucQ/backend/repository"
	"gorm.io/gorm"
)

type Server struct {
	repo *repository.Repository
}

func NewServer(db *gorm.DB) *Server {
	return &Server{repo: repository.NewRepository(db)}
}
