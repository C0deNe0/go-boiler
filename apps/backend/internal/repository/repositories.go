package repository

import "github.com/C0deNe0/go-boiler/internal/server"

type Repositories struct{}

func NewRepositories(s *server.Server) *Repositories {
	return &Repositories{}
}
