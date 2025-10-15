package normalizer

import "net"

// Merge combines multiple lists of IP networks and normalizes them
func Merge(lists ...[]net.IPNet) []net.IPNet {
	// Calculate total capacity
	capacity := 0
	for _, list := range lists {
		capacity += len(list)
	}

	// Merge all lists
	merged := make([]net.IPNet, 0, capacity)
	for _, list := range lists {
		merged = append(merged, list...)
	}

	// Normalize (deduplicate and sort)
	return Normalize(merged)
}
