package main

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2"
)

func findTraversalSpec(got hcl.Traversal, candidates []*TestFileExpectTraversal) *TestFileExpectTraversal {
	for _, candidate := range candidates {
		if traversalsAreEquivalent(candidate.Traversal, got) {
			return candidate
		}
	}
	return nil
}

func findTraversalForSpec(want *TestFileExpectTraversal, have []hcl.Traversal) hcl.Traversal {
	for _, candidate := range have {
		if traversalsAreEquivalent(candidate, want.Traversal) {
			return candidate
		}
	}
	return nil
}

func traversalsAreEquivalent(a, b hcl.Traversal) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		aStep := a[i]
		bStep := b[i]

		if reflect.TypeOf(aStep) != reflect.TypeOf(bStep) {
			return false
		}

		// We can now assume that both are of the same type.
		switch ts := aStep.(type) {

		case hcl.TraverseRoot:
			if bStep.(hcl.TraverseRoot).Name != ts.Name {
				return false
			}

		case hcl.TraverseAttr:
			if bStep.(hcl.TraverseAttr).Name != ts.Name {
				return false
			}

		case hcl.TraverseIndex:
			if !bStep.(hcl.TraverseIndex).Key.RawEquals(ts.Key) {
				return false
			}

		default:
			return false
		}
	}
	return true
}

// checkTraversalsMatch determines if a given traversal matches the given
// expectation, which must've been produced by an earlier call to
// findTraversalSpec for the same traversal.
func checkTraversalsMatch(got hcl.Traversal, filename string, match *TestFileExpectTraversal) hcl.Diagnostics {
	var diags hcl.Diagnostics

	gotRng := got.SourceRange()
	wantRng := match.Range

	if got, want := gotRng.Filename, filename; got != want {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Incorrect filename in detected traversal",
			Detail: fmt.Sprintf(
				"Filename was reported as %q, but was expecting %q.",
				got, want,
			),
			Subject: match.Traversal.SourceRange().Ptr(),
		})
		return diags
	}

	// If we have the expected filename then we'll use that to construct the
	// full "want range" here so that we can use it to point to the appropriate
	// location in the remaining diagnostics.
	wantRng.Filename = filename

	if got, want := gotRng.Start, wantRng.Start; got != want {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Incorrect start position in detected traversal",
			Detail: fmt.Sprintf(
				"Start position was reported as line %d column %d byte %d, but was expecting line %d column %d byte %d.",
				got.Line, got.Column, got.Byte,
				want.Line, want.Column, want.Byte,
			),
			Subject: &wantRng,
		})
	}
	if got, want := gotRng.End, wantRng.End; got != want {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Incorrect end position in detected traversal",
			Detail: fmt.Sprintf(
				"End position was reported as line %d column %d byte %d, but was expecting line %d column %d byte %d.",
				got.Line, got.Column, got.Byte,
				want.Line, want.Column, want.Byte,
			),
			Subject: &wantRng,
		})
	}
	return diags
}
