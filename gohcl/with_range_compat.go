//go:build !go1.18
// +build !go1.18

package gohcl

// analyzeWithRange is an internal adapter to allow Go-version-agnostic callers
// to compile regardless of whether we are using Go 1.18 features or not.
//
// On versions of Go prior to 1.18, this just immediately returns nil to
// indicate that no value can possibly have a WithRange type on prior versions;
// that type isn't declared at all, then.
func analyzeWithRange(v interface{}) *withRangeReflect {
	return nil
}
