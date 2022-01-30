package gohcl

import (
	"reflect"
)

// Our WithRange[T] support is currently conditional on whether we're running
// in a Go 1.18 or later toolchain, and thus we can use generics.
//
// This file contains some items we need regardless of whether we have that
// turned on, just so that our version-agnostic callers can still work even
// when it's disabled.
//
// See with_range_118.go for the parts that are active only in Go 1.18 or later,
// and with_range_compat.go for a stub we'll use in earlier Go versions.

// withRangeReflect is a reflection-oriented description of a value of any
// specific WithRange[T] type, which a caller can therefore interpret without
// using any Go 1.18-only language features.
type withRangeReflect struct {
	containerPtr reflect.Value
	container    reflect.Value

	value reflect.Value
	rng   reflect.Value
}
