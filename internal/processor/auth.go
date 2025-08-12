package processor

import (
	"context"
	"log/slog"
	"net/http"
	"xrf197ilz35aq/internal"
	"xrf197ilz35aq/internal/client"
	"xrf197ilz35aq/internal/model"
)

type verifyTokenResponse struct {
	UserId string `json:"userId"`
}

type AuthProcessor struct {
	apiClient client.ApiClient
}

func (ap *AuthProcessor) GetAuthToken(ctx context.Context, log slog.Logger, authReq model.AuthRequest) (*model.AuthResponse, error) {
	// 1. Validate authentication request
	if err := authReq.Validate(); err != nil {
		return nil, &internal.ExternalError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
	}

	// 2. Make request to create a user
	var response client.ApiClientResponse[model.AuthResponse]
	if err := ap.apiClient.Post(ctx, "/auth/token", authReq, nil, &response, log); err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func (ap *AuthProcessor) ValidateAuthToken(ctx context.Context, log slog.Logger, req model.VerifyRevokeTokenReq) (bool, error) {
	// 1. Validate request
	if req.Token == "" {
		return false, &internal.ExternalError{
			Message: "Invalid token",
			Code:    http.StatusBadRequest,
		}
	}

	// 2. Make request to validate token
	var response client.ApiClientResponse[verifyTokenResponse]
	if err := ap.apiClient.Post(ctx, "/auth/token/verify", req, nil, &response, log); err != nil {
		return false, err
	}
	return true, nil
}

func NewAuthProcessor(apiClient client.ApiClient) *AuthProcessor {
	return &AuthProcessor{apiClient: apiClient}
}
