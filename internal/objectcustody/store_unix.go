//go:build linux || darwin

// Package objectcustody stores bounded immutable bytes behind mandatory trusted
// accounting and audit callbacks. It is an internal foundation, not a provider,
// task-output/export endpoint or proof of artifact origin/format. Production
// admission remains blocked until real quota, isolated-source and journal
// adapters exist. A source fork CAN supply dishonest callbacks or a host root,
// invalidating storage/accounting/isolation; no supported fallback permits that.
package objectcustody

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"
	"syscall"
	"unicode/utf8"
)

const MaxObjectBytes int64 = 16 * 1024 * 1024

var (
	ErrRefused     = errors.New("object custody refused")
	ErrUnconfirmed = errors.New("object custody unconfirmed; retain charge and bytes")
	identifier     = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)
)

// Scope is checked against the authenticated principal and actual isolated job
// by Journal. These identifiers/hashes are evidence bindings, never authority.
type ExecutionScope struct {
	CompanyID        string `json:"companyId"`
	RoomID           string `json:"roomId"`
	TaskID           string `json:"taskId"`
	ExecutionID      string `json:"executionId"`
	JobID            string `json:"jobId"`
	RepositoryDigest string `json:"repositoryDigest"`
	RevisionDigest   string `json:"revisionDigest"`
	DefinitionDigest string `json:"definitionDigest"`
}

// Scope adds immutable output/member labels. Labels never become filesystem
// paths here; the trusted source adapter must verify their actual origin.
type Scope struct {
	ExecutionScope
	OutputName string `json:"outputName"`
	MemberName string `json:"memberName"`
}
type Intent struct {
	ObjectID string `json:"objectId"`
	Scope    Scope  `json:"scope"`
	MaxBytes int64  `json:"maxBytes"`
}
type Receipt struct {
	Intent Intent `json:"intent"`
	Bytes  int64  `json:"bytes"`
	Digest string `json:"digest"`
}

// Journal is a mandatory TRUSTED controller/worker integration boundary, never
// a model/plugin callback. Reserve must durably charge maximum capacity across
// retained/staged/uncertain account data AND establish an actual physical quota
// before any bytes are read/created. Seal must revalidate current job, source
// revision, policy and authority and durably acknowledge the exact content.
// AuthorizeRead requires an exact existing sealed receipt plus current room,
// principal, revision and export policy. AcknowledgeRead must revalidate current
// authority again and durably record bytes authorized for release, not proof
// that a remote client received them.
// A missing, no-op or synthetic implementation cannot admit production work.
// There is deliberately no automatic refund, deletion, adoption or retry API.
type Journal interface {
	Reserve(context.Context, Intent) error
	Seal(context.Context, Receipt) error
	AuthorizeRead(context.Context, Scope, Receipt) error
	AcknowledgeRead(context.Context, Receipt) error
}

// Store borrows a descriptor-pinned, private, quota-backed object root. Trusted
// setup must keep it open and prevent all other writers/mount/device replacement.
// It must never be the engine credential volume or a model-selected/host worktree.
// os.Root is path containment, not kernel/container or physical-quota proof.
type Store struct {
	root          *os.Root
	journal       Journal
	syncFile      func(*os.File) error
	syncDirectory func(*os.File) error
}

func New(root *os.Root, journal Journal) (*Store, error) {
	if root == nil || absent(journal) {
		return nil, ErrRefused
	}
	store := &Store{root: root, journal: journal, syncFile: func(f *os.File) error { return f.Sync() }, syncDirectory: func(f *os.File) error { return f.Sync() }}
	dir, err := store.directory()
	if err != nil {
		return nil, err
	}
	if dir.Close() != nil {
		return nil, ErrRefused
	}
	return store, nil
}
func absent(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	}
	return false
}
func active(ctx context.Context) bool { return ctx != nil && ctx.Err() == nil }
func lowerDigest(value string, size int) bool {
	raw, err := hex.DecodeString(value)
	return err == nil && len(raw) == size && hex.EncodeToString(raw) == value
}

