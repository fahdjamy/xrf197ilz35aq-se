package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
	"xrf197ilz35aq/internal"
)

type ApiClient struct {
	baseURL        string
	logger         slog.Logger
	httpClient     *http.Client
	defaultHeaders map[string]string
	appConfig      internal.AppConfig
}

type Option func(*ApiClient)

func (c *ApiClient) do(ctx context.Context, method, path string, body interface{}, customHeaders map[string]string, into interface{}) error {
	// 1. Create the full URL path.
	fullURL := c.baseURL + path

	// 2. Marshal the request body into JSON, if it exists
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			c.logger.Error("failed to marshal request body", "error", err)
			return &internal.ServerError{Err: fmt.Errorf("failed to marshal request body: %w", err)}
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	// 3. Create the HTTP request with context
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		c.logger.Error("failed to create request", "error", err)
		return &internal.ServerError{
			Err: fmt.Errorf("failed to create request: %w", err),
		}
	}

	// 4. Set headers
	// Start with default headers
	for k, v := range c.defaultHeaders {
		req.Header.Set(k, v)
	}
	// Add/overwrite with custom headers is specified
	if customHeaders != nil {
		for k, v := range customHeaders {
			req.Header.Set(k, v)
		}
	}
	// Always set Content-Type to application/json body is not nil
	if body != nil {
		req.Header.Set(internal.ContentType, internal.ApplicationJson)
	}
	// Always, set service to service token
	req.Header.Set(internal.SrvToSrvToken, generateSrvToSrvToken())

	// 5. Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("failed to execute request", "error", err)
		return &internal.ServerError{
			Err: fmt.Errorf("failed to execute request: %w", err),
		}
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode < 200 || statusCode >= 400 {
		responseBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			errMsg := "failed to read client response body"
			c.logger.Error("failed to read client response body", "error", errMsg)
			return &internal.ServerError{Err: fmt.Errorf("failed to read client response body: %w", err)}
		}
		// decode api client error
		var apiClientError internal.APIClientError
		if err := json.NewDecoder(bytes.NewReader(responseBytes)).Decode(&apiClientError); err != nil {
			errMsg := "failed to decode client response body"
			c.logger.Error(errMsg, "error", errMsg, "err", err)
			return &internal.ServerError{Err: fmt.Errorf("failed to decode client response body: %w", err)}
		}
		c.logger.Error("API client returned error", "code", statusCode, "error", string(responseBytes))
		return &apiClientError
	}

	// 7. Decode the successful response body into the provided struct 'into'
	if into != nil {
		responseBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			errMsg := "failed to read client response body"
			c.logger.Error("failed to read client response body", "error", errMsg)
			return &internal.ServerError{Err: fmt.Errorf("failed to read client response body: %w", err)}
		}
		if err := json.NewDecoder(bytes.NewReader(responseBytes)).Decode(into); err != nil {
			errMsg := "failed to decode client response body"
			c.logger.Error(errMsg, "error", errMsg)
			return &internal.ServerError{Err: fmt.Errorf("failed to decode client response body: %w", err)}
		}
		c.logger.Info("API client returned successfully", "code", statusCode, "body", into)
	}
	return nil
}

// Get performs a GET request.
// - Param -> 'path' is the endpoint path (e.g., "/users/123").
// - Param -> 'customHeaders' allows for adding request-specific headers.
// - Param -> 'into' is the struct to decode the JSON response into.
func (c *ApiClient) Get(ctx context.Context, path string, customHeaders map[string]string, into interface{}) error {
	return c.do(ctx, http.MethodGet, path, nil, customHeaders, into)
}

// Post performs a POST request.
// 'Body' is the request payload, which will be marshaled to JSON.
func (c *ApiClient) Post(ctx context.Context, path string, body interface{}, customHeaders map[string]string, into interface{}) error {
	extraHeaders := make(map[string]string)
	for k, v := range customHeaders {
		extraHeaders[k] = v
	}
	extraHeaders["Accept"] = "application/json"
	return c.do(ctx, http.MethodPost, path, body, extraHeaders, into)
}

// Put performs a PUT request.
func (c *ApiClient) Put(ctx context.Context, path string, body interface{}, customHeaders map[string]string, into interface{}) error {
	return c.do(ctx, http.MethodPut, path, body, customHeaders, into)
}

// Delete performs a DELETE request.
func (c *ApiClient) Delete(ctx context.Context, path string, customHeaders map[string]string, into interface{}) error {
	return c.do(ctx, http.MethodDelete, path, nil, customHeaders, into)
}

func NewApiClient(baseURL string, logger slog.Logger, appConfig internal.AppConfig, options ...Option) *ApiClient {
	apiClient := &ApiClient{
		logger:         logger,
		baseURL:        baseURL,
		appConfig:      appConfig,
		httpClient:     http.DefaultClient,
		defaultHeaders: make(map[string]string),
	}

	// apply all the options
	for _, option := range options {
		option(apiClient)
	}
	return apiClient
}

func WithTimeout(timeout time.Duration) Option {
	return func(apiClient *ApiClient) {
		apiClient.httpClient.Timeout = timeout
	}
}

func WithDefaultHeader(defaultHeaders map[string]string) Option {
	return func(apiClient *ApiClient) {
		if apiClient.defaultHeaders == nil {
			for k, v := range defaultHeaders {
				if _, ok := apiClient.defaultHeaders[k]; !ok {
					apiClient.defaultHeaders[k] = v
				}
			}
		}
	}
}

func generateSrvToSrvToken() string {
	return "srv-to-srv-token/test"
}
