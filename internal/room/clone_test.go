package room

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestCloneIsDeepAndKeepsNilAndEmpty(t *testing.T) {
	t.Parallel()
	original := State{
		Messages:     []Message{{ID: "m1", Text: "hello"}},
		Rooms:        []Room{{ID: "r1", Name: "One"}},
		Tasks:        []Task{},
		Participants: nil,
		FeatureFlags: map[string]bool{"a": true},
		Incident:     true,
		EventCount:   7,
		ToolCalls:    map[string]ToolCallAudit{"c": {CallID: "c"}},
		Evidence:     []Evidence{{ID: "e1"}},
	}
	original.Rooms[0].ContextExport = &ContextExportPolicy{Providers: []string{"p"}}
	clone := original.Clone()
	if !reflect.DeepEqual(original, clone) {
		t.Fatal("clone differs from the original")
	}
	if clone.Tasks == nil || clone.Participants != nil {
		t.Fatal("nil and empty were not preserved")
	}
	clone.Messages[0].Text = "changed"
	clone.Rooms[0].Name = "changed"
	clone.Rooms[0].ContextExport.Providers[0] = "changed"
	clone.FeatureFlags["a"] = false
	clone.FeatureFlags["b"] = true
	clone.ToolCalls["c"] = ToolCallAudit{CallID: "other"}
	clone.Messages = append(clone.Messages, Message{ID: "m2"})
	if original.Messages[0].Text != "hello" || original.Rooms[0].Name != "One" || original.Rooms[0].ContextExport.Providers[0] != "p" ||
		!original.FeatureFlags["a"] || len(original.FeatureFlags) != 1 || original.ToolCalls["c"].CallID != "c" || len(original.Messages) != 1 {
		t.Fatalf("changing the clone changed the original: %+v", original)
	}
}

// Every field that holds a reference must be copied by Clone. A type with an unexported
// reference field would be shared silently, so the model may not contain one (time.Time is
// immutable and allowed).
func TestStateHasNoUnexportedReferenceFields(t *testing.T) {
	t.Parallel()
	seen := map[reflect.Type]bool{}
	var walk func(reflect.Type, string)
	walk = func(typ reflect.Type, path string) {
		if seen[typ] {
			return
		}
		seen[typ] = true
		switch typ.Kind() {
		case reflect.Slice, reflect.Array, reflect.Pointer:
			walk(typ.Elem(), path+"[]")
		case reflect.Map:
			walk(typ.Key(), path+"{key}")
			walk(typ.Elem(), path+"{}")
		case reflect.Struct:
			if typ == reflect.TypeOf(time.Time{}) {
				return
			}
			for i := 0; i < typ.NumField(); i++ {
				f := typ.Field(i)
				if !f.IsExported() {
					t.Errorf("%s.%s is unexported and would be shared by Clone", path, f.Name)
					continue
				}
				walk(f.Type, path+"."+f.Name)
			}
		case reflect.Func, reflect.Chan, reflect.UnsafePointer:
			t.Errorf("%s has kind %s, which Clone cannot copy", path, typ.Kind())
		}
	}
	walk(reflect.TypeOf(State{}), "State")
}

// A state that went through JSON (as the real ones do) clones to an equal value.
func TestCloneOfDecodedStateIsEqual(t *testing.T) {
	t.Parallel()
	raw := `{"messages":[{"id":"m","text":"x","room":"r"}],"rooms":[{"id":"r","name":"R"}],"featureFlags":{"f":true},"incident":true}`
	var s State
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(s, s.Clone()) {
		t.Fatal("decoded state clone differs")
	}
}
