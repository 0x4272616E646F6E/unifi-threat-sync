package sync

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"sort"

	"github.com/0x4272616E646F6E/unifi-threat-sync/internal/config"
	"github.com/0x4272616E646F6E/unifi-threat-sync/internal/normalizer"
	"github.com/0x4272616E646F6E/unifi-threat-sync/internal/parser"
	"github.com/0x4272616E646F6E/unifi-threat-sync/internal/unifi"
)

// HealthRecorder is an interface for recording health metrics
type HealthRecorder interface {
	RecordSync()
	RecordError()
}

// Syncer handles the synchronization process
type Syncer struct {
	config         *config.Config
	unifiClient    *unifi.Client
	lastHash       string
	healthRecorder HealthRecorder
}

// New creates a new Syncer
func New(cfg *config.Config, unifiClient *unifi.Client) *Syncer {
	return &Syncer{
		config:      cfg,
		unifiClient: unifiClient,
		lastHash:    "",
	}
}

// SetHealthRecorder sets the health recorder for metrics
func (s *Syncer) SetHealthRecorder(hr HealthRecorder) {
	s.healthRecorder = hr
}

// Run performs a full synchronization cycle
func (s *Syncer) Run(ctx context.Context) error {
	fmt.Println("Starting sync cycle...")

	// Fetch and parse all enabled feeds
	allNetworks, err := s.fetchAllFeeds(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch feeds: %w", err)
	}

	fmt.Printf("Fetched %d total IPs/CIDRs from feeds\n", len(allNetworks))

	// Normalize (deduplicate and sort)
	normalized := normalizer.Normalize(allNetworks)
	fmt.Printf("After deduplication: %d unique IPs/CIDRs\n", len(normalized))

	// Calculate hash of normalized list
	currentHash := s.calculateHash(normalized)

	// Check if update is needed
	if currentHash == s.lastHash {
		fmt.Println("No changes detected, skipping update")
		return nil
	}

	fmt.Println("Changes detected, updating UniFi...")

	// Convert to strings for UniFi API
	members := normalizer.ToStrings(normalized)

	// Get or create firewall group
	group, err := s.unifiClient.GetFirewallGroup(ctx, s.config.UniFi.GroupName)
	if err != nil {
		fmt.Printf("Group '%s' not found, creating...\n", s.config.UniFi.GroupName)
		group, err = s.unifiClient.CreateFirewallGroup(ctx, s.config.UniFi.GroupName, members)
		if err != nil {
			return fmt.Errorf("failed to create firewall group: %w", err)
		}
		fmt.Printf("Created firewall group '%s'\n", s.config.UniFi.GroupName)
	} else {
		// Update existing group
		fmt.Printf("Updating firewall group '%s'...\n", s.config.UniFi.GroupName)
		if err := s.unifiClient.UpdateFirewallGroup(ctx, group.ID, members); err != nil {
			return fmt.Errorf("failed to update firewall group: %w", err)
		}
		fmt.Printf("Updated firewall group '%s'\n", s.config.UniFi.GroupName)
	}

	// Update last hash
	s.lastHash = currentHash

	// Record successful sync
	if s.healthRecorder != nil {
		s.healthRecorder.RecordSync()
	}

	fmt.Println("Sync completed successfully")
	return nil
}

// fetchAllFeeds fetches and parses all enabled feeds
func (s *Syncer) fetchAllFeeds(ctx context.Context) ([]net.IPNet, error) {
	var allNetworks []net.IPNet

	enabledFeeds := s.config.Feeds.GetEnabled()

	for _, feedConfig := range enabledFeeds {
		fmt.Printf("Fetching feed: %s (%s)\n", feedConfig.Name, feedConfig.Parser)

		// Get parser
		p, err := parser.Get(feedConfig.Parser)
		if err != nil {
			fmt.Printf("  Warning: %v, skipping\n", err)
			continue
		}

		// Parse feed
		networks, err := p.Parse(ctx, feedConfig)
		if err != nil {
			fmt.Printf("  Warning: failed to parse feed: %v, skipping\n", err)
			continue
		}

		fmt.Printf("  Found %d IPs/CIDRs\n", len(networks))
		allNetworks = append(allNetworks, networks...)
	}

	return allNetworks, nil
}

// calculateHash calculates a SHA256 hash of the normalized network list
func (s *Syncer) calculateHash(networks []net.IPNet) string {
	// Convert to sorted string list
	strs := normalizer.ToStrings(networks)
	sort.Strings(strs)

	// Create concatenated string
	combined := ""
	for _, str := range strs {
		combined += str + "\n"
	}

	// Calculate hash
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}
