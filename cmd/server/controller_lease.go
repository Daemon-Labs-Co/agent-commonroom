package main

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
)

type controllerLease struct{ file *os.File }

// acquireControllerLease enforces one controller per trusted local data directory
// before database/identity/worker startup. Never unlink the lease while a process
// can hold it: that would create a second lock inode. This is not a distributed
// database lease; separate directories/hosts must not share the same installation.
func acquireControllerLease(path string) (*controllerLease, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	f, err := openPrivateRegular(path, os.O_RDWR|os.O_CREATE, 128)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		if err == syscall.EWOULDBLOCK || err == syscall.EAGAIN {
			return nil, errors.New("Another controller holds this installation's data directory; stop it before starting a replacement")
		}
		return nil, err
	}
	return &controllerLease{file: f}, nil
}

func (lease *controllerLease) Close() {
	if lease == nil || lease.file == nil {
		return
	}
	_ = syscall.Flock(int(lease.file.Fd()), syscall.LOCK_UN)
	_ = lease.file.Close()
	lease.file = nil
}
