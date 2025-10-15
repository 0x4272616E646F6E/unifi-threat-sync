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

// PlainParser parses plain text feeds with one IP/CIDR per line
type PlainParser struct{}

func init() {
	Register(&PlainParser{})
}

// Name returns the parser identifier
func (p *PlainParser) Name() string {
	return "plain"
}

// Parse fetches and parses a plain text feed
func (p *PlainParser) Parse(ctx context.Context, feedConfig config.FeedConfig) ([]net.IPNet, error) {
	// Parse timeout
	timeout := 30 * time.Second
	if feedConfig.Timeout != "" {
		if t, err := time.ParseDuration(feedConfig.Timeout); err == nil {
			timeout = t
		}
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: timeout,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", feedConfig.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
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

// parseBody parses the response body line by line
func (p *PlainParser) parseBody(body io.Reader) ([]net.IPNet, error) {
	var networks []net.IPNet
	scanner := bufio.NewScanner(body)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip comments (common prefixes)
		if strings.HasPrefix(line, "#") ||
			strings.HasPrefix(line, ";") ||
			strings.HasPrefix(line, "//") {
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

// ValidateConfig validates the plain parser configuration
func (p *PlainParser) ValidateConfig(feedConfig config.FeedConfig) error {
	if feedConfig.URL == "" {
		return fmt.Errorf("url is required")
	}
	return nil
}
