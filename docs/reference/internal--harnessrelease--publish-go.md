# internal/harnessrelease/publish.go

[Module index](internal.md) · [Open source](../../internal/harnessrelease/publish.go)

Go implementation for publish; declarations and calls are listed below.

Package: `harnessrelease`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`ReadPrivate`** — ReadPrivate binds bounds, ownership and permissions to the opened inode. Parent directories and the signing workstation remain trusted infrastructure.
  Signature: `func(path string, max int64) ([]byte, error)`.
  Calls: `f.Close`, `f.Stat`, `filepath.IsAbs`, `info.Mode`, `info.Mode().IsRegular`, `info.Mode().Perm`, `info.Size`, `info.Sys`, `int64`, `io.LimitReader`, `io.ReadAll`, `len`, `os.Geteuid`, `os.OpenFile`, `uint32`.
- **`Sign`** — Sign validates an operator-authored canonical review before signing it. The seed/public components must agree; no key generation or provider access occurs here. The private key never becomes part of the published envelope.
  Signature: `func(payload []byte, key ed25519.PrivateKey, now time.Time) ([]byte, error)`.
  Calls: `Verify`, `base64.StdEncoding.EncodeToString`, `bytes.Equal`, `canonicalJSON`, `ed25519.NewKeyFromSeed`, `ed25519.Sign`, `json.Marshal`, `json.RawMessage`, `key.Public`, `len`.
- **`Publish`** — Publish writes a complete owner-only envelope before making it visible with an exclusive hard-link publication. Existing paths (including symlinks) are never overwritten. Keep the destination directory private and trusted.
  Signature: `func(path string, envelope []byte) error`.
  Calls: `f.Close`, `f.Name`, `f.Sync`, `f.Write`, `filepath.Dir`, `filepath.IsAbs`, `len`, `os.CreateTemp`, `os.Link`, `os.Remove`.
