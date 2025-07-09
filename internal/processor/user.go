package processor

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"xrf197ilz35aq/internal"
	"xrf197ilz35aq/internal/client"
	"xrf197ilz35aq/internal/model"
)

type UserClientResponse struct {
	Code int                `json:"code"`
	Data model.UserResponse `json:"data"`
}

type UserProcessor struct {
	apiClient client.ApiClient
}

func (up *UserProcessor) CreateUser(ctx context.Context, log slog.Logger, userReq *model.UserRequest) (*model.UserResponse, error) {
	// 1. Validate user request
	if err := userReq.Validate(); err != nil {
		return nil, &internal.ExternalError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
	}

	// 2. Make request to create user
	var clientResponse UserClientResponse
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := up.apiClient.Post(timeoutCtx, "/user", userReq, nil, &clientResponse, log)
	if err != nil {
		return nil, err
	}

	return &clientResponse.Data, nil
}

func (up *UserProcessor) GetUserProfile(ctx context.Context, userId string, log slog.Logger) (*model.UserResponse, error) {
	var userResponse UserClientResponse

	if err := up.apiClient.Get(ctx, fmt.Sprintf("/user/%s", userId), nil, userResponse, log); err != nil {
		return nil, err
	}

	return &userResponse.Data, nil
}

func NewUserProcessor(apiClient client.ApiClient) *UserProcessor {
	return &UserProcessor{
		apiClient: apiClient,
	}
}
