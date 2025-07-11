package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"
	"xrf197ilz35aq/internal"
	"xrf197ilz35aq/internal/model"
	"xrf197ilz35aq/internal/processor"
	"xrf197ilz35aq/internal/server/api/request"
	"xrf197ilz35aq/internal/server/api/response"
)

type userHandler struct {
	defaultLogger slog.Logger
	processor     processor.UserProcessor
}

func (uh *userHandler) createUser(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logger := internal.LoggerFromContext(r.Context(), uh.defaultLogger)
	defer func() {
		logger.Info("createUser latency", slog.String("method", "createUser"), "timeTaken", time.Since(startTime))
	}()
	var userReq model.UserRequest

	err := request.DecodeJSONBody(r, &userReq)
	if err != nil {
		response.WriteErrorResponse(err, w, *logger)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	//// Call processor
	createdUser, err := uh.processor.CreateUser(ctx, *logger, &userReq)
	if err != nil {
		response.WriteErrorResponse(err, w, *logger)
		return
	}

	data := response.DataResponse{
		Code: http.StatusCreated,
		Data: createdUser,
	}
	response.WriteResponse(data, w, *logger)
}

func (uh *userHandler) getUser(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logger := internal.LoggerFromContext(r.Context(), uh.defaultLogger)
	defer func() {
		logger.Info("getUser latency", slog.String("method", "getUser"), "timeTaken", time.Since(startTime))
	}()

	userId, isValid := getAndValidateId(r, "userId")
	if !isValid {
		externalError := internal.ExternalError{Message: "invalid user id", Code: http.StatusBadRequest}
		response.WriteErrorResponse(&externalError, w, *logger)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	//// Call processor
	createdUser, err := uh.processor.GetUserProfile(ctx, userId, *logger)
	if err != nil {
		response.WriteErrorResponse(err, w, *logger)
		return
	}

	data := response.DataResponse{Code: http.StatusCreated, Data: createdUser}
	response.WriteResponse(data, w, *logger)
}

func (uh *userHandler) RegisterRoutes(serveMux *http.ServeMux) {
	serveMux.HandleFunc("POST /api/v1/user", uh.createUser)
	serveMux.HandleFunc("GET /api/v1/user/{userId}", uh.getUser)
}

func NewUserReqHandler(logger slog.Logger, userProcessor processor.UserProcessor) RequestHandler {
	return &userHandler{defaultLogger: logger, processor: userProcessor}
}

func getAndValidateId(req *http.Request, reqIdKey string) (string, bool) {
	value := req.PathValue(reqIdKey)
	if value == "" {
		return "", false
	}
	return value, true
}
