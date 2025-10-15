package unifi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// FirewallGroup represents a UniFi firewall group
type FirewallGroup struct {
	ID      string   `json:"_id,omitempty"`
	Name    string   `json:"name"`
	Type    string   `json:"group_type"`
	Members []string `json:"group_members"`
}

// GetFirewallGroup retrieves a firewall group by name
func (c *Client) GetFirewallGroup(ctx context.Context, name string) (*FirewallGroup, error) {
	if err := c.ensureLoggedIn(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/proxy/network/api/s/%s/rest/firewallgroup", c.baseURL, c.config.Site)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []FirewallGroup `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Find group by name
	for _, group := range result.Data {
		if group.Name == name {
			return &group, nil
		}
	}

	return nil, fmt.Errorf("group not found: %s", name)
}

// CreateFirewallGroup creates a new firewall group
func (c *Client) CreateFirewallGroup(ctx context.Context, name string, members []string) (*FirewallGroup, error) {
	if err := c.ensureLoggedIn(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/proxy/network/api/s/%s/rest/firewallgroup", c.baseURL, c.config.Site)

	group := FirewallGroup{
		Name:    name,
		Type:    "address-group",
		Members: members,
	}

	body, err := json.Marshal(group)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal group: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []FirewallGroup `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no group returned in response")
	}

	return &result.Data[0], nil
}

// UpdateFirewallGroup updates an existing firewall group
func (c *Client) UpdateFirewallGroup(ctx context.Context, groupID string, members []string) error {
	if err := c.ensureLoggedIn(ctx); err != nil {
		return err
	}

	url := fmt.Sprintf("%s/proxy/network/api/s/%s/rest/firewallgroup/%s", c.baseURL, c.config.Site, groupID)

	update := map[string]interface{}{
		"group_members": members,
	}

	body, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal update: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