// ValidExecutionScope validates metadata shape only, never admission/authority.
func ValidExecutionScope(s ExecutionScope) bool {
	return identifier.MatchString(s.CompanyID) && identifier.MatchString(s.RoomID) && identifier.MatchString(s.TaskID) && lowerDigest(s.ExecutionID, 16) && lowerDigest(s.JobID, 16) && lowerDigest(s.RepositoryDigest, 32) && lowerDigest(s.RevisionDigest, 32) && lowerDigest(s.DefinitionDigest, 32)
}
func validScope(s Scope) bool {
	name := s.MemberName
	return ValidExecutionScope(s.ExecutionScope) && identifier.MatchString(s.OutputName) && utf8.ValidString(name) && strings.TrimSpace(name) != "" && utf8.RuneCountInString(name) <= 200 && !strings.ContainsAny(name, "\\\x00\r\n:") && !path.IsAbs(name) && name != "." && name != ".." && !strings.HasPrefix(name, "../") && path.Clean(name) == name
}
func validIntent(v Intent) bool {
	return lowerDigest(v.ObjectID, 16) && validScope(v.Scope) && v.MaxBytes > 0 && v.MaxBytes <= MaxObjectBytes
}
func validReceipt(v Receipt) bool {
	return validIntent(v.Intent) && v.Bytes >= 0 && v.Bytes <= v.Intent.MaxBytes && lowerDigest(v.Digest, 32)
}
func safeDirectory(info os.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && info.IsDir() && info.Mode().Perm() == 0700 && info.Mode()&(os.ModeSetuid|os.ModeSetgid|os.ModeSticky) == 0 && stat.Uid == uint32(os.Geteuid())
}
func safeObject(info os.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && info.Mode().IsRegular() && info.Mode().Perm() == 0600 && info.Mode()&(os.ModeSetuid|os.ModeSetgid|os.ModeSticky) == 0 && stat.Nlink == 1 && stat.Uid == uint32(os.Geteuid())
}
func (s *Store) directory() (*os.File, error) {
	if s == nil || s.root == nil || s.journal == nil {
		return nil, ErrRefused
	}
	dir, err := s.root.OpenFile(".", os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, ErrRefused
	}
	info, err := dir.Stat()
	if err != nil || !safeDirectory(info) {
		dir.Close()
		return nil, ErrRefused
	}
	return dir, nil
}
func objectName(id string) string { return "object-" + id + ".blob" }

