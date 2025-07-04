package router

import "github.com/traPtitech/rucQ/repository"

type Server struct {
	repo  repository.Repository
	debug bool
}

func NewServer(repo repository.Repository, debug bool) *Server {
	return &Server{repo: repo, debug: debug}
}
