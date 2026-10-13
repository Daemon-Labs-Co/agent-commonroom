package main

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// workerRecordsIdle rejects unknown/nonterminal records instead of inferring
// that a missing container proves a suspended worker cannot launch it later.
func workerRecordsIdle(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()
	entries, err := f.ReadDir(4097)
	if err != nil && err != io.EOF {
		return false, err
	}
	if len(entries) > 4096 {
		return false, errors.New("worker record scan limit exceeded")
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		job := strings.TrimSuffix(entry.Name(), ".json")
		if !validWorkerJobID(job) {
			return false, errors.New("invalid worker record name")
		}
		info, e := os.Lstat(filepath.Join(dir, entry.Name()))
		if e != nil || !info.Mode().IsRegular() {
			return false, errors.New("invalid worker record file")
		}
		record, e := readWorkerRecord(dir, job)
		if e != nil {
			return false, e
		}
		if !workerJobSettled(record) {
			return false, nil
		}
	}
	return true, nil
}

func workerIdleProbe(ctx context.Context) (bool, error) {
	if err := configureWorker(); err != nil {
		return false, err
	}
	var info limitedAgentOutput
	cmd := exec.CommandContext(ctx, "docker", "info", "--format", "{{json .}}")
	cmd.Stdout = &info
	if cmd.Run() != nil || !dockerSecurityReady(info.Bytes()) {
		return false, errors.New("rootless Docker unavailable")
	}
	var ids limitedAgentOutput
	cmd = exec.CommandContext(ctx, "docker", "ps", "-a", "--filter", "label=commonroom.managed=true", "--format", "{{.ID}}")
	cmd.Stdout = &ids
	if cmd.Run() != nil {
		return false, errors.New("container state unavailable")
	}
	if strings.TrimSpace(ids.String()) != "" {
		return false, nil
	}
	dir, err := workerStateDirectory()
	if err != nil {
		return false, err
	}
	snapshot, err := workerSnapshot(dir, nil)
	if err != nil {
		return false, err
	}
	return snapshot.Unresolved == 0, nil

}
