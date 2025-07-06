package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"xrf197ilz35aq/internal"
	"xrf197ilz35aq/internal/client"
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
	w.Header().Set(internal.ContentType, internal.ApplicationJson+"; charset=utf-8")
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

func WriteErrorResponse(errObj error, w http.ResponseWriter, logger slog.Logger) {
	msg := "Something went wrong"
	statusCode := http.StatusInternalServerError

	var externalError *internal.ExternalError
	var apiClientError *client.APIError

	switch {
	case errors.As(errObj, &externalError):
		if externalError.Code >= 400 {
			statusCode = externalError.Code
		} else {
			statusCode = http.StatusInternalServerError
		}
		msg = externalError.Message
	case errors.As(errObj, &apiClientError):
		if apiClientError.Message != "" {
			msg = apiClientError.Message
		} else {
			msg = "internal server error"
		}
		statusCode = apiClientError.Code
	default:
		// default values are set while setting the variables
	}

	w.Header().Set(internal.ContentType, internal.ApplicationJson)
	w.WriteHeader(statusCode)

	errResp := errorResponse{Error: msg, Code: statusCode}

	err := json.NewEncoder(w).Encode(errResp)
	if err != nil {
		logger.Error(fmt.Sprintf("error writing error response: %s", err))
	}
}
