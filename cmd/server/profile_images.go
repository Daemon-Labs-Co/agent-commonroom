package main

import (
	"agentroom/internal/room"
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/jpeg"
	_ "image/png"
	"strings"
)

// humanProfileImagesEnabled defaults on; only the owner can persist an override.
func humanProfileImagesEnabled(st room.State) bool {
	enabled, exists := st.FeatureFlags["humanProfileImages"]
	return !exists || enabled
}

// normalizeProfileImage validates bounded raster input and re-encodes it as JPEG
// to discard metadata and prohibit SVG, animated images and external URLs.
func normalizeProfileImage(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	invalid := errors.New("Use a JPEG or PNG photo up to 32 KB and 256 × 256 pixels")
	if len(value) > 44000 {
		return "", invalid
	}
	var encoded string
	for _, prefix := range []string{"data:image/jpeg;base64,", "data:image/png;base64,"} {
		if strings.HasPrefix(value, prefix) {
			encoded = strings.TrimPrefix(value, prefix)
			break
		}
	}
	if encoded == "" {
		return "", invalid
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(raw) > 32768 {
		return "", invalid
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || (format != "jpeg" && format != "png") || config.Width < 1 || config.Height < 1 || config.Width > 256 || config.Height > 256 {
		return "", invalid
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return "", invalid
	}
	var out bytes.Buffer
	if jpeg.Encode(&out, img, &jpeg.Options{Quality: 80}) != nil || out.Len() > 32768 {
		return "", invalid
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(out.Bytes()), nil
}
