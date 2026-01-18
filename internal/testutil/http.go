package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// HTTPTestClient wraps http.Client with test helpers
type HTTPTestClient struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

// NewHTTPTestClient creates a new test HTTP client
func NewHTTPTestClient(baseURL, token string) *HTTPTestClient {
	return &HTTPTestClient{
		BaseURL: baseURL,
		Token:   token,
		Client:  &http.Client{},
	}
}

// POST makes a POST request with JSON body
func (c *HTTPTestClient) POST(t *testing.T, path string, body interface{}) *http.Response {
	t.Helper()

	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	return resp
}

// POSTWithOrgHeader makes a POST request with JSON body and X-Organization-ID header
func (c *HTTPTestClient) POSTWithOrgHeader(t *testing.T, path string, body interface{}, orgID string) *http.Response {
	t.Helper()

	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Organization-ID", orgID)

	resp, err := c.Client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	return resp
}

// GET makes a GET request
func (c *HTTPTestClient) GET(t *testing.T, path string) *http.Response {
	t.Helper()

	req, err := http.NewRequest("GET", c.BaseURL+path, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.Client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	return resp
}

// PUT makes a PUT request with JSON body
func (c *HTTPTestClient) PUT(t *testing.T, path string, body interface{}) *http.Response {
	t.Helper()

	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, err := http.NewRequest("PUT", c.BaseURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	return resp
}

// DELETE makes a DELETE request
func (c *HTTPTestClient) DELETE(t *testing.T, path string) *http.Response {
	t.Helper()

	req, err := http.NewRequest("DELETE", c.BaseURL+path, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.Client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	return resp
}

// DecodeJSON decodes response body into target
func DecodeJSON(t *testing.T, resp *http.Response, target interface{}) {
	t.Helper()
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if err := json.Unmarshal(body, target); err != nil {
		t.Fatalf("Failed to decode response (body: %s): %v", string(body), err)
	}
}

// ReadBody reads and returns the response body as string
func ReadBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	return string(body)
}

// AssertStatusCode asserts the response status code
func AssertStatusCode(t *testing.T, resp *http.Response, expected int) {
	t.Helper()

	if resp.StatusCode != expected {
		body := ReadBody(t, resp)
		t.Errorf("Expected status %d, got %d. Body: %s", expected, resp.StatusCode, body)
	}
}
