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

type ApiClientResponse[T any] struct {
	Code int `json:"code"`
	Data T   `json:"data"`
}

type ApiClient struct {
	baseURL        string
	appId          string
	httpClient     *http.Client
	defaultHeaders map[string]string
	appConfig      internal.AppConfig
}

type Option func(*ApiClient)

func (c *ApiClient) do(ctx context.Context, method, path string, body interface{}, customHeaders map[string]string, into interface{}, log slog.Logger) error {
	// 1. Create the full URL path.
	fullURL := c.baseURL + path

	// 2. Marshal the request body into JSON, if it exists
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			log.Error("failed to marshal request body", "error", err)
			return &internal.ServerError{Err: fmt.Errorf("failed to marshal request body: %w", err)}
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	// 3. Create the HTTP request with context
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		log.Error("failed to create request", "error", err)
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
	// Always, set the service/app id header
	req.Header.Set(internal.XrfHeaderAppId, c.appId)

	// 5. Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error("failed to execute request", "error", err)
		return &internal.ServerError{
			Err: fmt.Errorf("failed to execute request: %w", err),
		}
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode < 200 || statusCode >= 400 {
		var apiClientError internal.APIClientError
		if err := parseClientResponse(resp.Body, &apiClientError, log); err != nil {
			log.Error("failed to parse client response body", "error", err)
			return err
		}
		return &apiClientError
	}

	// 7. Decode the successful response body into the provided struct 'into'
	if into != nil {
		if err := parseClientResponse(resp.Body, into, log); err != nil {
			log.Error("failed to parse client response body", "error", err)
			return err
		}
		log.Info("API client returned successfully", "code", statusCode, "body", into)
	}
	return nil
}

// Get performs a GET request.
// - Param -> 'path' is the endpoint path (e.g., "/users/123").
// - Param -> 'customHeaders' allows for adding request-specific headers.
// - Param -> 'into' is the struct to decode the JSON response into.
func (c *ApiClient) Get(ctx context.Context, path string, customHeaders map[string]string, into interface{}, log slog.Logger) error {
	return c.do(ctx, http.MethodGet, path, nil, customHeaders, into, log)
}

// Post performs a POST request.
// 'Body' is the request payload, which will be marshaled to JSON.
func (c *ApiClient) Post(ctx context.Context, path string, body interface{}, customHeaders map[string]string, into interface{}, log slog.Logger) error {
	extraHeaders := make(map[string]string)
	for k, v := range customHeaders {
		extraHeaders[k] = v
	}
	extraHeaders["Accept"] = "application/json"
	return c.do(ctx, http.MethodPost, path, body, extraHeaders, into, log)
}

// Put performs a PUT request.
func (c *ApiClient) Put(ctx context.Context, path string, body interface{}, customHeaders map[string]string, into interface{}, log slog.Logger) error {
	return c.do(ctx, http.MethodPut, path, body, customHeaders, into, log)
}

// Delete performs a DELETE request.
func (c *ApiClient) Delete(ctx context.Context, path string, customHeaders map[string]string, into interface{}, log slog.Logger) error {
	return c.do(ctx, http.MethodDelete, path, nil, customHeaders, into, log)
}

func parseClientResponse(body io.Reader, into interface{}, log slog.Logger) error {
	responseBytes, err := io.ReadAll(body)
	if err != nil {
		log.Error("failed to read response body", "error", err)
		return &internal.ServerError{Err: fmt.Errorf("error reading client respnse body: %w", err)}
	}
	if err := json.NewDecoder(bytes.NewReader(responseBytes)).Decode(into); err != nil {
		log.Error("failed to parse client response body", "error", err)
		return &internal.ServerError{Err: fmt.Errorf("error UnMarshalling/decoding client response body: %w", err)}
	}

	return nil
}

func NewApiClient(baseURL string, appConfig internal.AppConfig, options ...Option) *ApiClient {
	apiClient := &ApiClient{
		baseURL:        baseURL,
		appConfig:      appConfig,
		appId:          getAppId(),
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

func getAppId() string {
	return "xrf-aq-SE"
}
