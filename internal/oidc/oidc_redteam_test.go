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

// RT-6 (dial time): a name that resolves to an internal address is refused when the
// connection is made, whatever the name looks like.
func TestRedteamInternalAddressesRefusedAtDialTime(t *testing.T) {
	t.Parallel()
	for _, address := range []string{"127.0.0.1:443", "10.1.2.3:443", "192.168.0.9:443", "169.254.169.254:80", "[::1]:443", "[fe80::1]:443", "0.0.0.0:443", "[fd00::1]:443"} {
		if refuseInternalAddress("tcp", address, nil) == nil {
			t.Errorf("%s was allowed", address)
		}
	}
	for _, address := range []string{"93.184.216.34:443", "[2606:2800:220:1:248:1893:25c8:1946]:443"} {
		if err := refuseInternalAddress("tcp", address, nil); err != nil {
			t.Errorf("%s was refused: %v", address, err)
		}
	}
}
