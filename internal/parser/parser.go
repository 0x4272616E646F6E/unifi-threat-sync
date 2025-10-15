package parser

import (
	"context"
	"fmt"
	"net"

	"github.com/0x4272616E646F6E/unifi-threat-sync/internal/config"
)

// Parser defines the interface all feed parsers must implement
type Parser interface {
	// Name returns the parser identifier (e.g., "plain", "netset", "abuseipdb")
	Name() string

	// Parse fetches and parses the feed, returning a list of IPs/CIDRs
	Parse(ctx context.Context, feedConfig config.FeedConfig) ([]net.IPNet, error)

	// ValidateConfig validates parser-specific configuration
	ValidateConfig(feedConfig config.FeedConfig) error
}

// Registry holds all registered parsers
var registry = make(map[string]Parser)

// Register adds a parser to the registry
func Register(p Parser) {
	registry[p.Name()] = p
}

// Get retrieves a parser by name
func Get(name string) (Parser, error) {
	p, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown parser: %s", name)
	}
	return p, nil
}

// List returns all registered parser names
func List() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
