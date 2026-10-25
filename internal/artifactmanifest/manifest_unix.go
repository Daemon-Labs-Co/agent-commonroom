//go:build linux || darwin

// Package artifactmanifest binds finite file/bundle contracts to actual sealed
// bytes in objectcustody. It is not connected to managed dispatch or public
// delivery. Trusted isolated-source, quota and journal adapters remain required;
// model names, receipts and format checks are never proof of origin/authority.
package artifactmanifest

import (
	"agentroom/internal/objectcustody"
	"agentroom/internal/taskcontract"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"mime"
	"reflect"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

var ErrRefused = errors.New("artifact manifest validation refused")

const Version = 1

type Claim struct {
	Name    string                `json:"name"`
	Receipt objectcustody.Receipt `json:"receipt"`
}
type File struct {
	Name            string                `json:"name"`
	MediaType       string                `json:"mediaType"`
	StructuralCheck string                `json:"structuralCheck"`
	Receipt         objectcustody.Receipt `json:"receipt"`
}

// Manifest contains immutable identities, not bytes, host paths, capabilities or
// observed build/test results. Its digest is not a substitute for revalidation.
type Manifest struct {
	Version        int                          `json:"version"`
	Scope          objectcustody.ExecutionScope `json:"scope"`
	OutputName     string                       `json:"outputName"`
	OutputType     string                       `json:"outputType"`
	ContractDigest string                       `json:"contractDigest"`
	Files          []File                       `json:"files"`
	TotalBytes     int64                        `json:"totalBytes"`
	Digest         string                       `json:"digest"`
}
type format struct{ media, check string }

func registeredFormat(value string) (format, bool) {
	media, params, err := mime.ParseMediaType(value)
	if err != nil {
		return format{}, false
	}
	check := ""
	switch media {
	case "application/octet-stream":
		check = "bounded-opaque-v1"
		if len(params) != 0 {
			return format{}, false
		}
	case "text/plain", "text/markdown":
		check = "utf8-text-v1"
	case "application/json":
		check = "strict-json-v1"
	default:
		return format{}, false
	}
	for key, value := range params {
		if key != "charset" || !strings.EqualFold(value, "utf-8") {
			return format{}, false
		}
		params[key] = "utf-8"
	}
	return format{media: mime.FormatMediaType(media, params), check: check}, true
}
func validFormatBytes(f format, raw []byte) bool {
	switch f.check {
	case "bounded-opaque-v1":
		return true
	case "utf8-text-v1":
		if !utf8.Valid(raw) {
			return false
		}
		for _, r := range string(raw) {
			if unicode.IsControl(r) && r != '\r' && r != '\n' && r != '\t' {
				return false
			}
		}
		return true
	case "strict-json-v1":
		return taskcontract.ValidJSON(raw, len(raw))
	}
	return false
}
func digest(value any) string {
	raw, _ := json.Marshal(value)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
func manifestDigest(value Manifest) string { value.Digest = ""; return digest(value) }

// Resolve preflights the entire finite claim set BEFORE reading any object.
// Each actual read repeats Store authority/integrity/audit checks. No partial
// manifest is returned, and retained objects/reservations are never refunded or
// deleted when format/member validation fails. The source owns cleanup/retention.
func Resolve(ctx context.Context, definition taskcontract.Definition, scope objectcustody.ExecutionScope, outputName string, claims []Claim, store *objectcustody.Store) (Manifest, error) {
	if ctx == nil || ctx.Err() != nil || store == nil || taskcontract.Validate(definition) != nil || !objectcustody.ValidExecutionScope(scope) || scope.DefinitionDigest != taskcontract.Digest(definition) {
		return Manifest{}, ErrRefused
	}
	var contract taskcontract.DataContract
	for _, c := range definition.DesiredOutputs {
		if c.Name == outputName {
			contract = c
			break
		}
	}
	declared := []taskcontract.BundleMember{}
	switch contract.Type {
	case "file":
		declared = append(declared, taskcontract.BundleMember{Name: contract.FileName, MediaType: contract.MediaType, MaxBytes: contract.MaxBytes})
	case "bundle":
		declared = contract.Members
	default:
		return Manifest{}, ErrRefused
	}
	if len(claims) != len(declared) || len(claims) < 1 || len(claims) > 32 {
		return Manifest{}, ErrRefused
	}
	declarations := map[string]taskcontract.BundleMember{}
	formats := map[string]format{}
	for _, member := range declared {
		f, ok := registeredFormat(member.MediaType)
		if !ok {
			return Manifest{}, ErrRefused
		}
		declarations[member.Name] = member
		formats[member.Name] = f
	}
	seenNames, seenObjects := map[string]bool{}, map[string]bool{}
	var total int64
	for _, claim := range claims {
		member, ok := declarations[claim.Name]
		receipt := claim.Receipt
		expected := objectcustody.Scope{ExecutionScope: scope, OutputName: outputName, MemberName: claim.Name}
		if !ok || seenNames[claim.Name] || seenObjects[receipt.Intent.ObjectID] || receipt.Intent.Scope != expected || receipt.Intent.MaxBytes != member.MaxBytes || receipt.Bytes < 0 || receipt.Bytes > member.MaxBytes || receipt.Bytes > contract.MaxBytes-total {
			return Manifest{}, ErrRefused
		}
		total += receipt.Bytes
		seenNames[claim.Name] = true
		seenObjects[receipt.Intent.ObjectID] = true
	}
	files := []File{}
	for _, claim := range claims {
		raw, err := store.Get(ctx, claim.Receipt.Intent.Scope, claim.Receipt)
		if err != nil || ctx.Err() != nil || !validFormatBytes(formats[claim.Name], raw) {
			return Manifest{}, ErrRefused
		}
		files = append(files, File{Name: claim.Name, MediaType: formats[claim.Name].media, StructuralCheck: formats[claim.Name].check, Receipt: claim.Receipt})
	}
	if ctx.Err() != nil {
		return Manifest{}, ErrRefused
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
	result := Manifest{Version: Version, Scope: scope, OutputName: outputName, OutputType: contract.Type, ContractDigest: digest(contract), Files: files, TotalBytes: total}
	result.Digest = manifestDigest(result)
	return result, nil
}

// Revalidate requires the exact current definition/execution/output scope and
// repeats actual byte checks. A digest-only model claim is never accepted.
// This does not authorize external delivery or certify observed runtime checks.
func Revalidate(ctx context.Context, definition taskcontract.Definition, scope objectcustody.ExecutionScope, outputName string, claimed Manifest, store *objectcustody.Store) error {
	if claimed.Version != Version || claimed.Scope != scope || claimed.OutputName != outputName || len(claimed.Files) < 1 || len(claimed.Files) > 32 || claimed.Digest != manifestDigest(claimed) {
		return ErrRefused
	}
	claims := make([]Claim, len(claimed.Files))
	for i, file := range claimed.Files {
		claims[i] = Claim{Name: file.Name, Receipt: file.Receipt}
	}
	actual, err := Resolve(ctx, definition, scope, outputName, claims, store)
	if err != nil || !reflect.DeepEqual(actual, claimed) {
		return ErrRefused
	}
	return nil
}
