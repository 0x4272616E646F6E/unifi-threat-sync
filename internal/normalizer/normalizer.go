package normalizer

import (
	"net"
	"sort"
)

// Normalize takes a list of IP networks and deduplicates them
func Normalize(networks []net.IPNet) []net.IPNet {
	if len(networks) == 0 {
		return networks
	}

	// Use map for deduplication
	seen := make(map[string]net.IPNet)
	for _, network := range networks {
		key := network.String()
		seen[key] = network
	}

	// Convert back to slice
	result := make([]net.IPNet, 0, len(seen))
	for _, network := range seen {
		result = append(result, network)
	}

	// Sort for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].String() < result[j].String()
	})

	return result
}

// Contains checks if a network is in the list
func Contains(networks []net.IPNet, target net.IPNet) bool {
	targetStr := target.String()
	for _, network := range networks {
		if network.String() == targetStr {
			return true
		}
	}
	return false
}

// ToStrings converts IPNet slice to string slice
func ToStrings(networks []net.IPNet) []string {
	result := make([]string, len(networks))
	for i, network := range networks {
		result[i] = network.String()
	}
	return result
}

// FromStrings converts string slice to IPNet slice
func FromStrings(strs []string) ([]net.IPNet, error) {
	result := make([]net.IPNet, 0, len(strs))
	for _, s := range strs {
		_, ipnet, err := net.ParseCIDR(s)
		if err != nil {
			// Try as single IP
			ip := net.ParseIP(s)
			if ip == nil {
				continue
			}
			var mask net.IPMask
			if ip.To4() != nil {
				mask = net.CIDRMask(32, 32)
			} else {
				mask = net.CIDRMask(128, 128)
			}
			ipnet = &net.IPNet{IP: ip, Mask: mask}
		}
		result = append(result, *ipnet)
	}
	return result, nil
}
