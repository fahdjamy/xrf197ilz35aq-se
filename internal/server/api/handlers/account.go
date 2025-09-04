package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"time"
	"xrf197ilz35aq/internal"
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
	if !ok || userCtx == nil || userCtx.Fingerprint == "" {
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

func (ah *accountHandler) getAccountById(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logger := server.LoggerFromContext(r.Context(), ah.defaultLogger)
	defer logLatency(startTime, "getAccountById", *logger)

	accountId, isValid := getAndValidateId(r, "accountId")
	if !isValid {
		externalError := internal.ExternalError{Message: "Account not found", Code: http.StatusBadRequest}
		response.WriteErrorResponse(&externalError, w, *logger)
		return
	}

	userCtx, ok := server.UserFromContext(r.Context())
	if !ok || userCtx == nil || userCtx.Fingerprint == "" {
		response.WriteErrorResponse(errors.New("invalid user context object in context"), w, *logger)
		return
	}

	//// Call processor
	savedAccount, err := ah.processor.FindAccountByID(r.Context(), *userCtx, accountId)

	handleProcessorResponse(savedAccount, err, w, *logger, http.StatusOK)
}

func (ah *accountHandler) updateAccount(w http.ResponseWriter, r *http.Request) {}

func (ah *accountHandler) lockAccount(w http.ResponseWriter, r *http.Request) {}

func (ah *accountHandler) unlockAccount(w http.ResponseWriter, r *http.Request) {}

func (ah *accountHandler) getAccounts(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logger := server.LoggerFromContext(r.Context(), ah.defaultLogger)
	defer logLatency(startTime, "getAccounts", *logger)

	var req model.FindAccountRequest
	if err := request.DecodeJSONBody(r, &req); err != nil {
		response.WriteErrorResponse(err, w, *logger)
		return
	}

	userCtx, ok := server.UserFromContext(r.Context())
	if !ok || userCtx == nil || userCtx.Fingerprint == "" {
		logger.Error("invalid user context object in context", "userCtx", userCtx)
		response.WriteErrorResponse(errors.New("invalid user context object in context"), w, *logger)
		return
	}

	userAccounts, err := ah.processor.FindAccounts(r.Context(), *userCtx, req)
	if err != nil {
		response.WriteErrorResponse(err, w, *logger)
		return
	}

	data := response.DataResponse{
		Code: http.StatusOK,
		Data: userAccounts,
	}
	response.WriteResponse(data, w, *logger)
}

func (ah *accountHandler) RegisterRoutes(serveMux *http.ServeMux) {
	serveMux.HandleFunc("POST /api/v1/accounts", ah.getAccounts)
	serveMux.HandleFunc("POST /api/v1/account", ah.createAccount)
	serveMux.HandleFunc("PUT /api/v1/accounts/{accountId}", ah.updateAccount)
	serveMux.HandleFunc("GET /api/v1/accounts/{accountId}", ah.getAccountById)
	serveMux.HandleFunc("POST /api/v1/accounts/{accountId}/lock", ah.lockAccount)
	serveMux.HandleFunc("POST /api/v1/accounts/{accountId}/unlock", ah.unlockAccount)
}

func NewAccountHandler(defaultLogger slog.Logger, processor processor.AccountProcessor) RequestHandler {
	return &accountHandler{defaultLogger, processor}
}
