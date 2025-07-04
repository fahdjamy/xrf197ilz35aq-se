package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type APIError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Err     error  `json:"error"`
}

func (aErr *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", aErr.Code, aErr.Message)
}

type ApiClient struct {
	baseURL        string
	httpClient     *http.Client
	defaultHeaders map[string]string
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
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	// 3. Create the HTTP request with context
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
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
		req.Header.Set("Content-Type", "application/json")
	}

	// 5. Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		responseBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return &APIError{
				Message: "failed to read client response body",
				Code:    500,
				Err:     err,
			}
		}
		return &APIError{
			Code:    resp.StatusCode,
			Message: "client error response",
			Err:     errors.New(string(responseBytes)),
		}
	}

	// 7. Decode the successful response body into the provided struct 'into'
	if into != nil {
		if err := json.NewDecoder(resp.Body).Decode(into); err != nil {
			return fmt.Errorf("failed to decode response body: %w", err)
		}
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
	return c.do(ctx, http.MethodPost, path, body, customHeaders, into)
}

// Put performs a PUT request.
func (c *ApiClient) Put(ctx context.Context, path string, body interface{}, customHeaders map[string]string, into interface{}) error {
	return c.do(ctx, http.MethodPut, path, body, customHeaders, into)
}

// Delete performs a DELETE request.
func (c *ApiClient) Delete(ctx context.Context, path string, customHeaders map[string]string, into interface{}) error {
	return c.do(ctx, http.MethodDelete, path, nil, customHeaders, into)
}

func NewApiClient(baseURL string, options ...Option) *ApiClient {
	apiClient := &ApiClient{
		baseURL:        baseURL,
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
