//go:build !linux

package workspacemount

import "context"

// Unsupported platforms refuse; there is no host or simulated fallback.
func Observe(context.Context, string) (Proof, error) { return Proof{}, ErrRefused }
