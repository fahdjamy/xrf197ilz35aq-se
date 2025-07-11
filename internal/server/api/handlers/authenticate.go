package handlers

import (
	"log/slog"
	"net/http"
	"time"
	"xrf197ilz35aq/internal"
	"xrf197ilz35aq/internal/model"
	"xrf197ilz35aq/internal/processor"
	"xrf197ilz35aq/internal/server/api/request"
	"xrf197ilz35aq/internal/server/api/response"
)

type AuthHandler struct {
	defaultLogger slog.Logger
	authProcessor processor.AuthProcessor
}

func (auth *AuthHandler) authenticateUser(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logger := internal.LoggerFromContext(r.Context(), auth.defaultLogger)
	defer func() {
		logger.Info("event=authenticateUser, latency", "timeTaken", time.Since(startTime))
	}()

	var req model.AuthRequest

	if err := request.DecodeJSONBody(r, &req); err != nil {
		response.WriteErrorResponse(err, w, *logger)
		return
	}

	authTokenData, err := auth.authProcessor.AuthenticateUser(r.Context(), *logger, req)
	if err != nil {
		response.WriteErrorResponse(err, w, *logger)
		return
	}

	data := response.DataResponse{
		Code: http.StatusCreated,
		Data: authTokenData,
	}
	response.WriteResponse(data, w, *logger)
}

func (auth *AuthHandler) RegisterRoutes(serveMux *http.ServeMux) {
	serveMux.HandleFunc("POST /api/v1/auth/token", auth.authenticateUser)
}

func NewAuthHandler(logger slog.Logger, authProcessor processor.AuthProcessor) *AuthHandler {
	return &AuthHandler{
		defaultLogger: logger,
		authProcessor: authProcessor,
	}
}
