package unifi

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/0x4272616E646F6E/unifi-threat-sync/internal/config"
)

// Client represents a UniFi controller client
type Client struct {
	config     config.UniFiConfig
	httpClient *http.Client
	baseURL    string
	loggedIn   bool
}

// NewClient creates a new UniFi client
func NewClient(cfg config.UniFiConfig) (*Client, error) {
	// Create cookie jar for session management
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	// Create HTTP client with cookie jar and TLS skip verify for self-signed certs
	httpClient := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // UniFi controllers often use self-signed certs
			},
		},
	}

	return &Client{
		config:     cfg,
		httpClient: httpClient,
		baseURL:    cfg.URL,
		loggedIn:   false,
	}, nil
}

// ensureLoggedIn checks if logged in and logs in if needed
func (c *Client) ensureLoggedIn(ctx context.Context) error {
	if c.loggedIn {
		return nil
	}
	return c.Login(ctx)
}
