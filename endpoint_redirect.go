package traefik_payload_router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Config the plugin configuration.
type Config struct {
	// FieldName is the JSON field name to use for routing decisions (default: "endpoint_id")
	FieldName string `json:"fieldName,omitempty"`

	// RedirectMappings maps field values to redirect URLs
	RedirectMappings map[string]string `json:"redirectMappings,omitempty"`

	// DefaultRedirect is the default URL to redirect to if no mapping is found
	DefaultRedirect string `json:"defaultRedirect,omitempty"`

	// WebhookPath is the path to match for webhook requests (default: "/webhooks")
	WebhookPath string `json:"webhookPath,omitempty"`

	// StatusCode is the HTTP status code to use for redirects (default: 302)
	StatusCode int `json:"statusCode,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		FieldName:        "endpoint_id",
		RedirectMappings: make(map[string]string),
		DefaultRedirect:  "",
		WebhookPath:      "/webhooks",
		StatusCode:       302,
	}
}

// EndpointRedirect a plugin to redirect based on a configurable JSON field.
type EndpointRedirect struct {
	next             http.Handler
	name             string
	fieldName        string
	redirectMappings map[string]string
	defaultRedirect  string
	webhookPath      string
	statusCode       int
}

// New creates a new EndpointRedirect plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config.FieldName == "" {
		config.FieldName = "endpoint_id"
	}

	if config.WebhookPath == "" {
		config.WebhookPath = "/webhooks"
	}

	if config.StatusCode == 0 {
		config.StatusCode = 302
	}

	// Validate redirect URLs
	for fieldValue, redirectURL := range config.RedirectMappings {
		if _, err := url.Parse(redirectURL); err != nil {
			return nil, fmt.Errorf("invalid redirect URL for '%s': %w", fieldValue, err)
		}
	}

	if config.DefaultRedirect != "" {
		if _, err := url.Parse(config.DefaultRedirect); err != nil {
			return nil, fmt.Errorf("invalid default redirect URL: %w", err)
		}
	}

	return &EndpointRedirect{
		next:             next,
		name:             name,
		fieldName:        config.FieldName,
		redirectMappings: config.RedirectMappings,
		defaultRedirect:  config.DefaultRedirect,
		webhookPath:      config.WebhookPath,
		statusCode:       config.StatusCode,
	}, nil
}

func (e *EndpointRedirect) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Only process requests to the webhook path
	if !strings.HasPrefix(req.URL.Path, e.webhookPath) {
		e.next.ServeHTTP(rw, req)
		return
	}

	// Only process POST requests (typical for webhooks)
	if req.Method != http.MethodPost {
		e.next.ServeHTTP(rw, req)
		return
	}

	// Read the request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		e.next.ServeHTTP(rw, req)
		return
	}

	// Restore the body for the next handler (in case we don't redirect)
	req.Body = io.NopCloser(bytes.NewBuffer(body))

	// Parse the JSON payload into a generic map
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		// If JSON parsing fails, continue to next handler
		e.next.ServeHTTP(rw, req)
		return
	}

	// Extract the configured field value
	fieldValue, ok := payload[e.fieldName]
	if !ok {
		e.next.ServeHTTP(rw, req)
		return
	}

	// Convert field value to string
	var fieldStr string
	switch v := fieldValue.(type) {
	case string:
		fieldStr = v
	case float64:
		fieldStr = fmt.Sprintf("%.0f", v)
	default:
		e.next.ServeHTTP(rw, req)
		return
	}

	// If field is empty, proceed to next handler
	if fieldStr == "" {
		e.next.ServeHTTP(rw, req)
		return
	}

	// Clean the field value (trim whitespace)
	fieldStr = strings.TrimSpace(fieldStr)

	// Look for a matching redirect URL
	var redirectURL string
	if mappedURL, exists := e.redirectMappings[fieldStr]; exists {
		redirectURL = mappedURL
	} else if e.defaultRedirect != "" {
		redirectURL = e.defaultRedirect
	} else {
		// No mapping found and no default redirect, proceed to next handler
		e.next.ServeHTTP(rw, req)
		return
	}

	// Parse the redirect URL
	parsedURL, err := url.Parse(redirectURL)
	if err != nil {
		// Log error and proceed to next handler
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Invalid redirect URL"))
		return
	}

	// Preserve existing query parameters from the original request
	if req.URL.RawQuery != "" {
		if parsedURL.RawQuery != "" {
			parsedURL.RawQuery = parsedURL.RawQuery + "&" + req.URL.RawQuery
		} else {
			parsedURL.RawQuery = req.URL.RawQuery
		}
	}

	// Create a new request with the same body for the redirect target
	newReq, err := http.NewRequestWithContext(req.Context(), req.Method, parsedURL.String(), bytes.NewBuffer(body))
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Failed to create redirect request"))
		return
	}

	// Copy headers from original request
	for key, values := range req.Header {
		for _, value := range values {
			newReq.Header.Add(key, value)
		}
	}

	// Forward the request to the target URL
	client := &http.Client{}
	resp, err := client.Do(newReq)
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		rw.Write([]byte("Failed to forward request"))
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			rw.Header().Add(key, value)
		}
	}

	// Copy response status and body
	rw.WriteHeader(resp.StatusCode)
	io.Copy(rw, resp.Body)
}
