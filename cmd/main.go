package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	accountV1 "xrf197ilz35aq/gen/account/v1"
	xrfq3V1 "xrf197ilz35aq/gen/xrfq3/v1"
	"xrf197ilz35aq/internal"
	"xrf197ilz35aq/internal/client"
	"xrf197ilz35aq/internal/client/grpc"
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
	parsedUrl, err := url.Parse(config.Service.Organization.BaseURL)
	if err != nil {
		logger.Error("failed to parse organization base url", "err", err)
	}
	apiClient := client.NewApiClient(parsedUrl.String(), config.Application, client.WithTimeout(config.Service.Organization.APIClientTimeout), client.WithDefaultHeader(defaultHeaders))

	////// Create gRPC connection
	xrfQ3CertPath := "local/secrets/ssl/server.crt"
	connManager := grpc.NewConnectionManager(nil)
	xrfQ3Conn, err := connManager.CreateOrGetConnection(config.Service.Account.Address, *logger, xrfQ3CertPath)
	if err != nil {
		logger.Error("failed to create xrfQ3 connection", "err", err)
		return
	}

	////// register gRPC client
	acctServiceClient := accountV1.NewAccountServiceClient(xrfQ3Conn)
	xrfQ3AppServiceClient := xrfq3V1.NewAppServiceClient(xrfQ3Conn)

	///// Check xrfQ3 gRPC service health
	err = checkXrfQ3Health(context.Background(), xrfQ3AppServiceClient, *logger)

	if err != nil {
		logger.Error("XRF-Q3 app is not running", "err", err)
		return
	}

	///// Create request processors
	userProcessor := processor.NewUserProcessor(*apiClient)
	authProcessor := processor.NewAuthProcessor(*apiClient)
	accountProcessor := processor.NewAccountProcessor(acctServiceClient)

	processors := processor.Processors{
		UserProcessor:    *userProcessor,
		AuthProcessor:    *authProcessor,
		AccountProcessor: accountProcessor,
	}

	server := api.CreateServer(logger, config.Application, &processors)
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

	ctx, cancel := context.WithTimeout(context.Background(), config.Application.GracefulTimeout)
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

func checkXrfQ3Health(ctx context.Context, xrfQ3RPCClient xrfq3V1.AppServiceClient, log slog.Logger) error {
	resp, err := xrfQ3RPCClient.CheckHealth(ctx, &xrfq3V1.CheckHealthRequest{})
	if err != nil {
		return fmt.Errorf("xrfQ3RPC-service failure, err:: %w", err)
	}
	if !resp.IsUp {
		return fmt.Errorf("xrfQ3RPC is NOT running")
	}
	log.Info("xrfQ3RPC running healthy...", "port", resp.Region, "appId", resp.AppId)
	return nil
}
