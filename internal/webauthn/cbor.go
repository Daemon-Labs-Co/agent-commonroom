// Package webauthn verifies WebAuthn passkey registrations and assertions using only the
// Go standard library (ADR 0006). It supports `none` attestation and ES256 / RS256
// credentials with required user verification, and refuses everything else.
package webauthn

import (
	"errors"
	"sort"
)

// A deliberately small CBOR reader: definite lengths only, no tags, no floats, bounded
// depth and size. Everything it decodes comes from an attacker-controllable browser.

var errCBOR = errors.New("invalid CBOR")

type kind int

const (
	kUint kind = iota
	kNegInt
	kBytes
	kText
	kArray
	kMap
	kBool
	kNull
)

const (
	maxCBORDepth = 8
	maxCBORItems = 64
)

type value struct {
	kind  kind
	n     int64 // kUint, kNegInt (the actual negative number), kBool (0/1)
	bytes []byte
	text  string
	array []value
	pairs []pair
}

type pair struct{ key, val value }

// decode reads one item and returns the remaining bytes.
func decode(data []byte) (value, []byte, error) {
	items := 0
	return decodeItem(data, 0, &items)
}

func readHead(data []byte) (major byte, arg uint64, rest []byte, err error) {
	if len(data) == 0 {
		return 0, 0, nil, errCBOR
	}
	major, info := data[0]>>5, data[0]&0x1f
	data = data[1:]
	switch {
	case info < 24:
		return major, uint64(info), data, nil
	case info == 24 && len(data) >= 1:
		return major, uint64(data[0]), data[1:], nil
	case info == 25 && len(data) >= 2:
		return major, uint64(data[0])<<8 | uint64(data[1]), data[2:], nil
	case info == 26 && len(data) >= 4:
		return major, uint64(data[0])<<24 | uint64(data[1])<<16 | uint64(data[2])<<8 | uint64(data[3]), data[4:], nil
	case info == 27 && len(data) >= 8:
		var v uint64
		for _, b := range data[:8] {
			v = v<<8 | uint64(b)
		}
		return major, v, data[8:], nil
	}
	return 0, 0, nil, errCBOR // reserved or indefinite lengths
}

func decodeItem(data []byte, depth int, items *int) (value, []byte, error) {
	*items++
	if depth > maxCBORDepth || *items > maxCBORItems {
		return value{}, nil, errCBOR
	}
	major, arg, rest, err := readHead(data)
	if err != nil {
		return value{}, nil, err
	}
	switch major {
	case 0:
		if arg > 1<<62 {
			return value{}, nil, errCBOR
		}
		return value{kind: kUint, n: int64(arg)}, rest, nil
	case 1:
		if arg > 1<<62 {
			return value{}, nil, errCBOR
		}
		return value{kind: kNegInt, n: -1 - int64(arg)}, rest, nil
	case 2, 3:
		if arg > uint64(len(rest)) {
			return value{}, nil, errCBOR
		}
		chunk := append([]byte(nil), rest[:arg]...)
		if major == 2 {
			return value{kind: kBytes, bytes: chunk}, rest[arg:], nil
		}
		return value{kind: kText, text: string(chunk)}, rest[arg:], nil
	case 4, 5:
		count := arg
		if major == 5 {
			if count > maxCBORItems/2 {
				return value{}, nil, errCBOR
			}
		} else if count > maxCBORItems {
			return value{}, nil, errCBOR
		}
		if count > uint64(len(rest)) { // every item needs at least one byte
			return value{}, nil, errCBOR
		}
		if major == 4 {
			out := value{kind: kArray}
			for i := uint64(0); i < count; i++ {
				var item value
				if item, rest, err = decodeItem(rest, depth+1, items); err != nil {
					return value{}, nil, err
				}
				out.array = append(out.array, item)
			}
			return out, rest, nil
		}
		out := value{kind: kMap}
		for i := uint64(0); i < count; i++ {
			var k, v value
			if k, rest, err = decodeItem(rest, depth+1, items); err != nil {
				return value{}, nil, err
			}
			if k.kind != kUint && k.kind != kNegInt && k.kind != kText {
				return value{}, nil, errCBOR
			}
			for _, p := range out.pairs {
				if p.key.kind == k.kind && p.key.n == k.n && p.key.text == k.text {
					return value{}, nil, errCBOR // duplicate key
				}
			}
			if v, rest, err = decodeItem(rest, depth+1, items); err != nil {
				return value{}, nil, err
			}
			out.pairs = append(out.pairs, pair{k, v})
		}
		return out, rest, nil
	case 7:
		switch arg {
		case 20:
			return value{kind: kBool, n: 0}, rest, nil
		case 21:
			return value{kind: kBool, n: 1}, rest, nil
		case 22:
			return value{kind: kNull}, rest, nil
		}
	}
	return value{}, nil, errCBOR // tags (6), floats, undefined, simple values, break
}

func (v value) intKey(n int64) (value, bool) {
	if v.kind != kMap {
		return value{}, false
	}
	for _, p := range v.pairs {
		if (p.key.kind == kUint || p.key.kind == kNegInt) && p.key.n == n {
			return p.val, true
		}
	}
	return value{}, false
}

func (v value) textKey(name string) (value, bool) {
	if v.kind != kMap {
		return value{}, false
	}
	for _, p := range v.pairs {
		if p.key.kind == kText && p.key.text == name {
			return p.val, true
		}
	}
	return value{}, false
}

func (v value) keys() []string {
	out := []string{}
	for _, p := range v.pairs {
		if p.key.kind == kText {
			out = append(out, p.key.text)
		}
	}
	sort.Strings(out)
	return out
}
