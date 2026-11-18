package room

import (
	"reflect"
	"sync"
	"time"
)

// Clone returns a deep copy of the state: no slice, map or pointer is shared with the
// original, so the copy can be changed freely. Nil and empty (non-nil) slices and maps keep
// their distinction, because the API renders them differently. Fields that are not exported
// are copied by value; the only such type in State is time.Time, which is immutable.
func (s State) Clone() State {
	return cloneValue(reflect.ValueOf(s)).Interface().(State)
}

func cloneValue(in reflect.Value) reflect.Value {
	if isPlain(in.Type()) {
		return in
	}
	switch in.Kind() {
	case reflect.Slice:
		if in.IsNil() {
			return in
		}
		out := reflect.MakeSlice(in.Type(), in.Len(), in.Len())
		if isPlain(in.Type().Elem()) {
			reflect.Copy(out, in)
			return out
		}
		for i := 0; i < in.Len(); i++ {
			out.Index(i).Set(cloneValue(in.Index(i)))
		}
		return out
	case reflect.Map:
		if in.IsNil() {
			return in
		}
		out := reflect.MakeMapWithSize(in.Type(), in.Len())
		iter := in.MapRange()
		for iter.Next() {
			out.SetMapIndex(cloneValue(iter.Key()), cloneValue(iter.Value()))
		}
		return out
	case reflect.Pointer:
		if in.IsNil() {
			return in
		}
		out := reflect.New(in.Type().Elem())
		out.Elem().Set(cloneValue(in.Elem()))
		return out
	case reflect.Interface:
		if in.IsNil() {
			return in
		}
		out := reflect.New(in.Type()).Elem()
		out.Set(cloneValue(in.Elem()))
		return out
	case reflect.Struct:
		out := reflect.New(in.Type()).Elem()
		out.Set(in) // unexported fields and plain fields by value
		for i := 0; i < in.NumField(); i++ {
			if in.Type().Field(i).IsExported() && !isPlain(in.Type().Field(i).Type) {
				out.Field(i).Set(cloneValue(in.Field(i)))
			}
		}
		return out
	case reflect.Array:
		out := reflect.New(in.Type()).Elem()
		for i := 0; i < in.Len(); i++ {
			out.Index(i).Set(cloneValue(in.Index(i)))
		}
		return out
	}
	return in
}

var (
	plainTypes sync.Map // reflect.Type -> bool
	timeType   = reflect.TypeOf(time.Time{})
)

// isPlain reports types that contain no reference that a caller could change, so a plain copy
// is a deep copy: numbers, booleans, strings (immutable), time.Time (immutable) and arrays and
// structs made only of those. The answer is remembered per type.
func isPlain(t reflect.Type) bool {
	if known, ok := plainTypes.Load(t); ok {
		return known.(bool)
	}
	plain := false
	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String:
		plain = true
	case reflect.Array:
		plain = isPlain(t.Elem())
	case reflect.Struct:
		plain = true
		if t != timeType {
			for i := 0; i < t.NumField(); i++ {
				if !isPlain(t.Field(i).Type) {
					plain = false
					break
				}
			}
		}
	}
	plainTypes.Store(t, plain)
	return plain
}
