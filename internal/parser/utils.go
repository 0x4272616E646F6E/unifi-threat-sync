package parser

import (
	"fmt"
	"net"
	"strings"
)

// parseIPOrCIDR parses a string as either an IP address or CIDR block
func parseIPOrCIDR(s string) (net.IPNet, error) {
	s = strings.TrimSpace(s)

	// Try parsing as CIDR first
	if strings.Contains(s, "/") {
		_, ipnet, err := net.ParseCIDR(s)
		if err != nil {
			return net.IPNet{}, fmt.Errorf("invalid CIDR: %w", err)
		}
		return *ipnet, nil
	}

	// Parse as IP address
	ip := net.ParseIP(s)
	if ip == nil {
		return net.IPNet{}, fmt.Errorf("invalid IP address: %s", s)
	}

	// Convert to CIDR with /32 for IPv4 or /128 for IPv6
	var mask net.IPMask
	if ip.To4() != nil {
		mask = net.CIDRMask(32, 32)
	} else {
		mask = net.CIDRMask(128, 128)
	}

	return net.IPNet{
		IP:   ip,
		Mask: mask,
	}, nil
}

// isValidIP checks if a string is a valid IP address
func isValidIP(s string) bool {
	return net.ParseIP(s) != nil
}

// isValidCIDR checks if a string is a valid CIDR block
func isValidCIDR(s string) bool {
	_, _, err := net.ParseCIDR(s)
	return err == nil
}
