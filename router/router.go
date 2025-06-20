//go:generate go tool oapi-codegen -config ../oapi-codegen.yaml ../openapi.yaml
package router

import (
	"gorm.io/gorm"

	"github.com/traP-jp/rucQ/backend/repository"
	gormRepository "github.com/traP-jp/rucQ/backend/repository/gorm"
)

type Server struct {
	repo  repository.Repository
	debug bool
}

func NewServer(db *gorm.DB, debug bool) *Server {
	return &Server{repo: gormRepository.NewGormRepository(db), debug: debug}
}
