package zcl

import (
	"fmt"

	"github.com/apparentlymart/go-cty/cty"
)

// A Traversal is a description of traversing through a value through a
// series of operations such as attribute lookup, index lookup, etc.
//
// It is used to look up values in scopes, for example.
//
// The traversal operations are implementations of interface Traverser.
// This is a closed set of implementations, so the interface cannot be
// implemented from outside this package.
//
// A traversal can be absolute (its first value is a symbol name) or relative
// (starts from an existing value).
type Traversal []Traverser

// TraversalJoin appends a relative traversal to an absolute traversal to
// produce a new absolute traversal.
func TraversalJoin(abs Traversal, rel Traversal) Traversal {
	if abs.IsRelative() {
		panic("first argument to TraversalJoin must be absolute")
	}
	if !rel.IsRelative() {
		panic("second argument to TraversalJoin must be relative")
	}

	ret := make(Traversal, len(abs)+len(rel))
	copy(ret, abs)
	copy(ret[len(abs):], rel)
	return ret
}

// TraverseRel applies the receiving traversal to the given value, returning
// the resulting value. This is supported only for relative traversals,
// and will panic if applied to an absolute traversal.
func (t Traversal) TraverseRel(val cty.Value) (cty.Value, Diagnostics) {
	current := val
	var diags Diagnostics
	for _, tr := range t {
		var newDiags Diagnostics
		current, newDiags = tr.TraversalStep(current)
		diags = append(diags, newDiags...)
		if newDiags.HasErrors() {
			return cty.DynamicVal, diags
		}
	}
	return current, diags
}

// TraverseAbs applies the receiving traversal to the given eval context,
// returning the resulting value. This is supported only for absolute
// traversals, and will panic if applied to a relative traversal.
func (t Traversal) TraverseAbs(ctx *EvalContext) (cty.Value, Diagnostics) {
	// TODO: implement
	return cty.DynamicVal, nil
}

// IsRelative returns true if the receiver is a relative traversal, or false
// otherwise.
func (t Traversal) IsRelative() bool {
	if len(t) == 0 {
		return true
	}
	if _, firstIsRoot := t[0].(TraverseRoot); firstIsRoot {
		return true
	}
	return false
}

// SimpleSplit returns a TraversalSplit where the name lookup is the absolute
// part and the remainder is the relative part. Supported only for
// absolute traversals, and will panic if applied to a relative traversal.
//
// This can be used by applications that have a relatively-simple variable
// namespace where only the top-level is directly populated in the scope, with
// everything else handled by relative lookups from those initial values.
func (t Traversal) SimpleSplit() TraversalSplit {
	if t.IsRelative() {
		panic("can't use SimpleSplit on a relative traversal")
	}
	return TraversalSplit{
		Abs: t[0:1],
		Rel: t[1:],
	}
}

// RootName returns the root name for a absolute traversal. Will panic if
// called on a relative traversal.
func (t Traversal) RootName() string {
	if t.IsRelative() {
		panic("can't use RootName on a relative traversal")

	}
	return t[0].(TraverseRoot).Name
}

// TraversalSplit represents a pair of traversals, the first of which is
// an absolute traversal and the second of which is relative to the first.
//
// This is used by calling applications that only populate prefixes of the
// traversals in the scope, with Abs representing the part coming from the
// scope and Rel representing the remaining steps once that part is
// retrieved.
type TraversalSplit struct {
	Abs Traversal
	Rel Traversal
}

// TraverseAbs traverses from a scope to the value resulting from the
// absolute traversal.
func (t TraversalSplit) TraverseAbs(ctx *EvalContext) (cty.Value, Diagnostics) {
	return t.Abs.TraverseAbs(ctx)
}

// TraverseRel traverses from a given value, assumed to be the result of
// TraverseAbs on some scope, to a final result for the entire split traversal.
func (t TraversalSplit) TraverseRel(val cty.Value) (cty.Value, Diagnostics) {
	return t.Rel.TraverseRel(val)
}

// Traverse is a convenience function to apply TraverseAbs followed by
// TraverseRel.
func (t TraversalSplit) Traverse(ctx *EvalContext) (cty.Value, Diagnostics) {
	v1, diags := t.TraverseAbs(ctx)
	if diags.HasErrors() {
		return cty.DynamicVal, diags
	}
	v2, newDiags := t.TraverseRel(v1)
	diags = append(diags, newDiags...)
	return v2, diags
}

// Join concatenates together the Abs and Rel parts to produce a single
// absolute traversal.
func (t TraversalSplit) Join() Traversal {
	return TraversalJoin(t.Abs, t.Rel)
}

// RootName returns the root name for the absolute part of the split.
func (t TraversalSplit) RootName() string {
	return t.Abs.RootName()
}

// A Traverser is a step within a Traversal.
type Traverser interface {
	TraversalStep(cty.Value) (cty.Value, Diagnostics)
	isTraverserSigil() isTraverser
}

// Embed this in a struct to declare it as a Traverser
type isTraverser struct {
}

func (tr isTraverser) isTraverserSigil() isTraverser {
	return isTraverser{}
}

// TraverseRoot looks up a root name in a scope. It is used as the first step
// of an absolute Traversal, and cannot itself be traversed directly.
type TraverseRoot struct {
	isTraverser
	Name     string
	SrcRange Range
}

