package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

// APIClient Helper for testing
type APIClient struct {
	BaseURL    string
	AuthToken  string
	HTTPClient *http.Client
}

// NewAPIClient creates a new client
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Request performs an HTTP request
func (c *APIClient) Request(t *testing.T, method, path string, body interface{}) (int, []byte) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal body: %v", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	return resp.StatusCode, respBody
}

// Get performs a GET request
func (c *APIClient) Get(t *testing.T, path string) (int, []byte) {
	return c.Request(t, http.MethodGet, path, nil)
}

// Post performs a POST request
func (c *APIClient) Post(t *testing.T, path string, body interface{}) (int, []byte) {
	return c.Request(t, http.MethodPost, path, body)
}

// Put performs a PUT request
func (c *APIClient) Put(t *testing.T, path string, body interface{}) (int, []byte) {
	return c.Request(t, http.MethodPut, path, body)
}

// Delete performs a DELETE request
func (c *APIClient) Delete(t *testing.T, path string) (int, []byte) {
	return c.Request(t, http.MethodDelete, path, nil)
}

// Common models for tests
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type SetupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
