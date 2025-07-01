package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"xrf197ilz35aq/internal"
)

type DataResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

type pagination struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Total  int `json:"total"`
	Start  int `json:"start"`
}

func WriteResponse(data DataResponse, w http.ResponseWriter, logger slog.Logger) {
	WritePaginatedResponse(data, nil, w, logger)
}

func WritePaginatedResponse(data DataResponse, pag *pagination, w http.ResponseWriter, logger slog.Logger) {
	w.Header().Set(internal.ContentType, internal.ApplicationJson)
	w.WriteHeader(data.Code)

	if pag == nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			logger.Error("event=writeResponseFailure", "error", err)
		}
	} else {
		err := json.NewEncoder(w).Encode(struct {
			*pagination
			DataResponse
		}{})
		if err != nil {
			logger.Error("event=writePaginatedResponseFailure", "error", err)
		}
	}
}

func WriteErrorResponse(_ error, w http.ResponseWriter, logger slog.Logger) {
	msg := "Something went wrong"
	statusCode := http.StatusInternalServerError

	w.Header().Set(internal.ContentType, internal.ApplicationJson)
	w.WriteHeader(statusCode)

	errResp := errorResponse{Error: msg, Code: statusCode}

	err := json.NewEncoder(w).Encode(errResp)
	if err != nil {
		logger.Error("error writing error response", "err", err)
	}
}
