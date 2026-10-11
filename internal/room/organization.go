package room

// OrganizationallyOutranks compares reporting ancestry, not job titles or saved
// legacy ranks. Unrelated branches and multiple profiles of one identity are peers.
func OrganizationallyOutranks(chart OrganizationChart, superior, report string) bool {
	if superior == "" || superior == report {
		return false
	}
	parents := map[string]string{}
	for _, n := range chart.Nodes {
		parents[n.ActorID] = n.ReportsTo
	}
	seen := map[string]bool{}
	for cursor := parents[report]; cursor != "" && !seen[cursor]; cursor = parents[cursor] {
		if cursor == superior {
			return true
		}
		seen[cursor] = true
	}
	return false
}

// OrganizationDecisionPeers identifies maximal available identities. Multiple
// candidates must reach consensus; an absent manager cannot resolve a conflict.
// Explicit human overrides remain final, independent of this automated candidate set.
func OrganizationDecisionPeers(chart OrganizationChart, available []string) []string {
	peers := []string{}
	seen := map[string]bool{}
	for _, candidate := range available {
		if candidate == "" || seen[candidate] {
			continue
		}
		seen[candidate] = true
		outranked := false
		for _, other := range available {
			if OrganizationallyOutranks(chart, other, candidate) {
				outranked = true
				break
			}
		}
		if !outranked {
			peers = append(peers, candidate)
		}
	}
	return peers
}
