package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"xrf197ilz35aq/internal"
	"xrf197ilz35aq/internal/processor"
	"xrf197ilz35aq/internal/server/api/handlers"
	"xrf197ilz35aq/internal/server/api/middleware"
)

func CreateServer(logger *slog.Logger, appConfig internal.AppConfig, processors *processor.Processors) *http.Server {
	serverMux := http.NewServeMux()

	reqHandlers := make([]handlers.RequestHandler, 0)

	// create request (routes) handlers
	assetReqHandler := handlers.NewAssetHandler(*logger)
	healthReqHandler := handlers.NewReqHealthHandlers(*logger)
	authReqHandler := handlers.NewAuthHandler(*logger, processors.AuthProcessor)
	userReqHandler := handlers.NewUserReqHandler(*logger, processors.UserProcessor)
	accountReqHandler := handlers.NewAccountHandler(*logger, processors.AccountProcessor)

	reqHandlers = append(
		reqHandlers,
		authReqHandler,
		userReqHandler,
		assetReqHandler,
		healthReqHandler,
		accountReqHandler,
	)

	for _, reqHandler := range reqHandlers {
		reqHandler.RegisterRoutes(serverMux)
	}

	// middlewares
	loggerMiddleware := middleware.NewLoggerHandler(logger)
	authMiddleware := middleware.NewAuthenticationMiddleware(*logger, processors.AuthProcessor)

	// wrap middlewares around the server
	handler := loggerMiddleware.Handler(
		authMiddleware.Handler(serverMux),
	)

	return &http.Server{
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 16 * time.Minute,
		IdleTimeout:  16 * time.Minute,
		Addr:         fmt.Sprintf(":%d", appConfig.Port),
	}
}
