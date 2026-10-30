package projectquota

import (
	"context"
	"encoding/binary"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

// Linux x86-64 uses syscall number 443 for quotactl_fd. Only fixed read-only
// XFS status/quota operations are reachable. ABI references are in the guide.
// Calls may block in the kernel; cancellation never detaches an in-flight query.
// No elevated helper, quota setter, command runner or privileged fallback exists.
type linuxReader struct{ fd int }

func newKernelReader(root *os.File) (kernelReader, error) {
	raw, err := root.SyscallConn()
	if err != nil {
		return nil, ErrUnavailable
	}
	fd := -1
	var errno syscall.Errno
	err = raw.Control(func(value uintptr) {
		result, _, e := syscall.Syscall(syscall.SYS_FCNTL, value, syscall.F_DUPFD_CLOEXEC, 0)
		errno = e
		if e == 0 {
			fd = int(result)
		}
	})
	if err != nil || errno != 0 || fd < 0 {
		if fd >= 0 {
			_ = syscall.Close(fd)
		}
		return nil, ErrUnavailable
	}
	return &linuxReader{fd}, nil
}
func (r *linuxReader) Close() error { return syscall.Close(r.fd) }
func (r *linuxReader) Root(ctx context.Context) (rootObservation, error) {
	if ctx.Err() != nil {
		return rootObservation{}, ErrUnavailable
	}
	var stat syscall.Stat_t
	var fs syscall.Statfs_t
	var attrs [28]byte
	if syscall.Fstat(r.fd, &stat) != nil || stat.Mode&syscall.S_IFMT != syscall.S_IFDIR || syscall.Fstatfs(r.fd, &fs) != nil {
		return rootObservation{}, ErrUnavailable
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(r.fd), 0x801c581f, uintptr(unsafe.Pointer(&attrs[0])))
	runtime.KeepAlive(&attrs)
	if errno != 0 || ctx.Err() != nil {
		return rootObservation{}, ErrUnavailable
	}
	return rootObservation{stat.Dev, stat.Ino, fs.Type, binary.LittleEndian.Uint32(attrs[12:16]), binary.LittleEndian.Uint32(attrs[:4])}, nil
}
func (r *linuxReader) query(ctx context.Context, command uintptr, project uint32, raw []byte) error {
	if ctx.Err() != nil {
		return ErrUnavailable
	}
	_, _, errno := syscall.Syscall6(443, uintptr(r.fd), command, uintptr(project), uintptr(unsafe.Pointer(&raw[0])), 0, 0)
	runtime.KeepAlive(raw)
	if errno != 0 || ctx.Err() != nil {
		return ErrUnavailable
	}
	return nil
}
func (r *linuxReader) Status(ctx context.Context) ([]byte, error) {
	raw := make([]byte, statusBytes)
	raw[0] = 1
	if r.query(ctx, 0x580802, 0, raw) != nil {
		return nil, ErrUnavailable
	}
	return raw, nil
}
func (r *linuxReader) Quota(ctx context.Context, project uint32) ([]byte, error) {
	raw := make([]byte, quotaBytes)
	raw[0] = 1
	if r.query(ctx, 0x580302, project, raw) != nil {
		return nil, ErrUnavailable
	}
	return raw, nil
}
