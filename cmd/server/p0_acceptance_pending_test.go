package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// This is test-only vocabulary, not a proposed public API or production state
// enum. Adapters normalize observations from real application entry points,
// mocked persistence/providers and fake clocks. Never return the Want fields or
// use real providers/credentials/database connections to satisfy these contracts.
type pendingP0Case struct {
	ID           string               `json:"id"`
	Area         string               `json:"area"`
	Requirements []string             `json:"requirements"`
	Given        map[string]string    `json:"given"`
	Action       string               `json:"action"`
	Want         pendingP0Expectation `json:"want"`
}
type pendingP0Expectation struct {
	State     string            `json:"state"`
	Effects   []string          `json:"effects"`
	Forbidden []string          `json:"forbidden"`
	Values    map[string]string `json:"values"`
}
type pendingP0Observation struct {
	State string
	// Ordered security-relevant effects, including denied/failed attempts. Do
	// not filter away unexpected effects merely to match the expected sequence.
	Effects []string
	Values  map[string]string
}
type pendingP0Driver func(*testing.T, pendingP0Case) pendingP0Observation

// Register each scenario's production-backed hermetic driver here as it lands.
// Leaving an absent implementation pending is intentional: there is no canned
// passing driver. A strict release-readiness run fails every unbound scenario.
var pendingP0Drivers = map[string]pendingP0Driver{
	"discard-missing":              driveP0DisposalRefusal,
	"discard-wrong-actor":          driveP0DisposalRefusal,
	"discard-wrong-volume":         driveP0DisposalRefusal,
	"storage-exact-boundary":       driveP0StorageReservation,
	"storage-one-byte-over":        driveP0StorageReservation,
	"storage-concurrent-last-byte": driveP0StorageReservation,
	"storage-count-retained":       driveP0StorageReservation,
	"storage-negative":             driveP0StorageReservation,
	"storage-no-paid-bypass":       driveP0StorageReservation,
	"refresh-atomic":               driveP0CredentialRefresh,
	"refresh-concurrent":           driveP0CredentialRefresh,
	"refresh-persist-fails":        driveP0CredentialRefresh,
	"refresh-foreign-connection":   driveP0CredentialRefresh,
}

func loadPendingP0Cases(t *testing.T) []pendingP0Case {
	t.Helper()
	raw, err := os.ReadFile("testdata/p0_remaining_acceptance.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []pendingP0Case
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&cases); err != nil {
		t.Fatal(err)
	}
	if decoder.Decode(new(any)) != io.EOF {
		t.Fatal("trailing acceptance catalog data")
	}
	return cases
}
func TestP0RemainingAcceptance(t *testing.T) {
	strict := os.Getenv("COMMONROOM_REQUIRE_P0_ACCEPTANCE") == "1"
	for _, scenario := range loadPendingP0Cases(t) {
		t.Run(scenario.Area+"/"+scenario.ID, func(t *testing.T) {
			driver := pendingP0Drivers[scenario.ID]
			if driver == nil {
				reason := fmt.Sprintf("PENDING %s (%s): no production-backed hermetic adapter; %s", scenario.ID, strings.Join(scenario.Requirements, ", "), scenario.Action)
				if strict {
					t.Fatal(reason)
				}
				t.Skip(reason)
			}
			actual := driver(t, scenario)
			for _, failure := range pendingP0Mismatches(scenario.Want, actual) {
				t.Error(failure)
			}
		})
	}
}
func pendingP0Mismatches(want pendingP0Expectation, actual pendingP0Observation) []string {
	var failures []string
	if actual.State != want.State {
		failures = append(failures, fmt.Sprintf("state: got %q, want %q", actual.State, want.State))
	}
	if !reflect.DeepEqual(append([]string{}, actual.Effects...), append([]string{}, want.Effects...)) {
		failures = append(failures, fmt.Sprintf("ordered effects: got %v, want %v", actual.Effects, want.Effects))
	}
	forbidden := append([]string{"host_execution", "credential_in_audit", "history_mutated", "reused_container", "silent_api_fallback"}, want.Forbidden...)
	for _, effect := range actual.Effects {
		for _, bad := range forbidden {
			if effect == bad {
				failures = append(failures, "forbidden effect: "+bad)
			}
		}
	}
	for key, value := range want.Values {
		if got, ok := actual.Values[key]; !ok || got != value {
			failures = append(failures, fmt.Sprintf("observation %s: got %q (present=%v), want %q", key, got, ok, value))
		}
	}
	return failures
}
func TestP0AcceptanceCatalog(t *testing.T) {
	cases := loadPendingP0Cases(t)
	spec, err := os.ReadFile("../../FUNCTIONAL_SPEC.md")
	if err != nil {
		t.Fatal(err)
	}
	references := map[string]bool{}
	for _, match := range regexp.MustCompile(`(?m)^(FR-[0-9]+):`).FindAllStringSubmatch(string(spec), -1) {
		references[match[1]] = true
	}
	ids := map[string]bool{}
	areas := map[string]map[string]bool{}
	for _, c := range cases {
		if c.ID == "" || ids[c.ID] || c.Area == "" || c.Action == "" || len(c.Given) == 0 || len(c.Requirements) == 0 {
			t.Fatalf("missing or duplicate scenario contract: %s", c.ID)
		}
		ids[c.ID] = true
		if c.Want.State != "allow" && c.Want.State != "deny" && c.Want.State != "block" {
			t.Fatalf("unknown expected state: %s", c.ID)
		}
		if areas[c.Area] == nil {
			areas[c.Area] = map[string]bool{}
		}
		areas[c.Area][c.Want.State] = true
		for _, ref := range c.Requirements {
			if !references[ref] {
				t.Errorf("%s references absent requirement %s", c.ID, ref)
			}
		}
		for _, effect := range c.Want.Effects {
			for _, bad := range c.Want.Forbidden {
				if effect == bad {
					t.Errorf("%s both requires and forbids %s", c.ID, effect)
				}
			}
		}
		if len(pendingP0Mismatches(c.Want, pendingP0Observation{State: c.Want.State, Effects: c.Want.Effects, Values: c.Want.Values})) != 0 {
			t.Errorf("internally inconsistent expectations: %s", c.ID)
		}
	}
	for _, area := range []string{"admission", "providers", "credentials", "storage", "upgrades", "idle", "commands", "artifacts", "disposal"} {
		states := areas[area]
		if !states["allow"] || !states["deny"] && !states["block"] {
			t.Errorf("%s lacks positive and refusal scenarios", area)
		}
	}
	for id := range pendingP0Drivers {
		if !ids[id] {
			t.Errorf("adapter has no scenario: %s", id)
		}
	}
}
func TestP0AcceptanceAssertionsRejectIncompleteOrUnsafeObservations(t *testing.T) {
	want := pendingP0Expectation{State: "allow", Effects: []string{"intent_durable", "isolated_command", "outcome_durable"}, Values: map[string]string{"revision": "r1"}}
	for _, actual := range []pendingP0Observation{
		{State: "allow"},
		{State: "deny", Effects: want.Effects, Values: want.Values},
		{State: "allow", Effects: []string{"isolated_command", "intent_durable", "outcome_durable"}, Values: want.Values},
		{State: "allow", Effects: append(append([]string{}, want.Effects...), "credential_in_audit"), Values: want.Values},
		{State: "allow", Effects: want.Effects, Values: map[string]string{"revision": "r2"}},
	} {
		if len(pendingP0Mismatches(want, actual)) == 0 {
			t.Fatal("unsafe/incomplete evidence accepted")
		}
	}
}
