package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHealth(t *testing.T) {
	// Initialize the server (Engine can be nil for this test)
	s := &Server{Engine: nil}
	handler := NewHandler(s.Engine)

	// Create a request
	req, _ := http.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check response body
	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "ok", resp["status"])
}

func TestGetInfo(t *testing.T) {
	// Initialize the server
	s := &Server{Engine: nil}
	handler := NewHandler(s.Engine)

	// Create a request
	req, _ := http.NewRequest("GET", "/info", nil)
	rr := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check response body
	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, "trellis-http", resp["app"])
	assert.NotEmpty(t, resp["version"])
	assert.Equal(t, "0.1.0", resp["api_version"])
}
