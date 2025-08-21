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
	"xrf197ilz35aq/internal/server"
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

func (up *UserProcessor) GetUserProfile(ctx context.Context, log slog.Logger, userId, authToken string) (*model.UserResponse, error) {
	var userResponse UserClientResponse
	path := fmt.Sprintf("/user/%s", userId)

	if err := up.apiClient.Get(ctx, path, server.CreateAuthTokenHeader(authToken), &userResponse, log); err != nil {
		return nil, err
	}

	if userResponse.Data.UserId != userId {
		log.Warn("user profile not found", "returnedResponse", userResponse.Data)
		return nil, &internal.ExternalError{
			Code:    http.StatusBadRequest,
			Message: "user profile not found",
		}
	}

	return &userResponse.Data, nil
}

func NewUserProcessor(apiClient client.ApiClient) *UserProcessor {
	return &UserProcessor{
		apiClient: apiClient,
	}
}
