package json

import (
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/hcl"
)

// jsonErrors are the errors built up from parsing. These should not
// be accessed directly.
var jsonErrors []error
var jsonLock sync.Mutex
var jsonResult *hcl.Object

// Parse parses the given string and returns the result.
func Parse(v string) (*hcl.Object, error) {
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

	// Build up the errors
	var err error
	if len(jsonErrors) > 0 {
		err = &multierror.Error{Errors: jsonErrors}
		return nil, err
	}

	// Remove any keys we consider comments
	removeComments(jsonResult)

	return jsonResult, nil
}

func removeComments(o *hcl.Object) {
	if o.Type != hcl.ValueTypeObject {
		return
	}

	members := o.Value.([]*hcl.Object)
	newMembers := make([]*hcl.Object, 0, len(members))

	for _, obj := range members {
		if isComment(obj.Key) {
			continue
		}

		newMembers = append(newMembers, obj)

		if obj.Type == hcl.ValueTypeObject {
			removeComments(obj)
		}
	}

	o.Value = newMembers
}

func isComment(key string) bool {
	return strings.HasPrefix(key, "//")
}
