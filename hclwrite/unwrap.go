// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package hclwrite

import "reflect"

type unwrapNode interface {
	unwrap() *node
}

// cross-reference: hcl.UnwrapExpressionUntil
func unwrapUntilType(n *node, typ reflect.Type) *node {
TEST:
	if n == nil || reflect.TypeOf(n.content) == typ {
		return n
	}

	if un, ok := n.content.(unwrapNode); ok {
		n = un.unwrap()
		goto TEST
	}

	return nil
}
