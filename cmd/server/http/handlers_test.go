package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"go-form-app/internal/scripts"
)

func TestNewHandlers(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	handlers := NewHandlers(logger)

	if handlers == nil {
		t.Error("NewHandlers() returned nil")
	}
	if handlers.logger != logger {
		t.Error("NewHandlers() logger not set correctly")
	}
	if handlers.executor == nil {
		t.Error("NewHandlers() executor not initialized")
	}
	if len(handlers.security.AllowedScripts) == 0 {
		t.Error("NewHandlers() no allowed scripts configured")
	}
}

func TestFormHandler(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	handlers := NewHandlers(logger)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectBody     bool
	}{
		{
			name:           "GET request should return form (may fail due to missing template)",
			method:         http.MethodGet,
			expectedStatus: http.StatusInternalServerError, // Template not found in test environment
			expectBody:     false,
		},
		{
			name:           "POST request should be rejected",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			expectBody:     false,
		},
		{
			name:           "PUT request should be rejected",
			method:         http.MethodPut,
			expectedStatus: http.StatusMethodNotAllowed,
			expectBody:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()

			handlers.FormHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("FormHandler() status = %d, want %d", w.Code, tt.expectedStatus)
			}

			if tt.expectBody && w.Body.Len() == 0 {
				t.Error("FormHandler() expected body but got empty response")
			}
		})
	}
}

func TestValidateUserID(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	handlers := NewHandlers(logger)

	tests := []struct {
		name     string
		userID   string
		expected bool
	}{
		{"valid user ID - 7 chars", "abc1234", true},
		{"valid user ID - 8 chars", "abc12345", true},
		{"valid user ID - 12 chars", "abc123456789", true},
		{"invalid user ID - too short", "abc123", false},
		{"invalid user ID - too long", "abc1234567890", false},
		{"invalid user ID - special chars", "abc@123", false},
		{"invalid user ID - spaces", "abc 123", false},
		{"invalid user ID - empty", "", false},
		{"valid user ID - all numbers", "1234567", true},
		{"valid user ID - all letters", "abcdefg", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handlers.validateUserID(tt.userID)
			if result != tt.expected {
				t.Errorf("validateUserID(%s) = %v, want %v", tt.userID, result, tt.expected)
			}
		})
	}
}

func TestValidateScript(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	handlers := NewHandlers(logger)

	tests := []struct {
		name     string
		script   string
		expected bool
	}{
		{"valid Python script", "script1.py", true},
		{"valid Python script 2", "script2.py", true},
		{"valid Bash script", "script1.sh", true},
		{"valid Zsh script", "script1.zsh", true},
		{"invalid script - not in whitelist", "malicious.py", false},
		{"invalid script - path traversal", "../script1.py", false},
		{"invalid script - absolute path", "/etc/passwd", false},
		{"invalid script - backslash", "script\\evil.py", false},
		{"invalid script - empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handlers.validateScript(tt.script)
			if result != tt.expected {
				t.Errorf("validateScript(%s) = %v, want %v", tt.script, result, tt.expected)
			}
		})
	}
}

func TestRunScriptHandler_MethodValidation(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	handlers := NewHandlers(logger)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{"GET request should be rejected", http.MethodGet, http.StatusMethodNotAllowed},
		{"PUT request should be rejected", http.MethodPut, http.StatusMethodNotAllowed},
		{"DELETE request should be rejected", http.MethodDelete, http.StatusMethodNotAllowed},
		{"POST request should proceed", http.MethodPost, http.StatusBadRequest}, // Will fail due to missing CSRF, but method is valid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/run-script", nil)
			w := httptest.NewRecorder()

			handlers.RunScriptHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("RunScriptHandler() status = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestRunScriptHandler_CSRFValidation(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	handlers := NewHandlers(logger)

	tests := []struct {
		name           string
		csrfToken      string
		csrfHeader     string
		expectedStatus int
		expectErrorMsg string
	}{
		{
			name:           "missing CSRF token",
			csrfToken:      "",
			csrfHeader:     "",
			expectedStatus: http.StatusBadRequest,
			expectErrorMsg: "Token CSRF manquant",
		},
		{
			name:           "CSRF token in form",
			csrfToken:      "valid-token",
			csrfHeader:     "",
			expectedStatus: http.StatusInternalServerError, // Will fail due to script path not safe
			expectErrorMsg: "",
		},
		{
			name:           "CSRF token in header",
			csrfToken:      "",
			csrfHeader:     "valid-token",
			expectedStatus: http.StatusInternalServerError, // Will fail due to script path not safe
			expectErrorMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := url.Values{}
			data.Set("userId", "test123")
			data.Set("script", "script1.py")
			if tt.csrfToken != "" {
				data.Set("csrf_token", tt.csrfToken)
			}

			req := httptest.NewRequest(http.MethodPost, "/run-script", strings.NewReader(data.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if tt.csrfHeader != "" {
				req.Header.Set("X-CSRF-Token", tt.csrfHeader)
			}

			w := httptest.NewRecorder()
			handlers.RunScriptHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("RunScriptHandler() status = %d, want %d", w.Code, tt.expectedStatus)
			}

			if tt.expectErrorMsg != "" {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				} else if message, ok := response["message"].(string); !ok || message != tt.expectErrorMsg {
					t.Errorf("RunScriptHandler() message = %s, want %s", message, tt.expectErrorMsg)
				}
			}
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		forwardedFor   string
		realIP         string
		expectedResult string
	}{
		{
			name:           "direct connection",
			remoteAddr:     "192.168.1.1:8080",
			expectedResult: "192.168.1.1:8080",
		},
		{
			name:           "forwarded for header",
			remoteAddr:     "127.0.0.1:8080",
			forwardedFor:   "203.0.113.1, 198.51.100.1",
			expectedResult: "203.0.113.1",
		},
		{
			name:           "real IP header",
			remoteAddr:     "127.0.0.1:8080",
			realIP:         "203.0.113.1",
			expectedResult: "203.0.113.1",
		},
		{
			name:           "both headers - forwarded takes precedence",
			remoteAddr:     "127.0.0.1:8080",
			forwardedFor:   "203.0.113.1",
			realIP:         "198.51.100.1",
			expectedResult: "203.0.113.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.forwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.forwardedFor)
			}
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}

			result := getClientIP(req)
			if result != tt.expectedResult {
				t.Errorf("getClientIP() = %s, want %s", result, tt.expectedResult)
			}
		})
	}
}

