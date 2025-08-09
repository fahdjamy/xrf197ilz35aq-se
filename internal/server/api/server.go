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
	healthReqHandler := handlers.NewReqHealthHandlers(*logger)
	authReqHandler := handlers.NewAuthHandler(*logger, processors.AuthProcessor)
	userReqHandler := handlers.NewUserReqHandler(*logger, processors.UserProcessor)

	reqHandlers = append(reqHandlers, healthReqHandler)
	reqHandlers = append(reqHandlers, userReqHandler)
	reqHandlers = append(reqHandlers, authReqHandler)

	for _, handler := range reqHandlers {
		handler.RegisterRoutes(serverMux)
	}

	// middlewares
	loggerMiddleware := middleware.NewLoggerHandler(logger)
	authMiddleware := middleware.NewAuthenticationMiddleware(*logger, processors.AuthProcessor)

	// wrap middlewares around the server
	wrappedServer := loggerMiddleware.Handler(serverMux)
	wrappedServer = authMiddleware.Handler(serverMux)

	return &http.Server{
		Handler:      wrappedServer,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 16 * time.Minute,
		IdleTimeout:  16 * time.Minute,
		Addr:         fmt.Sprintf(":%d", appConfig.Port),
	}
}
