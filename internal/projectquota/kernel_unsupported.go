//go:build !linux || !amd64

package projectquota

import "os"

// Unsupported platforms refuse. Never substitute directory size, filesystem
// free space, labels, polling or host execution for actual project enforcement.
func newKernelReader(*os.File) (kernelReader, error) { return nil, ErrUnavailable }
