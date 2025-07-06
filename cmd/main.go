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
	"xrf197ilz35aq/internal/client"
	"xrf197ilz35aq/internal/processor"
	"xrf197ilz35aq/internal/server/api"
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

	/// Create API Client
	defaultHeaders := make(map[string]string)
	defaultHeaders["Content-Type"] = "application/json"
	apiClient := client.NewApiClient(
		config.Service.Organization.BaseURL,
		*logger,
		client.WithTimeout(config.Service.Organization.APIClientTimeout),
		client.WithDefaultHeader(defaultHeaders))

	/// Create request processors
	userProcessor := processor.NewUserProcessor(*logger, *apiClient)
	processors := processor.Processors{UserProcessor: *userProcessor}

	server := api.CreateServer(*logger, config.Application, &processors)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("serverStarted=false :: error starting api server", "error", err)
		}
	}()

	logger.Info("***** xrf197ilz35aq started successfully *******", "port", config.Application.Port)

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
