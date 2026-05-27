package assert

import (
	"reflect"
	"strings"
)

type TestIF interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fail()
	FailNow()
	Failed() bool
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Helper()
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	Name() string
	Skip(args ...interface{})
	SkipNow()
	Skipf(format string, args ...interface{})
	Skipped() bool
}

type Assert struct {
	t TestIF
}

func NewAssert(t TestIF) *Assert {
	return &Assert{
		t: t,
	}
}

func (ast *Assert) Contains(src, want string) {
	ast.t.Helper()
	if !strings.Contains(src, want) {
		ast.t.Error("Not Contains:", "src:", src, " want:", want)
	}
}

func (ast *Assert) Equal(a, b interface{}) {
	ast.t.Helper()
	if a != b {
		ast.t.Error("Not Equal:", a, " : ", b)
	}
}

func (ast *Assert) True(a bool) {
	ast.t.Helper()
	if !a {
		ast.t.Errorf("Not True %t", a)
	}
}

func (ast *Assert) False(a bool) {
	ast.t.Helper()
	if a {
		ast.t.Errorf("Not True %t", a)
	}
}

func (ast *Assert) Nil(obj interface{}) {
	ast.t.Helper()
	if !isNil(obj) {
		ast.t.Errorf("Expected nil, but got: %#v", obj)
	}
}

func (ast *Assert) NotNil(obj interface{}) {
	ast.t.Helper()
	if isNil(obj) {
		ast.t.Error("Expected value not to be nil")
	}
}

func (ast *Assert) DeepEqual(a interface{}, b interface{}) {
	ast.t.Helper()
	if !reflect.DeepEqual(a, b) {
		ast.t.Error("Not Equal:", a, " : ", b)
	}
}

func (ast *Assert) In(items interface{}, item interface{}) {
	ast.t.Helper()
	value := reflect.ValueOf(items)
	if value.Kind() != reflect.Slice && value.Kind() != reflect.Array {
		ast.t.Error("parse error, items must be slice or arrary")
	}
	for i := 0; i < value.Len(); i++ {
		if item == value.Index(i).Interface() {
			return
		}
	}
	ast.t.Error(item, " no in ", items)
}

// containsKind checks if a specified kind in the slice of kinds.
func containsKind(kinds []reflect.Kind, kind reflect.Kind) bool {
	for i := 0; i < len(kinds); i++ {
		if kind == kinds[i] {
			return true
		}
	}
	return false
}

// isNil checks if a specified object is nil or not, without Failing.
func isNil(obj interface{}) bool {
	if obj == nil {
		return true
	}

	value := reflect.ValueOf(obj)
	kind := value.Kind()
	isNilableKind := containsKind(
		[]reflect.Kind{
			reflect.Chan, reflect.Func,
			reflect.Interface, reflect.Map,
			reflect.Pointer, reflect.Slice, reflect.UnsafePointer,
		}, kind)

	return isNilableKind && value.IsNil()
}
