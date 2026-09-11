package exception

import "reflect"

// isNil reports whether obj is a Java-style null: an untyped nil, or a typed
// nil pointer/interface. Assert.notNull only sees the reference, so a typed nil
// pointer must count as null here too.
func isNil(obj any) bool {
	if obj == nil {
		return true
	}
	v := reflect.ValueOf(obj)
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
		return v.IsNil()
	default:
		return false
	}
}
