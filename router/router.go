package router

import (
	"github.com/traPtitech/rucQ/repository"
	"github.com/traPtitech/rucQ/service"
)

type Server struct {
	repo        repository.Repository
	traqService service.TraqService
	isDev       bool
}

func NewServer(repo repository.Repository, traqService service.TraqService, isDev bool) *Server {
	return &Server{
		repo:        repo,
		traqService: traqService,
		isDev:       isDev,
	}
}
