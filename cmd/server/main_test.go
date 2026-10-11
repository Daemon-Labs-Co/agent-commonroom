package main

import "testing"

func TestChecksumDetectsAlteration(t *testing.T) {
	t.Parallel()
	original := checksum("prior", "message.created", `{"text": "hello"}`)
	for _, v := range []string{checksum("other", "message.created", `{"text": "hello"}`), checksum("prior", "memory.revised", `{"text": "hello"}`), checksum("prior", "message.created", `{"text": "changed"}`)} {
		if v == original {
			t.Fatal("Alteration undetected")
		}
	}
}
