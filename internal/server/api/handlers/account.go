package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"time"
	"xrf197ilz35aq/internal/model"
	"xrf197ilz35aq/internal/processor"
	"xrf197ilz35aq/internal/server"
	"xrf197ilz35aq/internal/server/api/request"
	"xrf197ilz35aq/internal/server/api/response"
)

type accountHandler struct {
	defaultLogger slog.Logger
	processor     processor.AccountProcessor
}

func (ah *accountHandler) createAccount(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logger := server.LoggerFromContext(r.Context(), ah.defaultLogger)
	defer logLatency(startTime, "createAccount", *logger)

	var req model.AccountRequest
	if err := request.DecodeJSONBody(r, &req); err != nil {
		response.WriteErrorResponse(err, w, *logger)
		return
	}

	userCtx, ok := server.UserFromContext(r.Context())
	if !ok || userCtx == nil || userCtx.UserId == "" {
		response.WriteErrorResponse(errors.New("invalid user context object in context"), w, *logger)
		return
	}

	//// Call processor
	savedAccount, err := ah.processor.CreateAccount(r.Context(), *userCtx, req)
	if err != nil {
		response.WriteErrorResponse(err, w, *logger)
		return
	}

	data := response.DataResponse{
		Code: http.StatusCreated,
		Data: savedAccount,
	}
	response.WriteResponse(data, w, *logger)
}

func (ah *accountHandler) getAccountById(w http.ResponseWriter, r *http.Request) {}

func (ah *accountHandler) updateAccount(w http.ResponseWriter, r *http.Request) {}

func (ah *accountHandler) lockAccount(w http.ResponseWriter, r *http.Request) {}

func (ah *accountHandler) unlockAccount(w http.ResponseWriter, r *http.Request) {}

func (ah *accountHandler) getAccounts(w http.ResponseWriter, r *http.Request) {}

func (ah *accountHandler) RegisterRoutes(serveMux *http.ServeMux) {
	serveMux.HandleFunc("POST /api/v1/accounts", ah.getAccounts)
	serveMux.HandleFunc("POST /api/v1/account", ah.createAccount)
	serveMux.HandleFunc("PUT /api/v1/account/{account}", ah.updateAccount)
	serveMux.HandleFunc("GET /api/v1/account/{account}", ah.getAccountById)
	serveMux.HandleFunc("POST /api/v1/account/{account}/lock", ah.lockAccount)
	serveMux.HandleFunc("POST /api/v1/account/{account}/unlock", ah.unlockAccount)
}

func NewAccountHandler(defaultLogger slog.Logger, processor processor.AccountProcessor) RequestHandler {
	return &accountHandler{defaultLogger, processor}
}
