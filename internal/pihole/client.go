package pihole

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"pihole-analyzer/internal/logger"
)

// Client represents the Pi-hole API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	SID        string
	CSRFToken  string
	Logger     *logger.Logger
	config     *Config
}

// Config represents Pi-hole API configuration
type Config struct {
	Host      string
	Port      int
	Password  string
	TOTP      string // For 2FA
	UseHTTPS  bool
	Timeout   time.Duration
	VerifyTLS bool
}

// NewClient creates a new Pi-hole API client
func NewClient(config *Config, log *logger.Logger) *Client {
	if log == nil {
		log = logger.New(&logger.Config{
			Level:        logger.LevelInfo,
			EnableColors: true,
			EnableEmojis: true,
			Component:    "pihole-api",
		})
	}

	scheme := "http"
	if config.UseHTTPS {
		scheme = "https"
	}

	baseURL := fmt.Sprintf("%s://%s:%d", scheme, config.Host, config.Port)

	// Create HTTP client with appropriate TLS configuration
	transport := &http.Transport{}
	if config.UseHTTPS && !config.VerifyTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	return &Client{
		BaseURL:    baseURL,
		HTTPClient: httpClient,
		Logger:     log,
		config:     config,
	}
}

// SessionInfo represents Pi-hole API session information
type SessionInfo struct {
	SID        string    `json:"sid"`
	CSRFToken  string    `json:"csrf"`
	ValidUntil time.Time `json:"valid_until"`
	Validity   int       `json:"validity"`
	TOTPReq    bool      `json:"totp"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Session SessionInfo `json:"session"`
	Took    float64     `json:"took"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error struct {
		Key     string      `json:"key"`
		Message string      `json:"message"`
		Hint    interface{} `json:"hint"`
	} `json:"error"`
	Took float64 `json:"took"`
}

// Authenticate performs authentication with Pi-hole API
func (c *Client) Authenticate(ctx context.Context) error {
	c.Logger.Info("Authenticating with Pi-hole API: host=%s port=%d https=%t",
		c.config.Host, c.config.Port, c.config.UseHTTPS)

	// Prepare authentication payload
	payload := map[string]string{
		"password": c.config.Password,
	}

	// Add TOTP if configured
	if c.config.TOTP != "" {
		payload["totp"] = c.config.TOTP
	}

	// Make authentication request
	resp, err := c.makeRequest(ctx, "POST", "/api/auth", payload, false)
	if err != nil {
		c.Logger.Error("Authentication request failed: %v", err)
		return fmt.Errorf("authentication request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if json.Unmarshal(body, &errorResp) == nil {
			c.Logger.Error("Authentication failed: error_key=%s error_message=%s status_code=%d",
				errorResp.Error.Key, errorResp.Error.Message, resp.StatusCode)
			return fmt.Errorf("authentication failed: %s", errorResp.Error.Message)
		}
		return fmt.Errorf("authentication failed with status %d", resp.StatusCode)
	}

	// Parse successful response
	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return fmt.Errorf("failed to parse authentication response: %w", err)
	}

	// Store session information
	c.SID = authResp.Session.SID
	c.CSRFToken = authResp.Session.CSRFToken

	c.Logger.Info("Authentication successful: sid=%s totp_required=%t validity_seconds=%d",
		c.SID[:8]+"...", authResp.Session.TOTPReq, authResp.Session.Validity)

	return nil
}

// makeRequest makes an HTTP request to the Pi-hole API
func (c *Client) makeRequest(ctx context.Context, method, path string, payload interface{}, requireAuth bool) (*http.Response, error) {
	// Build URL
	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	var reqBody io.Reader
	var contentType string

	// Handle payload
	if payload != nil {
		if method == "GET" {
			// For GET requests, add payload as query parameters
			if params, ok := payload.(map[string]string); ok {
				q := u.Query()
				for key, value := range params {
					q.Set(key, value)
				}
				u.RawQuery = q.Encode()
			}
		} else {
			// For other methods, JSON encode the payload
			jsonData, err := json.Marshal(payload)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal payload: %w", err)
			}
			reqBody = bytes.NewBuffer(jsonData)
			contentType = "application/json"
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// Add authentication headers if required
	if requireAuth && c.SID != "" {
		req.Header.Set("X-FTL-SID", c.SID)
		if c.CSRFToken != "" {
			req.Header.Set("X-FTL-CSRF", c.CSRFToken)
		}
	}

	// Make request
	return c.HTTPClient.Do(req)
}

// Close cleans up the client (logout if authenticated)
func (c *Client) Close(ctx context.Context) error {
	if c.SID != "" {
		c.Logger.Info("Logging out from Pi-hole API")

		resp, err := c.makeRequest(ctx, "DELETE", "/api/auth", nil, true)
		if err != nil {
			c.Logger.Warn("Failed to logout: %v", err)
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusGone {
			c.Logger.Info("Successfully logged out")
		} else {
			c.Logger.Warn("Unexpected logout response: status_code=%d", resp.StatusCode)
		}

		// Clear session data
		c.SID = ""
		c.CSRFToken = ""
	}

	return nil
}

// IsAuthenticated returns true if the client has a valid session
func (c *Client) IsAuthenticated() bool {
	return c.SID != ""
}
