package hcl

import (
	"sync"
)

// exprErrors are the errors built up from parsing. These should not
// be accessed directly.
var exprErrors []error
var exprLock sync.Mutex
var exprResult []map[string]interface{}

/*
// ExprParse parses the given expression and returns an executable
// Interpolation.
func ExprParse(v string) (Interpolation, error) {
	exprLock.Lock()
	defer exprLock.Unlock()
	exprErrors = nil
	exprResult = nil

	// Parse
	exprParse(&exprLex{input: v})

	// Build up the errors
	var err error
	if len(exprErrors) > 0 {
		err = &multierror.Error{Errors: exprErrors}
		exprResult = nil
	}

	return exprResult, err
}
*/