// TraversalStep on a TraverseName immediately panics, because absolute
// traversals cannot be directly traversed.
func (tn TraverseRoot) TraversalStep(cty.Value) (cty.Value, Diagnostics) {
	panic("Cannot traverse an absolute traversal")
}

// TraverseAttr looks up an attribute in its initial value.
type TraverseAttr struct {
	isTraverser
	Name     string
	SrcRange Range
}

func (tn TraverseAttr) TraversalStep(val cty.Value) (cty.Value, Diagnostics) {
	if val.IsNull() {
		return cty.DynamicVal, Diagnostics{
			{
				Severity: DiagError,
				Summary:  "Attempt to get attribute from null value",
				Detail:   "This value is null, so it does not have any attributes.",
				Subject:  &tn.SrcRange,
			},
		}
	}

	ty := val.Type()
	switch {
	case ty.IsObjectType():
		if !ty.HasAttribute(tn.Name) {
			return cty.DynamicVal, Diagnostics{
				{
					Severity: DiagError,
					Summary:  "Unsupported attribute",
					Detail:   fmt.Sprintf("This object does not have an attribute named %q.", tn.Name),
					Subject:  &tn.SrcRange,
				},
			}
		}

		if !val.IsKnown() {
			return cty.UnknownVal(ty.AttributeType(tn.Name)), nil
		}

		return val.GetAttr(tn.Name), nil
	case ty.IsMapType():
		if !val.IsKnown() {
			return cty.UnknownVal(ty.ElementType()), nil
		}

		idx := cty.StringVal(tn.Name)
		if val.HasIndex(idx).False() {
			return cty.DynamicVal, Diagnostics{
				{
					Severity: DiagError,
					Summary:  "Missing map element",
					Detail:   fmt.Sprintf("This map does not have an element with the key %q.", tn.Name),
					Subject:  &tn.SrcRange,
				},
			}
		}

		return val.Index(idx), nil
	case ty == cty.DynamicPseudoType:
		return cty.DynamicVal, nil
	default:
		return cty.DynamicVal, Diagnostics{
			{
				Severity: DiagError,
				Summary:  "Unsupported attribute",
				Detail:   "This value does not have any attributes.",
				Subject:  &tn.SrcRange,
			},
		}
	}
}

// TraverseIndex applies the index operation to its initial value.
type TraverseIndex struct {
	isTraverser
	Key      cty.Value
	SrcRange Range
}

func (tn TraverseIndex) TraversalStep(val cty.Value) (cty.Value, Diagnostics) {
	if val.IsNull() {
		return cty.DynamicVal, Diagnostics{
			{
				Severity: DiagError,
				Summary:  "Attempt to index null value",
				Detail:   "This value is null, so it does not have any indices.",
				Subject:  &tn.SrcRange,
			},
		}
	}
	if tn.Key.IsNull() {
		return cty.DynamicVal, Diagnostics{
			{
				Severity: DiagError,
				Summary:  "Invalid index",
				Detail:   "Can't use a null value as an index.",
				Subject:  &tn.SrcRange,
			},
		}
	}
	ty := val.Type()
	kty := tn.Key.Type()
	if kty == cty.DynamicPseudoType {
		return cty.DynamicVal, nil
	}

	switch {
	case ty.IsListType() || ty.IsTupleType() || ty.IsMapType():
		has := val.HasIndex(tn.Key)
		if !has.IsKnown() {
			if ty.IsTupleType() {
				return cty.DynamicVal, nil
			} else {
				return cty.UnknownVal(ty.ElementType()), nil
			}
		}
		if has.False() {
			return cty.DynamicVal, Diagnostics{
				{
					Severity: DiagError,
					Summary:  "Invalid index",
					Detail:   "The given index value does not identify an element in this collection value.",
					Subject:  &tn.SrcRange,
				},
			}
		}

		return val.Index(tn.Key), nil
	case ty.IsObjectType():
		if kty != cty.String {
			return cty.DynamicVal, Diagnostics{
				{
					Severity: DiagError,
					Summary:  "Invalid index",
					Detail:   "The given index value does not identify an element in this collection value.",
					Subject:  &tn.SrcRange,
				},
			}
		}
		if !val.IsKnown() {
			return cty.DynamicVal, nil
		}
		if !tn.Key.IsKnown() {
			return cty.DynamicVal, nil
		}

		attrName := tn.Key.AsString()

		if !ty.HasAttribute(attrName) {
			return cty.DynamicVal, Diagnostics{
				{
					Severity: DiagError,
					Summary:  "Invalid index",
					Detail:   "The given index value does not identify an element in this collection value.",
					Subject:  &tn.SrcRange,
				},
			}
		}

		return val.GetAttr(attrName), nil
	case ty == cty.DynamicPseudoType:
		return cty.DynamicVal, nil
	default:
		return cty.DynamicVal, Diagnostics{
			{
				Severity: DiagError,
				Summary:  "Invalid index",
				Detail:   "This value does not have any indices.",
				Subject:  &tn.SrcRange,
			},
		}
	}
}

// TraverseSplat applies the splat operation to its initial value.
type TraverseSplat struct {
	isTraverser
	Each     Traversal
	SrcRange Range
}

func (tn TraverseSplat) TraversalStep(val cty.Value) (cty.Value, Diagnostics) {
	panic("TraverseSplat not yet implemented")
}
