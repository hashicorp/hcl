package zcl

// MergeFiles combines the given files to produce a single body that contains
// configuration from all of the given files.
//
// The ordering of the given files decides the order in which contained
// elements will be returned. If any top-level attributes are defined with
// the same name across multiple files, a diagnostic will be produced from
// the Content and PartialContent methods describing this error in a
// user-friendly way.
func MergeFiles(files []*File) Body {
	var bodies []Body
	for _, file := range files {
		bodies = append(bodies, file.Body)
	}
	return MergeBodies(bodies)
}

// MergeBodies is like MergeFiles except it deals directly with bodies, rather
// than with entire files.
func MergeBodies(bodies []Body) Body {
	return mergedBodies(bodies)
}

type mergedBodies []Body

func (mb mergedBodies) Content(schema *BodySchema) (*BodyContent, Diagnostics) {
	// TODO: Implement
	return nil, nil
}

func (mb mergedBodies) PartialContent(schema *BodySchema) (*BodyContent, Body, Diagnostics) {
	// TODO: Implement
	return nil, nil, nil
}
