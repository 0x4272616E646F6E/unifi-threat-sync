package parser

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/0x4272616E646F6E/unifi-threat-sync/internal/config"
)

// NetsetParser parses FireHOL netset format feeds
type NetsetParser struct{}

func init() {
	Register(&NetsetParser{})
}

// Name returns the parser identifier
func (p *NetsetParser) Name() string {
	return "netset"
}

// Parse fetches and parses a netset format feed
func (p *NetsetParser) Parse(ctx context.Context, feedConfig config.FeedConfig) ([]net.IPNet, error) {
	// Parse timeout
	timeout := 30 * time.Second
	if feedConfig.Timeout != "" {
		if t, err := time.ParseDuration(feedConfig.Timeout); err == nil {
			timeout = t
		}
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: timeout,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", feedConfig.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "UniFi-Threat-Sync/1.0")

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	return p.parseBody(resp.Body)
}

// parseBody parses the netset format body
func (p *NetsetParser) parseBody(body io.Reader) ([]net.IPNet, error) {
	var networks []net.IPNet
	scanner := bufio.NewScanner(body)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip comments (lines starting with #)
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Parse IP or CIDR
		ipnet, err := parseIPOrCIDR(line)
		if err != nil {
			// Skip invalid lines silently
			continue
		}

		networks = append(networks, ipnet)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading feed: %w", err)
	}

	if len(networks) == 0 {
		return nil, fmt.Errorf("no valid IPs found in feed")
	}

	return networks, nil
}

// ValidateConfig validates the netset parser configuration
func (p *NetsetParser) ValidateConfig(feedConfig config.FeedConfig) error {
	if feedConfig.URL == "" {
		return fmt.Errorf("url is required")
	}
	return nil
}
