// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package hclwrite

type unwrapNode interface {
	unwrap() *node
}

// cross-reference: hcl.UnwrapExpressionUntil
func unwrapUntilType[T nodeContent](n *node) *node {
TEST:
	if n == nil {
		return n
	}

	if _, ok := n.content.(T); ok {
		return n
	}

	if un, ok := n.content.(unwrapNode); ok {
		n = un.unwrap()
		goto TEST
	}

	return nil
}