// Put reads actual bounded bytes only after accounting/admission acknowledgment.
// The source must be a context-aware isolated tool export, never engine output,
// a model path/claim, credentials or an unverified live mutable source. This
// primitive proves stored byte identity only; source quiescence, content/format
// policy and revision attestation remain mandatory caller/journal obligations.
func (s *Store) Put(ctx context.Context, scope Scope, maxBytes int64, source io.Reader) (Receipt, error) {
	if !active(ctx) || !validScope(scope) || maxBytes < 1 || maxBytes > MaxObjectBytes || absent(source) {
		return Receipt{}, ErrRefused
	}
	dir, err := s.directory()
	if err != nil {
		return Receipt{}, ErrRefused
	}
	defer dir.Close()
	var nonce [16]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return Receipt{}, ErrRefused
	}
	intent := Intent{ObjectID: hex.EncodeToString(nonce[:]), Scope: scope, MaxBytes: maxBytes}
	if s.journal.Reserve(ctx, intent) != nil || !active(ctx) {
		return Receipt{}, ErrRefused
	}
	// A private unsealed file is never discoverable through the public read path.
	// O_EXCL and random names prevent overwrite; no cleanup/refund is inferred
	// after reservation, even if creation, streaming, sync or acknowledgment fails.
	file, err := s.root.OpenFile(objectName(intent.ObjectID), os.O_RDWR|os.O_CREATE|os.O_EXCL|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0600)
	if err != nil {
		return Receipt{}, ErrUnconfirmed
	}
	closed := false
	defer func() {
		if !closed {
			file.Close()
		}
	}()
	hash := sha256.New()
	buffer := make([]byte, 32*1024)
	total := int64(0)
	emptyReads := 0
	for {
		if !active(ctx) {
			return Receipt{}, ErrUnconfirmed
		}
		available := maxBytes - total + 1
		if available > int64(len(buffer)) {
			available = int64(len(buffer))
		}
		n, readErr := source.Read(buffer[:int(available)])
		if n < 0 || n > int(available) || !active(ctx) {
			return Receipt{}, ErrUnconfirmed
		}
		if n > 0 {
			emptyReads = 0
			if int64(n) > maxBytes-total {
				return Receipt{}, ErrUnconfirmed
			}
			if written, e := file.Write(buffer[:n]); e != nil || written != n {
				return Receipt{}, ErrUnconfirmed
			}
			_, _ = hash.Write(buffer[:n])
			total += int64(n)
		} else if readErr == nil {
			emptyReads++
			if emptyReads >= 16 {
				return Receipt{}, ErrUnconfirmed
			}
		}
		if readErr != nil {
			if !errors.Is(readErr, io.EOF) {
				return Receipt{}, ErrUnconfirmed
			}
			break
		}
	}
	info, err := file.Stat()
	if err != nil || !safeObject(info) || info.Size() != total || s.syncFile(file) != nil || !active(ctx) {
		return Receipt{}, ErrUnconfirmed
	}
	// Verify persisted inode bytes, not merely the buffer passed to Write.
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return Receipt{}, ErrUnconfirmed
	}
	observed := sha256.New()
	count, err := io.Copy(observed, io.LimitReader(file, total+1))
	if err != nil || count != total || hex.EncodeToString(observed.Sum(nil)) != hex.EncodeToString(hash.Sum(nil)) || !active(ctx) {
		return Receipt{}, ErrUnconfirmed
	}
	if file.Close() != nil {
		return Receipt{}, ErrUnconfirmed
	}
	closed = true
	current, err := s.root.Lstat(objectName(intent.ObjectID))
	if err != nil || !safeObject(current) || !os.SameFile(info, current) || current.Size() != total || s.syncDirectory(dir) != nil || !active(ctx) {
		return Receipt{}, ErrUnconfirmed
	}
	receipt := Receipt{Intent: intent, Bytes: total, Digest: hex.EncodeToString(hash.Sum(nil))}
	if s.journal.Seal(ctx, receipt) != nil || !active(ctx) {
		return Receipt{}, ErrUnconfirmed
	}
	return receipt, nil
}

// Get verifies the actual private inode, size and content BEFORE acknowledging
// release. No raw bytes are returned on any authorization, integrity, cancellation
// or audit failure. It never accepts a model-provided filesystem path or URI.
func (s *Store) Get(ctx context.Context, scope Scope, receipt Receipt) ([]byte, error) {
	if !active(ctx) || !validReceipt(receipt) || scope != receipt.Intent.Scope {
		return nil, ErrRefused
	}
	dir, err := s.directory()
	if err != nil {
		return nil, ErrRefused
	}
	defer dir.Close()
	if s.journal.AuthorizeRead(ctx, scope, receipt) != nil || !active(ctx) {
		return nil, ErrRefused
	}
	file, err := s.root.OpenFile(objectName(receipt.Intent.ObjectID), os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, ErrRefused
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !safeObject(info) || info.Size() != receipt.Bytes {
		return nil, ErrRefused
	}
	raw, err := io.ReadAll(io.LimitReader(file, receipt.Bytes+1))
	if err != nil || int64(len(raw)) != receipt.Bytes || !active(ctx) {
		return nil, ErrRefused
	}
	digest := sha256.Sum256(raw)
	if hex.EncodeToString(digest[:]) != receipt.Digest {
		return nil, ErrRefused
	}
	after, err := file.Stat()
	current, e := s.root.Lstat(objectName(receipt.Intent.ObjectID))
	if err != nil || e != nil || !safeObject(after) || !safeObject(current) || !os.SameFile(info, after) || !os.SameFile(info, current) || after.Size() != receipt.Bytes || current.Size() != receipt.Bytes {
		return nil, ErrRefused
	}
	if s.journal.AcknowledgeRead(ctx, receipt) != nil || !active(ctx) {
		return nil, ErrRefused
	}
	return raw, nil
}
