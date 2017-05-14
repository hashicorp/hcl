package zcl

// ElementHeaderSchema represents the shape of an element header, and is
// used for matching elements within bodies.
type ElementHeaderSchema struct {
	Name       string
	LabelNames []string
	Single     bool
}

// BodySchema represents the desired shallow structure of a body.
type BodySchema struct {
	Attributes []string
	Elements   []ElementHeaderSchema
}
