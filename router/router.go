package router

import "github.com/traPtitech/rucQ/repository"

type Server struct {
	repo  repository.Repository
	isDev bool
}

func NewServer(repo repository.Repository, isDev bool) *Server {
	return &Server{repo: repo, isDev: isDev}
}
