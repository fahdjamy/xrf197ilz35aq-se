package processor

import (
	"context"
	"log/slog"
	"net/http"
	"xrf197ilz35aq/internal"
	"xrf197ilz35aq/internal/client"
	"xrf197ilz35aq/internal/model"
)

type AuthAPIClientResponse struct {
	Code int                `json:"code"`
	Data model.AuthResponse `json:"data"`
}

type AuthProcessor struct {
	apiClient client.ApiClient
}

func (ap *AuthProcessor) AuthenticateUser(ctx context.Context, log slog.Logger, authReq model.AuthRequest) (*model.AuthResponse, error) {
	// 1. Validate authentication request
	if err := authReq.Validate(); err != nil {
		return nil, &internal.ExternalError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
	}

	// 2. Make request to authenticate user
	var response AuthAPIClientResponse
	if err := ap.apiClient.Post(ctx, "/api/v1/auth", authReq, nil, response, log); err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func NewAuthProcessor(apiClient client.ApiClient) *AuthProcessor {
	return &AuthProcessor{apiClient: apiClient}
}
