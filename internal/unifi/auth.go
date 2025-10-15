package unifi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// loginRequest represents the login request body
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

// Login authenticates with the UniFi controller
func (c *Client) Login(ctx context.Context) error {
	loginURL := fmt.Sprintf("%s/api/auth/login", c.baseURL)

	// Prepare login payload
	payload := loginRequest{
		Username: c.config.Username,
		Password: c.config.Password,
		Remember: true,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", loginURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	c.loggedIn = true
	return nil
}

// Logout logs out from the UniFi controller
func (c *Client) Logout(ctx context.Context) error {
	if !c.loggedIn {
		return nil
	}

	logoutURL := fmt.Sprintf("%s/api/auth/logout", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "POST", logoutURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create logout request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("logout request failed: %w", err)
	}
	defer resp.Body.Close()

	c.loggedIn = false
	return nil
}
