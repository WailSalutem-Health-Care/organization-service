package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

// GETWithOrgHeader makes a GET request with X-Organization-ID header
func (c *HTTPTestClient) GETWithOrgHeader(t *testing.T, path string, orgID string) *http.Response {
	t.Helper()

	req, err := http.NewRequest("GET", c.BaseURL+path, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-Organization-ID", orgID)

	resp, err := c.Client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	return resp
}

// PUTWithOrgHeader makes a PUT request with X-Organization-ID header
func (c *HTTPTestClient) PUTWithOrgHeader(t *testing.T, path string, body interface{}, orgID string) *http.Response {
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
	req.Header.Set("X-Organization-ID", orgID)

	resp, err := c.Client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	return resp
}

// PATCHWithOrgHeader makes a PATCH request with X-Organization-ID header
func (c *HTTPTestClient) PATCHWithOrgHeader(t *testing.T, path string, body interface{}, orgID string) *http.Response {
	t.Helper()

	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, err := http.NewRequest("PATCH", c.BaseURL+path, bytes.NewReader(jsonBody))
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

// DELETEWithOrgHeader makes a DELETE request with X-Organization-ID header
func (c *HTTPTestClient) DELETEWithOrgHeader(t *testing.T, path string, orgID string) *http.Response {
	t.Helper()

	req, err := http.NewRequest("DELETE", c.BaseURL+path, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-Organization-ID", orgID)

	resp, err := c.Client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	return resp
}
