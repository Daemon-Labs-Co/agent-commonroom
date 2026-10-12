package main

import (
	"errors"
	"net/url"
	"strconv"
	"time"
)

const auditPageSQL = "SELECT sequence,kind,committed_at,hash,payload FROM events WHERE ($1::bigint IS NULL OR sequence < $1) AND ($2::timestamptz IS NULL OR committed_at >= $2) AND ($3::timestamptz IS NULL OR committed_at < $3) ORDER BY sequence DESC LIMIT 101"

// auditPageBounds validates keyset pagination and half-open date windows. No
// caller input is interpolated into SQL; page size is fixed and bounded.
func auditPageBounds(query url.Values) ([]any, error) {
	var before any
	var dates [2]*time.Time
	if raw := query.Get("before"); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || n <= 0 {
			return nil, errors.New("Invalid history cursor")
		}
		before = n
	}
	for i, key := range []string{"from", "until"} {
		if raw := query.Get(key); raw != "" {
			value, err := time.Parse(time.RFC3339Nano, raw)
			if err != nil {
				return nil, errors.New("Invalid history date")
			}
			dates[i] = &value
		}
	}
	if dates[0] != nil && dates[1] != nil && !dates[0].Before(*dates[1]) {
		return nil, errors.New("Invalid history date range")
	}
	var from, until any
	if dates[0] != nil {
		from = *dates[0]
	}
	if dates[1] != nil {
		until = *dates[1]
	}
	return []any{before, from, until}, nil
}
