package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"xrf197ilz35aq/internal"
	xrfHttp "xrf197ilz35aq/internal/server/http"
)

func main() {
	env := getAppEnv()
	config, err := internal.GetConfig(strings.ToLower(env))
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, err := internal.SetupLogger(env, config.Log)
	if err != nil {
		log.Fatalf("Failed to setup logger: %v", err)
	}

	server := xrfHttp.CreateServer(*logger, config.Application)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("serverStarted=false :: error starting http server", "error", err)
		}
	}()

	logger.Info("***** xrf197ilz35aq started successfully *******")

	ch := make(chan os.Signal, 1)
	// Accept graceful shutdowns when quit via SIGINT (Ctrl+C), SIGTERM (used by Docker, Kubernetes) or SIGQUIT.
	// SIGKILL will not be caught.
	signal.Notify(ch, os.Interrupt)
	signal.Notify(ch, syscall.SIGTERM)
	signal.Notify(ch, syscall.SIGQUIT)

	// Block until we receive shutdown signal.
	<-ch
	logger.Info("received shutdown signal, starting graceful shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server.
	err = server.Shutdown(ctx)
	if err != nil {
		logger.Error("xrf197ilz35aq shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("xrf197ilz35aq shutdown successfully")
	os.Exit(0)
}

func getAppEnv() string {
	env, ok := os.LookupEnv(internal.Environment)
	if !ok || env == "" {
		env = internal.DevelopEnv
	}

	switch env {
	case internal.StagingEnv:
		return "STAGING"
	case internal.ProductionEnv, internal.LiveEnv:
		return internal.LiveEnv
	default:
		return internal.DevelopEnv
	}
}
