package main

import (
	"agentroom/internal/objectcustody"
	"agentroom/internal/objectread"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func testRequest() objectread.Request {
	return objectread.Request{Receipt: objectcustody.Receipt{Intent: objectcustody.Intent{ObjectID: strings.Repeat("1", 32), Scope: objectcustody.Scope{ExecutionScope: objectcustody.ExecutionScope{CompanyID: "company", RoomID: "room", TaskID: "task", ExecutionID: strings.Repeat("2", 32), JobID: strings.Repeat("3", 32), RepositoryDigest: strings.Repeat("4", 64), RevisionDigest: strings.Repeat("5", 64), DefinitionDigest: strings.Repeat("6", 64)}, OutputName: "report", MemberName: "report.txt"}, MaxBytes: 128}, Bytes: 3, Digest: strings.Repeat("7", 64)}, Challenge: strings.Repeat("8", 32)}
}

type testStream struct {
	bytes.Buffer
	closed bool
}

func (s *testStream) Close() error { s.closed = true; return nil }
func TestReaderCommandFiniteCanonicalRequest(t *testing.T) {
	t.Parallel()
	q := testRequest()
	raw, _ := json.Marshal(q)
	encoded := base64.RawURLEncoding.EncodeToString(raw)
	if got, err := decodeRequest([]string{encoded}); err != nil || got != q {
		t.Fatal(got, err)
	}
	for _, args := range [][]string{nil, {encoded, encoded}, {"/home/private"}, {encoded + "="}, {strings.Repeat("x", 8193)}, {base64.RawURLEncoding.EncodeToString(append(raw, ' '))}, {base64.RawURLEncoding.EncodeToString([]byte(`{"receipt":{},"challenge":"bad","path":"/home"}`))}} {
		if _, err := decodeRequest(args); err == nil {
			t.Fatal("invalid args accepted")
		}
	}
	stream := &testStream{}
	called := false
	err := run(t.Context(), []string{encoded}, stream, func(_ context.Context, request objectread.Request, _ io.ReadWriteCloser) error {
		called = true
		if request != q {
			t.Fatal("binding changed")
		}
		return nil
	})
	if err != nil || !called || !stream.closed {
		t.Fatal(err, called, stream.closed)
	}
}
