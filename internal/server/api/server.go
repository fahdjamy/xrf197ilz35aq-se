package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"xrf197ilz35aq/internal"
	"xrf197ilz35aq/internal/server/api/handlers"
)

func CreateServer(logger slog.Logger, appConfig internal.AppConfig) *http.Server {
	serverMux := http.NewServeMux()

	reqHandlers := make([]handlers.RequestHandler, 0)

	// create request (routes) handlers
	healthReqHandler := handlers.NewReqHealthHandlers(logger)
	userReqHandler := handlers.NewUserReqHandler(logger)

	reqHandlers = append(reqHandlers, healthReqHandler)
	reqHandlers = append(reqHandlers, userReqHandler)

	for _, handler := range reqHandlers {
		handler.RegisterRoutes(serverMux)
	}

	return &http.Server{
		Handler:      serverMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 16 * time.Minute,
		IdleTimeout:  16 * time.Minute,
		Addr:         fmt.Sprintf(":%d", appConfig.Port),
	}
}
