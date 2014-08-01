package hcl

import (
	"sync"

	"github.com/hashicorp/terraform/helper/multierror"
)

// hclErrors are the errors built up from parsing. These should not
// be accessed directly.
var hclErrors []error
var hclLock sync.Mutex
var hclResult *ObjectNode

// Parse parses the given string and returns the result.
func Parse(v string) (*ObjectNode, error) {
	hclLock.Lock()
	defer hclLock.Unlock()
	hclErrors = nil
	hclResult = nil

	// Parse
	hclParse(&hclLex{Input: v})

	// Build up the errors
	var err error
	if len(hclErrors) > 0 {
		err = &multierror.Error{Errors: hclErrors}
		hclResult = nil
	}

	return hclResult, err
}
