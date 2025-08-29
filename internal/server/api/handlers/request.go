package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"xrf197ilz35aq/internal/server/api/response"
)

type RequestHandler interface {
	RegisterRoutes(serveMux *http.ServeMux)
}

func logLatency(startTime time.Time, event string, logger slog.Logger) {
	msg := fmt.Sprintf("event=%s", event)
	logger.Info(msg, "latency", time.Since(startTime))
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
