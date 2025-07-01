package handlers

import (
	"log/slog"
	"net/http"
)

type HealthRoutes struct {
	logger slog.Logger
}

func (hr *HealthRoutes) healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		hr.logger.Error("event=healthCheckFailure", "message", "Setting header failed", "err", err.Error())
		return
	}
}

func (hr *HealthRoutes) RegisterRoutes(serveMux *http.ServeMux) {
	serveMux.Handle("GET /health", http.Handler(http.HandlerFunc(hr.healthCheck)))
}

func NewReqHealthHandlers(logger slog.Logger) RequestHandler {
	return &HealthRoutes{
		logger: logger,
	}
}
