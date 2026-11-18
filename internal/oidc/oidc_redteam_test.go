package oidc

import "testing"

// RT-6: the "DNS host name only, never loopback or an IP" rule is meant to stop the server
// being pointed at internal addresses, but it is a string comparison and many spellings of
// loopback and private addresses pass it.
func TestRedteamIssuerHostFilterAcceptsLoopbackSpellings(t *testing.T) {
	t.Parallel()
	for _, issuer := range []string{
		"https://LOCALHOST",        // case: the check is case-sensitive
		"https://localhost.",       // trailing dot resolves to loopback
		"https://foo.LOCALHOST",    // suffix check is case-sensitive
		"https://2130706433",       // decimal form of 127.0.0.1
		"https://0x7f.0.0.1",       // hex octet
		"https://127.1",            // short form
		"https://[fe80::1%25eth0]", // zoned IPv6 literal
	} {
		if _, err := New(Config{Issuer: issuer, ClientID: "c"}); err == nil {
			t.Errorf("issuer %q accepted although it addresses the local machine", issuer)
		}
	}
}
