package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
)

type localConfiguration struct {
	DatabaseURL         string `json:"databaseUrl"`
	RemoteWorkerAlias   string `json:"remoteWorkerAlias"`
	RemoteWorkerEnabled bool   `json:"remoteWorkerEnabled"`
}

// loadLocalConfiguration reads a bounded private descriptor, never an unchecked
// path. Invalid files stop startup; absence preserves environment/default setup.
// Decode errors are intentionally generic so credential-containing input is not
// copied into startup logs. Parent directories remain operator-controlled.
func loadLocalConfiguration(path string) (localConfiguration, error) {
	var config localConfiguration
	raw, err := readPrivateBounded(path, 8192)
	if os.IsNotExist(err) {
		return config, nil
	}
	if err != nil {
		return config, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' || d.Decode(&config) != nil || d.Decode(new(any)) != io.EOF {
		return localConfiguration{}, errors.New("Invalid local configuration")
	}
	return config, nil
}
