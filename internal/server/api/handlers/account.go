package handlers

import (
	"errors"
	"log/slog"
	"net/http"
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

var invalidUserCtxErr = errors.New("invalid user context object in context")

func (ah *accountHandler) createAccount(w http.ResponseWriter, r *http.Request) {
	logger := server.LoggerFromContext(r.Context(), ah.defaultLogger)

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
	logger := server.LoggerFromContext(r.Context(), ah.defaultLogger)
	accountId, isValid := getAndValidateId(r, "accountId")

	if !isValid {
		externalError := internal.ExternalError{Message: "Invalid/missing accountId", Code: http.StatusBadRequest}
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

func (ah *accountHandler) updateAccount(w http.ResponseWriter, r *http.Request) {
	logger := server.LoggerFromContext(r.Context(), ah.defaultLogger)

	var req model.UpdateAccountRequest
	if err := request.DecodeJSONBody(r, &req); err != nil {
		response.WriteErrorResponse(err, w, *logger)
		return
	}

	userCtx, ok := server.UserFromContext(r.Context())
	if !ok || userCtx == nil || userCtx.Fingerprint == "" {
		response.WriteErrorResponse(invalidUserCtxErr, w, *logger)
		return
	}

	accountId, isValid := getAndValidateId(r, "accountId")
	if !isValid {
		externalError := internal.ExternalError{Message: "Invalid/missing accountId", Code: http.StatusBadRequest}
		response.WriteErrorResponse(&externalError, w, *logger)
		return
	}

	accountUpdated, err := ah.processor.UpdateAccount(r.Context(), *userCtx, accountId, req)
	if err == nil && !accountUpdated {
		err = errors.New("account not updated")
	}

	handleProcessorResponse(accountUpdated, err, w, *logger, http.StatusOK)
}

func (ah *accountHandler) lockAccount(w http.ResponseWriter, r *http.Request) {
	logger := server.LoggerFromContext(r.Context(), ah.defaultLogger)

	userCtx, ok := server.UserFromContext(r.Context())
	if !ok || userCtx == nil || userCtx.Fingerprint == "" {
		response.WriteErrorResponse(invalidUserCtxErr, w, *logger)
		return
	}

	accountId, isValid := getAndValidateId(r, "accountId")
	if !isValid {
		externalError := internal.ExternalError{Message: "Invalid/missing accountId", Code: http.StatusBadRequest}
		response.WriteErrorResponse(&externalError, w, *logger)
		return
	}

	locked, err := ah.processor.LockAccount(r.Context(), *userCtx, accountId)

	if err == nil && !locked {
		err = errors.New("account not locked")
	}

	handleProcessorResponse(locked, err, w, *logger, http.StatusOK)
}

func (ah *accountHandler) unlockAccount(w http.ResponseWriter, r *http.Request) {
	logger := server.LoggerFromContext(r.Context(), ah.defaultLogger)

	userCtx, ok := server.UserFromContext(r.Context())
	if !ok || userCtx == nil || userCtx.Fingerprint == "" {
		response.WriteErrorResponse(invalidUserCtxErr, w, *logger)
		return
	}

	accountId, isValid := getAndValidateId(r, "accountId")
	if !isValid {
		externalError := internal.ExternalError{Message: "Invalid/missing accountId", Code: http.StatusBadRequest}
		response.WriteErrorResponse(&externalError, w, *logger)
		return
	}

	unlocked, err := ah.processor.LockAccount(r.Context(), *userCtx, accountId)

	if err == nil && !unlocked {
		err = errors.New("account has not been unlocked")
	}

	handleProcessorResponse(unlocked, err, w, *logger, http.StatusOK)
}

func (ah *accountHandler) getAccounts(w http.ResponseWriter, r *http.Request) {
	logger := server.LoggerFromContext(r.Context(), ah.defaultLogger)

	var req model.FindAccountRequest
	if err := request.DecodeJSONBody(r, &req); err != nil {
		response.WriteErrorResponse(err, w, *logger)
		return
	}

	userCtx, ok := server.UserFromContext(r.Context())
	if !ok || userCtx == nil || userCtx.Fingerprint == "" {
		response.WriteErrorResponse(invalidUserCtxErr, w, *logger)
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
	serveMux.HandleFunc("PATCH /api/v1/accounts/{accountId}/lock", ah.lockAccount)
	serveMux.HandleFunc("PATCH /api/v1/accounts/{accountId}/unlock", ah.unlockAccount)
}

func NewAccountHandler(defaultLogger slog.Logger, processor processor.AccountProcessor) RequestHandler {
	return &accountHandler{defaultLogger, processor}
}
