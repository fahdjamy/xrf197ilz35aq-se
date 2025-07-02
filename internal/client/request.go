package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type ApiClient struct {
	baseURL        string
	httpClient     *http.Client
	logger         slog.Logger
	defaultHeaders map[string]string
}

func (c *ApiClient) do(ctx context.Context, method, path string, body interface{}, customHeaders map[string]string, responseStruct interface{}) error {
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
			return fmt.Errorf("failed to read response body: %w", err)
		}
		return fmt.Errorf("client error: %d %s", resp.StatusCode, string(responseBytes))
	}

	// 7. Decode the successful response body into the provided struct 'v'
	if responseStruct != nil {
		if err := json.NewDecoder(resp.Body).Decode(responseStruct); err != nil {
			return fmt.Errorf("failed to decode response body: %w", err)
		}
	}
	return nil
}
