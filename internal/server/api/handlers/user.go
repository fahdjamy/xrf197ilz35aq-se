package handlers

import (
	"log/slog"
	"net/http"
	"xrf197ilz35aq/internal/server/api/request"
	"xrf197ilz35aq/internal/server/api/response"
)

type userHandler struct {
	logger slog.Logger
}

func (user *userHandler) createUser(w http.ResponseWriter, r *http.Request) {
	var userReq request.UserRequest

	err := request.DecodeJSONBody(r, &userReq)
	if err != nil {
		response.WriteErrorResponse(err, w, user.logger)
		return
	}

	data := response.DataResponse{
		Code: http.StatusCreated,
		Data: nil,
	}
	response.WriteResponse(data, w, user.logger)
}

func (user *userHandler) RegisterRoutes(serveMux *http.ServeMux) {
	serveMux.HandleFunc("POST /user/", user.createUser)
}

func CreateUserHandler(logger slog.Logger) RequestHandler {
	return &userHandler{logger: logger}
}
