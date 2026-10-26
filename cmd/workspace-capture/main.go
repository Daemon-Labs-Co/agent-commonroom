// workspace-capture runs only in the reviewed credential-free capture container.
// Its stdio belongs to the supervisor, never a model or user terminal. There are
// no configurable source/object paths, provider commands or host fallbacks.
package main

import (
	"agentroom/internal/objectcustody"
	"agentroom/internal/workspacecapture"
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

const maxEncodedRequest = 64 * 1024

type captureFunc func(context.Context, workspacecapture.Request, workspacecapture.Authority, *os.Root, objectcustody.Journal) (workspacecapture.CapturedObjects, error)

func decodeRequest(args []string) (workspacecapture.Request, error) {
	var request workspacecapture.Request
	if len(args) != 1 || len(args[0]) == 0 || len(args[0]) > maxEncodedRequest {
		return request, workspacecapture.ErrRefused
	}
	raw, err := base64.RawURLEncoding.DecodeString(args[0])
	if err != nil || base64.RawURLEncoding.EncodeToString(raw) != args[0] || json.Unmarshal(raw, &request) != nil {
		return request, workspacecapture.ErrRefused
	}
	canonical, err := json.Marshal(request)
	if err != nil || !bytes.Equal(raw, canonical) {
		return workspacecapture.Request{}, workspacecapture.ErrRefused
	}
	if _, err := workspacecapture.RequestDigest(request); err != nil {
		return workspacecapture.Request{}, workspacecapture.ErrRefused
	}
	return request, nil
}
func run(ctx context.Context, args []string, stream io.ReadWriteCloser, openObjects func() (*os.Root, error), capture captureFunc) error {
	if ctx == nil || ctx.Err() != nil || stream == nil || openObjects == nil || capture == nil {
		return workspacecapture.ErrRefused
	}
	defer stream.Close()
	request, err := decodeRequest(args)
	if err != nil {
		return err
	}
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	channel, err := workspacecapture.NewCaptureChannel(child, stream, request, cancel)
	if err != nil {
		return err
	}
	defer channel.Close()
	root, err := openObjects()
	if err != nil || root == nil {
		return workspacecapture.ErrRefused
	}
	defer root.Close()
	result, err := capture(child, request, channel, root, channel)
	if err != nil {
		return err
	}
	return channel.Finish(child, result)
}

type supervisorStdio struct{ in, out *os.File }

func (s *supervisorStdio) Read(p []byte) (int, error)  { return s.in.Read(p) }
func (s *supervisorStdio) Write(p []byte) (int, error) { return s.out.Write(p) }
func (s *supervisorStdio) Close() error                { return errors.Join(s.in.Close(), s.out.Close()) }
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 35*time.Second)
	defer cancel()
	err := run(ctx, os.Args[1:], &supervisorStdio{os.Stdin, os.Stdout}, func() (*os.Root, error) { return os.OpenRoot("/objects") }, workspacecapture.CaptureToCustody)
	if err != nil {
		fmt.Fprintln(os.Stderr, workspacecapture.ErrRefused)
		os.Exit(2)
	}
}