func TestGenerateSecureCSRFToken(t *testing.T) {
	// Test multiple generations to ensure uniqueness
	tokens := make(map[string]bool)

	for i := 0; i < 100; i++ {
		token, err := generateSecureCSRFToken()
		if err != nil {
			t.Errorf("generateSecureCSRFToken() error = %v", err)
		}

		// Check length (32 bytes = 64 hex characters)
		if len(token) != 64 {
			t.Errorf("generateSecureCSRFToken() length = %d, want 64", len(token))
		}

		// Check uniqueness
		if tokens[token] {
			t.Errorf("generateSecureCSRFToken() generated duplicate token: %s", token)
		}
		tokens[token] = true

		// Check format (hex string)
		for _, char := range token {
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
				t.Errorf("generateSecureCSRFToken() contains invalid hex character: %c", char)
			}
		}
	}
}

func TestSendJSONResponse(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	handlers := NewHandlers(logger)

	data := map[string]interface{}{
		"status":  "success",
		"message": "Test message",
		"data":    []string{"item1", "item2"},
	}

	w := httptest.NewRecorder()
	handlers.sendJSONResponse(w, data)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("sendJSONResponse() status = %d, want %d", w.Code, http.StatusOK)
	}

	// Check content type
	expectedContentType := "application/json"
	if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("sendJSONResponse() Content-Type = %s, want %s", contentType, expectedContentType)
	}

	// Check JSON content
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("sendJSONResponse() failed to unmarshal response: %v", err)
	}

	if response["status"] != data["status"] {
		t.Errorf("sendJSONResponse() status = %v, want %v", response["status"], data["status"])
	}
}

func TestSendJSONError(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	handlers := NewHandlers(logger)

	message := "Test error message"
	statusCode := http.StatusBadRequest

	w := httptest.NewRecorder()
	handlers.sendJSONError(w, message, statusCode)

	// Check status code
	if w.Code != statusCode {
		t.Errorf("sendJSONError() status = %d, want %d", w.Code, statusCode)
	}

	// Check content type
	expectedContentType := "application/json"
	if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("sendJSONError() Content-Type = %s, want %s", contentType, expectedContentType)
	}

	// Check JSON content
	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("sendJSONError() failed to unmarshal response: %v", err)
	}

	if response["status"] != "error" {
		t.Errorf("sendJSONError() status = %s, want error", response["status"])
	}
	if response["message"] != message {
		t.Errorf("sendJSONError() message = %s, want %s", response["message"], message)
	}
}

// Mock executor for testing
type mockExecutor struct {
	shouldSucceed bool
	output        string
	duration      time.Duration
}

func (m *mockExecutor) Execute(ctx context.Context, req scripts.ExecutionRequest) (*scripts.ExecutionResult, error) {
	return &scripts.ExecutionResult{
		Success:    m.shouldSucceed,
		Output:     m.output,
		Duration:   m.duration,
		ExecutedAt: time.Now(),
		ExitCode:   0,
	}, nil
}

func BenchmarkValidateUserID(b *testing.B) {
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	handlers := NewHandlers(logger)
	userID := "test1234"

	for i := 0; i < b.N; i++ {
		handlers.validateUserID(userID)
	}
}

func BenchmarkValidateScript(b *testing.B) {
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	handlers := NewHandlers(logger)
	script := "script1.py"

	for i := 0; i < b.N; i++ {
		handlers.validateScript(script)
	}
}

func BenchmarkGenerateCSRFToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateSecureCSRFToken()
	}
}
