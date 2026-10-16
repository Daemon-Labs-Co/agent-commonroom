# cmd/server/profile_images.go

[Module index](cmd.md) · [Open source](../../cmd/server/profile_images.go)

Go implementation for profile images; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`humanProfileImagesEnabled`** — humanProfileImagesEnabled defaults on; only the owner can persist an override.
  Signature: `func(st room.State) bool`.
- **`normalizeProfileImage`** — normalizeProfileImage validates bounded raster input and re-encodes it as JPEG to discard metadata and prohibit SVG, animated images and external URLs.
  Signature: `func(value string) (string, error)`.
  Calls: `base64.StdEncoding.DecodeString`, `base64.StdEncoding.EncodeToString`, `bytes.NewReader`, `errors.New`, `image.Decode`, `image.DecodeConfig`, `jpeg.Encode`, `len`, `out.Bytes`, `out.Len`, `strings.HasPrefix`, `strings.TrimPrefix`.
