package handlers

import (
	"log/slog"
	"net/http"
	"xrf197ilz35aq/internal/server/api/response"
)

type healthRoutes struct {
	logger slog.Logger
}

func (hr *healthRoutes) healthCheck(w http.ResponseWriter, _ *http.Request) {
	data := response.DataResponse{
		Code: http.StatusOK,
		Data: struct {
			Health bool `json:"health"`
		}{
			Health: true,
		},
	}
	response.WriteResponse(data, w, hr.logger)
}

func (hr *healthRoutes) RegisterRoutes(serveMux *http.ServeMux) {
	serveMux.HandleFunc("GET /health", hr.healthCheck)
}

func NewReqHealthHandlers(logger slog.Logger) RequestHandler {
	return &healthRoutes{
		logger: logger,
	}
}
