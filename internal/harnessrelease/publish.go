package harnessrelease

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// ReadPrivate binds bounds, ownership and permissions to the opened inode.
// Parent directories and the signing workstation remain trusted infrastructure.
func ReadPrivate(path string, max int64) ([]byte, error) {
	if !filepath.IsAbs(path) || max < 1 || max > 4096 {
		return nil, ErrReview
	}
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, ErrReview
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, ErrReview
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 || info.Size() > max || stat.Uid != uint32(os.Geteuid()) || stat.Nlink != 1 {
		return nil, ErrReview
	}
	raw, err := io.ReadAll(io.LimitReader(f, max+1))
	if err != nil || int64(len(raw)) > max {
		return nil, ErrReview
	}
	return raw, nil
}

// Sign validates an operator-authored canonical review before signing it.
// The seed/public components must agree; no key generation or provider access
// occurs here. The private key never becomes part of the published envelope.
func Sign(payload []byte, key ed25519.PrivateKey, now time.Time) ([]byte, error) {
	var review Review
	if len(key) != ed25519.PrivateKeySize || !bytes.Equal(key, ed25519.NewKeyFromSeed(key[:ed25519.SeedSize])) || !canonicalJSON(payload, &review) {
		return nil, ErrReview
	}
	envelope := Envelope{Review: json.RawMessage(payload), Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(key, payload))}
	raw, err := json.Marshal(envelope)
	if err != nil {
		return nil, ErrReview
	}
	if _, err = Verify(raw, key.Public().(ed25519.PublicKey), now); err != nil {
		return nil, err
	}
	return raw, nil
}

// Publish writes a complete owner-only envelope before making it visible with
// an exclusive hard-link publication. Existing paths (including symlinks) are
// never overwritten. Keep the destination directory private and trusted.
func Publish(path string, envelope []byte) error {
	if !filepath.IsAbs(path) || len(envelope) == 0 || len(envelope) > 4096 {
		return ErrReview
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".harness-review-*")
	if err != nil {
		return ErrReview
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err = f.Write(envelope); err != nil {
		f.Close()
		return ErrReview
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return ErrReview
	}
	if err = f.Close(); err != nil {
		return ErrReview
	}
	if err = os.Link(tmp, path); err != nil {
		return ErrReview
	}
	return nil
}
