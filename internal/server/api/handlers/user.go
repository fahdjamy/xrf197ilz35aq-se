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

type userHandler struct {
	defaultLogger slog.Logger
	processor     processor.UserProcessor
}

func (uh *userHandler) createUser(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logger := internal.LoggerFromContext(r.Context(), uh.defaultLogger)
	defer logLatency(startTime, "createUser", *logger)
	var userReq model.UserRequest

	err := request.DecodeJSONBody(r, &userReq)
	if err != nil {
		response.WriteErrorResponse(err, w, *logger)
		return
	}

	//// Call processor
	createdUser, err := uh.processor.CreateUser(r.Context(), *logger, &userReq)
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
	defer logLatency(startTime, "getUser", *logger)

	userId, isValid := getAndValidateId(r, "userId")
	if !isValid {
		externalError := internal.ExternalError{Message: "invalid user id", Code: http.StatusBadRequest}
		response.WriteErrorResponse(&externalError, w, *logger)
		return
	}

	authToken := r.Header.Get(internal.XrfAuthToken)

	//// Call processor
	createdUser, err := uh.processor.GetUserProfile(r.Context(), *logger, userId, authToken)
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
