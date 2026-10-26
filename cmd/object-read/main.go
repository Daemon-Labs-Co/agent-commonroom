// object-read runs only in the reviewed credential-free reader container. Its
// stdio belongs exclusively to the supervisor, never a model/user terminal.
// No paths, provider commands, database credentials or host fallback are accepted.
package main

import (
	"agentroom/internal/objectcustody"
	"agentroom/internal/objectread"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func decodeRequest(args []string) (objectread.Request, error) {
	var q objectread.Request
	if len(args) != 1 || len(args[0]) == 0 || len(args[0]) > 8192 {
		return q, objectcustody.ErrRefused
	}
	raw, err := base64.RawURLEncoding.DecodeString(args[0])
	if err != nil || base64.RawURLEncoding.EncodeToString(raw) != args[0] || json.Unmarshal(raw, &q) != nil {
		return objectread.Request{}, objectcustody.ErrRefused
	}
	canonical, err := json.Marshal(q)
	if err != nil || !bytes.Equal(raw, canonical) || !objectread.ValidRequest(q) {
		return objectread.Request{}, objectcustody.ErrRefused
	}
	return q, nil
}
func run(ctx context.Context, args []string, stream io.ReadWriteCloser, serve func(context.Context, objectread.Request, io.ReadWriteCloser) error) error {
	if ctx == nil || ctx.Err() != nil || stream == nil || serve == nil {
		return objectcustody.ErrRefused
	}
	defer stream.Close()
	q, err := decodeRequest(args)
	if err != nil {
		return err
	}
	return serve(ctx, q, stream)
}

type supervisorStdio struct{ in, out *os.File }

func (s *supervisorStdio) Read(b []byte) (int, error)  { return s.in.Read(b) }
func (s *supervisorStdio) Write(b []byte) (int, error) { return s.out.Write(b) }
func (s *supervisorStdio) Close() error                { return errors.Join(s.in.Close(), s.out.Close()) }
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 35*time.Second)
	defer cancel()
	if run(ctx, os.Args[1:], &supervisorStdio{os.Stdin, os.Stdout}, objectread.Serve) != nil {
		fmt.Fprintln(os.Stderr, objectcustody.ErrRefused)
		os.Exit(2)
	}
}
