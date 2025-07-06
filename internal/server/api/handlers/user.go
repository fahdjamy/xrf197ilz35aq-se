package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"
	"xrf197ilz35aq/internal/model"
	"xrf197ilz35aq/internal/processor"
	"xrf197ilz35aq/internal/server/api/request"
	"xrf197ilz35aq/internal/server/api/response"
)

type userHandler struct {
	logger    slog.Logger
	processor processor.UserProcessor
}

func (uh *userHandler) createUser(w http.ResponseWriter, r *http.Request) {
	var userReq model.UserRequest

	err := request.DecodeJSONBody(r, &userReq)
	if err != nil {
		response.WriteErrorResponse(err, w, uh.logger)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	createdUser, err := uh.processor.CreateUser(ctx, userReq)
	if err != nil {
		response.WriteErrorResponse(err, w, uh.logger)
		return
	}

	//// Call processor

	data := response.DataResponse{
		Code: http.StatusCreated,
		Data: createdUser,
	}
	response.WriteResponse(data, w, uh.logger)
}

func (uh *userHandler) RegisterRoutes(serveMux *http.ServeMux) {
	serveMux.HandleFunc("POST /user", uh.createUser)
}

func NewUserReqHandler(logger slog.Logger) RequestHandler {
	return &userHandler{logger: logger}
}
