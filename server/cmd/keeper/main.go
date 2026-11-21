package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"keeper/endpoints/httpserver"
	"keeper/internal/auth"
	"keeper/internal/env"
	"keeper/internal/handlers/authhandler"
	"keeper/internal/handlers/notehandler"
	"keeper/internal/logger"
	"keeper/internal/notes"
	"keeper/internal/postgres"
)

func main() {
	os.Exit(run())
}

func run() int {
	stopAppChan := make(chan os.Signal, 1)
	signal.Notify(stopAppChan, syscall.SIGINT, syscall.SIGTERM)

	vars, err := env.Load()
	if err != nil {
		logger.Error(err)
		return 1
	}

	pool, err := postgres.NewConnection(context.Background(), vars.PostgresDSN)
	if err != nil {
		logger.Error(err)
		return 1
	}
	defer pool.Close()

	ttl := time.Duration(vars.AppJWTTTLHours) * time.Hour
	authService := auth.NewService(auth.NewRepository(pool), vars.AppJWTSecret, ttl)

	if len(os.Args) > 1 {
		if err := runUserCommand(context.Background(), authService, os.Args[1:]); err != nil {
			logger.Error(err)
			return 1
		}
		return 0
	}

	authHandler := authhandler.NewHandler(authService, ttl, vars.AppCookieSecure)
	noteHandler := notehandler.NewHandler(notes.NewService(notes.NewRepository(pool)))

	srv := httpserver.NewServer(vars, authHandler, noteHandler)
	go func() {
		if err := srv.Run(); err != nil {
			stopAppChan <- syscall.SIGQUIT
		}
	}()
	defer srv.Stop()

	sig := <-stopAppChan
	if sig == syscall.SIGQUIT {
		return 1
	}
	logger.Info("Received signal %s, stopping keeper", sig)
	return 0
}
