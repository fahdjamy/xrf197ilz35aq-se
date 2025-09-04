package handlers

import (
	"log/slog"
	"net/http"
	"xrf197ilz35aq/internal/server/api/response"
)

type RequestHandler interface {
	RegisterRoutes(serveMux *http.ServeMux)
}

func handleProcessorResponse[T any](data T, err error, w http.ResponseWriter, logger slog.Logger, code int) {
	if err != nil {
		response.WriteErrorResponse(err, w, logger)
		return
	}

	successResponse := response.DataResponse{
		Code: code,
		Data: data,
	}
	response.WriteResponse(successResponse, w, logger)
}
