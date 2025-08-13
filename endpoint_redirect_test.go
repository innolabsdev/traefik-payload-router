package traefik_payload_router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateConfig(t *testing.T) {
	config := CreateConfig()

	if config == nil {
		t.Fatal("CreateConfig returned nil")
	}
	if config.RedirectMappings == nil {
		t.Fatal("RedirectMappings should not be nil")
	}
	if config.DefaultRedirect != "" {
		t.Errorf("Expected DefaultRedirect to be empty, got %s", config.DefaultRedirect)
	}
	if config.WebhookPath != "/webhooks" {
		t.Errorf("Expected WebhookPath to be '/webhooks', got %s", config.WebhookPath)
	}
	if config.StatusCode != 302 {
		t.Errorf("Expected StatusCode to be 302, got %d", config.StatusCode)
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &Config{
				RedirectMappings: map[string]string{
					"KHUID000001": "https://api.example.com/webhook1",
				},
				DefaultRedirect: "https://example.com/webhook",
			},
			expectError: false,
		},
		{
			name: "invalid redirect URL",
			config: &Config{
				RedirectMappings: map[string]string{
					"KHUID000001": "://invalid-url",
				},
			},
			expectError: true,
		},
		{
			name: "invalid default redirect URL",
			config: &Config{
				DefaultRedirect: "://invalid-url",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

			handler, err := New(ctx, next, tt.config, "test-plugin")

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if handler != nil {
					t.Error("Expected nil handler when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if handler == nil {
					t.Error("Expected non-nil handler when no error")
				}
			}
		})
	}
}

func TestEndpointRedirect_ServeHTTP(t *testing.T) {
	// Mock server to capture forwarded requests
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Forwarded successfully"))
	}))
	defer mockServer.Close()

	// Test webhook request with matching endpoint_id
	t.Run("webhook request with matching endpoint_id", func(t *testing.T) {
		config := &Config{
			RedirectMappings: map[string]string{
				"KHUID000001": mockServer.URL,
			},
			WebhookPath: "/webhooks",
			StatusCode:  302,
		}

		ctx := context.Background()
		nextCalled := false
		next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			nextCalled = true
			rw.WriteHeader(http.StatusOK)
		})

		handler, err := New(ctx, next, config, "test-plugin")
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}

		jsonPayload := `{
			"type": "DOCUMENT.RECEIVED",
			"document_id": "df5dcf78-e78b-4396-87bb-2d2a44c1302e",
			"endpoint_id": "KHUID000001"
		}`

		req := httptest.NewRequest("POST", "/webhooks", strings.NewReader(jsonPayload))
		req.Header.Set("Content-Type", "application/json")
		rw := httptest.NewRecorder()

		handler.ServeHTTP(rw, req)

		if nextCalled {
			t.Error("Expected next handler NOT to be called when forwarding")
		}

		// Check that we got a successful response (forwarded from mock server)
		if rw.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
		}
	})

	// Test non-webhook request should pass through
	t.Run("non-webhook request passes through", func(t *testing.T) {
		config := &Config{
			RedirectMappings: map[string]string{
				"KHUID000001": mockServer.URL,
			},
			WebhookPath: "/webhooks",
		}

		ctx := context.Background()
		nextCalled := false
		next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			nextCalled = true
			rw.WriteHeader(http.StatusOK)
		})

		handler, err := New(ctx, next, config, "test-plugin")
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}

		req := httptest.NewRequest("GET", "/other-path", nil)
		rw := httptest.NewRecorder()

		handler.ServeHTTP(rw, req)

		if !nextCalled {
			t.Error("Expected next handler to be called for non-webhook request")
		}
	})

	// Test GET request to webhook path should pass through
	t.Run("GET request to webhook path passes through", func(t *testing.T) {
		config := &Config{
			RedirectMappings: map[string]string{
				"KHUID000001": mockServer.URL,
			},
			WebhookPath: "/webhooks",
		}

		ctx := context.Background()
		nextCalled := false
		next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			nextCalled = true
			rw.WriteHeader(http.StatusOK)
		})

		handler, err := New(ctx, next, config, "test-plugin")
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}

		req := httptest.NewRequest("GET", "/webhooks", nil)
		rw := httptest.NewRecorder()

		handler.ServeHTTP(rw, req)

		if !nextCalled {
			t.Error("Expected next handler to be called for GET request")
		}
	})

	// Test invalid JSON should pass through
	t.Run("invalid JSON passes through", func(t *testing.T) {
		config := &Config{
			RedirectMappings: map[string]string{
				"KHUID000001": mockServer.URL,
			},
			WebhookPath: "/webhooks",
		}

		ctx := context.Background()
		nextCalled := false
		next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			nextCalled = true
			rw.WriteHeader(http.StatusOK)
		})

		handler, err := New(ctx, next, config, "test-plugin")
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}

		req := httptest.NewRequest("POST", "/webhooks", strings.NewReader("invalid json"))
		rw := httptest.NewRecorder()

		handler.ServeHTTP(rw, req)

		if !nextCalled {
			t.Error("Expected next handler to be called for invalid JSON")
		}
	})

	// Test missing endpoint_id should pass through
	t.Run("missing endpoint_id passes through", func(t *testing.T) {
		config := &Config{
			RedirectMappings: map[string]string{
				"KHUID000001": mockServer.URL,
			},
			WebhookPath: "/webhooks",
		}

		ctx := context.Background()
		nextCalled := false
		next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			nextCalled = true
			rw.WriteHeader(http.StatusOK)
		})

		handler, err := New(ctx, next, config, "test-plugin")
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}

		jsonPayload := `{
			"type": "DOCUMENT.RECEIVED",
			"document_id": "df5dcf78-e78b-4396-87bb-2d2a44c1302e"
		}`

		req := httptest.NewRequest("POST", "/webhooks", strings.NewReader(jsonPayload))
		rw := httptest.NewRecorder()

		handler.ServeHTTP(rw, req)

		if !nextCalled {
			t.Error("Expected next handler to be called when endpoint_id is missing")
		}
	})

	// Test unknown endpoint_id with default redirect
	t.Run("unknown endpoint_id with default redirect", func(t *testing.T) {
		config := &Config{
			RedirectMappings: map[string]string{
				"KHUID000001": mockServer.URL,
			},
			DefaultRedirect: mockServer.URL,
			WebhookPath:     "/webhooks",
		}

		ctx := context.Background()
		nextCalled := false
		next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			nextCalled = true
			rw.WriteHeader(http.StatusOK)
		})

		handler, err := New(ctx, next, config, "test-plugin")
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}

		jsonPayload := `{
			"type": "DOCUMENT.RECEIVED",
			"document_id": "df5dcf78-e78b-4396-87bb-2d2a44c1302e",
			"endpoint_id": "UNKNOWN_ID"
		}`

		req := httptest.NewRequest("POST", "/webhooks", strings.NewReader(jsonPayload))
		rw := httptest.NewRecorder()

		handler.ServeHTTP(rw, req)

		if nextCalled {
			t.Error("Expected next handler NOT to be called when using default redirect")
		}

		// Should forward to default redirect
		if rw.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
		}
	})
}
