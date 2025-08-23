package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/C0deNe0/go-boiler/internal/config"
	"github.com/C0deNe0/go-boiler/internal/database"
	"github.com/C0deNe0/go-boiler/internal/handler"
	"github.com/C0deNe0/go-boiler/internal/logger"
	"github.com/C0deNe0/go-boiler/internal/repository"
	"github.com/C0deNe0/go-boiler/internal/router"
	"github.com/C0deNe0/go-boiler/internal/server"
	"github.com/C0deNe0/go-boiler/internal/service"
)

const DefaultContextTimeout = 30

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config:" + err.Error())
	}

	//init newrelic loggerService
	loggerService := logger.NewLoggerService(cfg.Observeability)
	defer loggerService.Shutdown()

	log := logger.NewLoggerWithService(cfg.Observeability, loggerService)
	if cfg.Primary.Env != "local" {
		if err := database.Migrate(context.Background(), &log, cfg); err != nil {
			log.Fatal().Err(err).Msg("failed to migrate database")
		}
	}

	//init server
	srv, err := server.New(cfg, &log, loggerService)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init server")
	}

	//init repo , handler , service

	repos := repository.NewRepositories(srv)
	services, serviceErr := service.NewServices(srv, repos)
	if serviceErr != nil {
		log.Fatal().Err(serviceErr).Msg("could not create services")

	}
	handlers := handler.NewHandlers(srv, services)

	//init router
	r := router.NewRouter(srv, handlers, services)

	//setup httpserver
	srv.SetupHTTPServer(r)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	go func() {
		if err = srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultContextTimeout*time.Second)

	if err = srv.ShutDown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	stop()
	cancel()

	log.Info().Msg("server exited properly")
}
