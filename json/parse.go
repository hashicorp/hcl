package json

import (
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/hcl/ast"
)

// jsonErrors are the errors built up from parsing. These should not
// be accessed directly.
var jsonErrors []error
var jsonLock sync.Mutex
var jsonResult *ast.File

// Parse parses the given string and returns the result.
func Parse(v string) (*ast.File, error) {
	jsonLock.Lock()
	defer jsonLock.Unlock()
	jsonErrors = nil
	jsonResult = nil

	// Parse
	lex := &jsonLex{Input: v}
	jsonParse(lex)

	// If we have an error in the lexer itself, return it
	if lex.err != nil {
		return nil, lex.err
	}

	// If we have a result, flatten it. This is an operation we take on
	// to make our AST look more like traditional HCL. This makes parsing
	// it a lot easier later.
	if jsonResult != nil {
		flattenObjects(jsonResult)
	}

	// Build up the errors
	var err error
	if len(jsonErrors) > 0 {
		err = &multierror.Error{Errors: jsonErrors}
		jsonResult = nil
	}

	return jsonResult, err
}

// flattenObjects takes an AST node, walks it, and flattens
func flattenObjects(node ast.Node) {
	ast.Walk(jsonResult, func(n ast.Node) bool {
		// We only care about lists, because this is what we modify
		list, ok := n.(*ast.ObjectList)
		if !ok {
			return true
		}

		// Rebuild the item list
		items := make([]*ast.ObjectItem, 0, len(list.Items))
		frontier := make([]*ast.ObjectItem, len(list.Items))
		copy(frontier, list.Items)
		for len(frontier) > 0 {
			// Pop the current item
			n := len(frontier)
			item := frontier[n-1]
			frontier = frontier[:n-1]

			// We only care if the value of this item is an object
			ot, ok := item.Val.(*ast.ObjectType)
			if !ok {
				items = append(items, item)
				continue
			}

			// All the elements of this object must also be objects!
			match := true
			for _, item := range ot.List.Items {
				if _, ok := item.Val.(*ast.ObjectType); !ok {
					match = false
					break
				}
			}
			if !match {
				items = append(items, item)
				continue
			}

			// Great! We have a match go through all the items and flatten
			for _, subitem := range ot.List.Items {
				// Copy the new key
				keys := make([]*ast.ObjectKey, len(item.Keys)+len(subitem.Keys))
				copy(keys, item.Keys)
				copy(keys[len(item.Keys):], subitem.Keys)

				// Add it to the frontier so that we can recurse
				frontier = append(frontier, &ast.ObjectItem{
					Keys:        keys,
					Assign:      item.Assign,
					Val:         subitem.Val,
					LeadComment: item.LeadComment,
					LineComment: item.LineComment,
				})
			}
		}

		// Reverse the list since the frontier model runs things backwards
		for i := len(items)/2 - 1; i >= 0; i-- {
			opp := len(items) - 1 - i
			items[i], items[opp] = items[opp], items[i]
		}

		// Done! Set the original items
		list.Items = items
		return true
	})
}
