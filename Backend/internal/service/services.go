package service

import (
	"github.com/C0deNe0/go-boiler/internal/lib/job"
	"github.com/C0deNe0/go-boiler/internal/repository"
	"github.com/C0deNe0/go-boiler/internal/server"
)

type Services struct {
	Auth *AuthService
	Job  *job.JobService
}

func NewServices(s *server.Server,repo *repository.Repositories)(*Services,error){
	authService := NewAuthService(s)
	return  &Services{
		Job: s.Job,
		Auth: authService,

	},nil
}