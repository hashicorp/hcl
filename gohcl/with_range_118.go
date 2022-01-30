//go:build go1.18
// +build go1.18

package gohcl

import (
	"reflect"

	hcl "github.com/hashicorp/hcl/v2"
)

type WithRange[T any] struct {
	Value T
	Range hcl.Range
}

// withRange is an internal interface implemented only by WithRange[T]
// types, which we'll use to recognize uses of it even though we can't
// predict ahead of time all of the possible type arguments.
type withRange interface {
	// withRangeReflect returns a reflect-package-oriented interpretation of
	// the reciever and its fields.
	withRangeReflect() withRangeReflect
}

func (wr *WithRange[T]) withRangeReflect() withRangeReflect {
	containerPtrV := reflect.ValueOf(wr)
	var containerV reflect.Value

	// If we don't yet have a container (wr is nil) then we'll instantiate
	// one during our work here and describe _that_ to the caller so that
	// they can write it into the appropriate location in the surrounding
	// type once it's all populated.
	if containerPtrV.IsNil() {
		newContainerPtrV := reflect.New(containerPtrV.Elem().Type())
		containerV = newContainerPtrV.Elem()
	} else {
		containerV = containerPtrV.Elem()
	}

	return withRangeReflect{
		containerPtr: containerPtrV,
		container:    containerV,

		value: containerV.FieldByName("Value"),
		rng:   containerV.FieldByName("Range"),
	}
}

// analyzeWithRange is an internal adapter to allow Go-version-agnostic callers
// to compile regardless of whether we are using Go 1.18 features or not.
//
// On Go 1.18 or later, will return a non-nil withRangeReflect pointer if the
// given value has a WithRange type, or nil if it doesn't.
func analyzeWithRange(v interface{}) *withRangeReflect {
	wrI, ok := v.(withRange)
	if !ok {
		return nil
	}

	ret := wrI.withRangeReflect()
	return &ret
}
