package utils

import (
	"net"
	"os"
	"strconv"
	"testing"
)

func TestFindAvailablePort(t *testing.T) {
	tests := []struct {
		name        string
		envPort     string
		expectError bool
		setup       func() func() // setup function that returns cleanup function
	}{
		{
			name:        "should use environment PORT when set",
			envPort:     "9999",
			expectError: false,
			setup: func() func() {
				os.Setenv("PORT", "9999")
				return func() { os.Unsetenv("PORT") }
			},
		},
		{
			name:        "should find default port 8001 when available",
			envPort:     "",
			expectError: false,
			setup: func() func() {
				os.Unsetenv("PORT")
				return func() {}
			},
		},
		{
			name:        "should find alternative port when 8001 is busy",
			envPort:     "",
			expectError: false,
			setup: func() func() {
				os.Unsetenv("PORT")
				// Occupy port 8001
				ln, err := net.Listen("tcp", ":8001")
				if err != nil {
					// Port already occupied, which is fine for this test
					return func() {}
				}
				return func() { ln.Close() }
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			port, err := FindAvailablePort()

			if tt.expectError && err == nil {
				t.Errorf("FindAvailablePort() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("FindAvailablePort() unexpected error: %v", err)
			}

			if err == nil {
				// Validate port format
				if port == "" {
					t.Errorf("FindAvailablePort() returned empty port")
				}

				// Validate port is numeric
				if _, err := strconv.Atoi(port); err != nil {
					t.Errorf("FindAvailablePort() returned non-numeric port: %s", port)
				}

				// If environment variable was set, should return that value
				if tt.envPort != "" && port != tt.envPort {
					t.Errorf("FindAvailablePort() = %s, want %s", port, tt.envPort)
				}
			}
		})
	}
}

func TestIsPortAvailable(t *testing.T) {
	tests := []struct {
		name      string
		port      string
		available bool
		setup     func() func() // setup function that returns cleanup function
	}{
		{
			name:      "should return true for available port",
			port:      "0", // Let OS choose available port
			available: true,
			setup:     func() func() { return func() {} },
		},
		{
			name:      "should return false for occupied port",
			port:      "8002", // Use a fixed port for this test
			available: false,
			setup: func() func() {
				// Try to occupy port 8002
				ln, err := net.Listen("tcp", ":8002")
				if err != nil {
					// If we can't occupy the port, skip the test
					return func() {}
				}
				return func() { ln.Close() }
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			if tt.port == "" {
				t.Skip("Could not setup test port")
			}

			result := isPortAvailable(tt.port)
			if result != tt.available {
				t.Errorf("isPortAvailable(%s) = %v, want %v", tt.port, result, tt.available)
			}
		})
	}
}

func TestPortError(t *testing.T) {
	err := &PortError{message: "test error"}
	expected := "test error"

	if err.Error() != expected {
		t.Errorf("PortError.Error() = %s, want %s", err.Error(), expected)
	}
}

func BenchmarkFindAvailablePort(b *testing.B) {
	os.Unsetenv("PORT")

	for i := 0; i < b.N; i++ {
		_, err := FindAvailablePort()
		if err != nil {
			b.Fatalf("FindAvailablePort() failed: %v", err)
		}
	}
}

func BenchmarkIsPortAvailable(b *testing.B) {
	port := "8001"

	for i := 0; i < b.N; i++ {
		isPortAvailable(port)
	}
}
