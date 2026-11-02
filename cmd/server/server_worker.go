package main

import (
	"errors"
	"os"
	"strings"
)

// commonroom-server is an internal worker identity, never an SSH host. Both transports use
// the same fenced lifecycle, audit channel and mandatory container implementation.
func serverWorkerBinding(endpoint string) (string, string, error) {
	switch endpoint {
	case "ssh://commonroom-server":
		return "commonroom-server", "codex", nil
	case "ssh://commonroom-server/claude":
		return "commonroom-server", "claude", nil
	}
	return "", "", errors.New("Invalid server worker provider")
}

// Historical binding validation does not grant permission to execute a worker.
func boundWorkerEndpoint(endpoint, alias string) (string, string, error) {
	if alias == "commonroom-server" {
		return serverWorkerBinding(endpoint)
	}
	return remoteWorkerEndpoint(endpoint, alias, "true")
}
func approvedWorkerEndpoint() string {
	if os.Getenv("COMMONROOM_SERVER_WORKER") == "true" {
		return "ssh://commonroom-server"
	}
	return "ssh://" + os.Getenv("COMMONROOM_REMOTE_WORKER_ALIAS")
}
func workerProcessEnvironment(env []string) []string {
	out := make([]string, 0, len(env))
	for _, v := range env {
		key, _, _ := strings.Cut(v, "=")
		// The worker has no database role and must not inherit controller credentials.
		if key == "DATABASE_URL" || key == "MIGRATION_DATABASE_URL" || strings.HasPrefix(key, "PG") || strings.HasPrefix(key, "POSTGRES_") {
			continue
		}
		out = append(out, v)
	}
	return out
}

func configuredWorkerIdentity() string {
	if os.Getenv("COMMONROOM_SERVER_WORKER") == "true" {
		return "commonroom-server"
	}
	return os.Getenv("COMMONROOM_REMOTE_WORKER_ALIAS")
}
