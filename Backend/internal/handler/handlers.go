package handler

import (
	"github.com/C0deNe0/go-boiler/internal/server"
	"github.com/C0deNe0/go-boiler/internal/service"
)

type Handlers struct{}

func NewHandlers(s *server.Server, services *service.Services) *Handlers {
	return &Handlers{}
}
