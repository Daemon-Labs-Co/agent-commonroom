package main

import (
	"os"
	"reflect"
	"regexp"
	"sort"
	"testing"
)

// Every event kind the original replay handled has exactly one reducer, and no reducer
// exists for a kind the original ignored. The original is the frozen referenceLoad.
func TestEveryOriginalEventKindHasAReducer(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("projection_reference_test.go")
	if err != nil {
		t.Fatal(err)
	}
	original := map[string]bool{}
	for _, line := range regexp.MustCompile(`(?m)^\t\tcase ((?:"[^"]+"(?:, )?)+):$`).FindAllStringSubmatch(string(source), -1) {
		for _, kind := range regexp.MustCompile(`"([^"]+)"`).FindAllStringSubmatch(line[1], -1) {
			original[kind[1]] = true
		}
	}
	if len(original) < 80 {
		t.Fatalf("found only %d kinds in the reference; the pattern no longer matches it", len(original))
	}
	var missing, extra []string
	for kind := range original {
		if projectionReducers[kind] == nil {
			missing = append(missing, kind)
		}
	}
	for kind := range projectionReducers {
		if !original[kind] {
			extra = append(extra, kind)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	if len(missing) > 0 || len(extra) > 0 {
		t.Fatalf("missing reducers %v; reducers for kinds the original ignored %v", missing, extra)
	}
}

func TestUnknownEventKindChangesNothingButTheCount(t *testing.T) {
	t.Parallel()
	p := newProjection()
	before := p.st.Clone()
	if err := p.apply("something.never.recorded", []byte(`{"a":1}`)); err != nil {
		t.Fatal(err)
	}
	before.EventCount++
	if !equalStates(before, p.st) {
		t.Fatal("an unknown kind changed the state")
	}
}

func equalStates(a, b interface{}) bool { return reflect.DeepEqual(a, b) }
