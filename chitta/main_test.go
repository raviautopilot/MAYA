package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"chitta/handlers"
	"chitta/models"
)

func TestParseAllowedOrigins(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string returns default",
			input:    "",
			expected: []string{"http://localhost:3000"},
		},
		{
			name:     "whitespace only returns default",
			input:    "   ",
			expected: []string{"http://localhost:3000"},
		},
		{
			name:     "single origin",
			input:    "http://localhost:3000",
			expected: []string{"http://localhost:3000"},
		},
		{
			name:     "multiple origins comma-separated",
			input:    "http://localhost:3000,http://localhost:3001,https://mykanban.example.com",
			expected: []string{"http://localhost:3000", "http://localhost:3001", "https://mykanban.example.com"},
		},
		{
			name:     "origins with extra whitespace",
			input:    " http://localhost:3000 , http://localhost:3001 ",
			expected: []string{"http://localhost:3000", "http://localhost:3001"},
		},
		{
			name:     "trailing comma ignored",
			input:    "http://localhost:3000,",
			expected: []string{"http://localhost:3000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAllowedOrigins(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d origins, got %d: %v", len(tt.expected), len(result), result)
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("origin[%d]: expected %q, got %q", i, tt.expected[i], v)
				}
			}
		})
	}
}

func TestCORSPreflightResponse(t *testing.T) {
	// Minimal config for SetupRouter — only fields used by the router
	cfg := &models.Config{
		JWTSecret:      "test-secret",
		AllowedOrigins: "http://localhost:3000,http://localhost:4000",
	}
	// Use a nil handler since we only test CORS middleware (OPTIONS never reaches handlers)
	router := SetupRouter(&handlers.Handler{}, cfg)

	// Send a preflight OPTIONS request
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type,Authorization")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("expected Access-Control-Allow-Origin 'http://localhost:3000', got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("expected Access-Control-Allow-Credentials 'true', got %q", w.Header().Get("Access-Control-Allow-Credentials"))
	}
	// Preflight should return 204 or 200
	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Errorf("expected status 200 or 204 for preflight, got %d", w.Code)
	}
}

func TestCORSRejectsUnknownOrigin(t *testing.T) {
	cfg := &models.Config{
		JWTSecret:      "test-secret",
		AllowedOrigins: "http://localhost:3000",
	}
	router := SetupRouter(&handlers.Handler{}, cfg)

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/login", nil)
	req.Header.Set("Origin", "http://evil.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// The middleware should NOT set Access-Control-Allow-Origin for unknown origins
	if origin := w.Header().Get("Access-Control-Allow-Origin"); origin == "http://evil.example.com" {
		t.Errorf("CORS should not allow unknown origin, but got %q", origin)
	}
}
