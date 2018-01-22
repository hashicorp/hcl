package hcldec

// Verify that all of our spec types implement the necessary interfaces
var _ Spec = ObjectSpec(nil)
var _ Spec = TupleSpec(nil)
var _ Spec = (*AttrSpec)(nil)
var _ Spec = (*LiteralSpec)(nil)
var _ Spec = (*ExprSpec)(nil)
var _ Spec = (*BlockSpec)(nil)
var _ Spec = (*BlockListSpec)(nil)
var _ Spec = (*BlockSetSpec)(nil)
var _ Spec = (*BlockMapSpec)(nil)
var _ Spec = (*BlockLabelSpec)(nil)
var _ Spec = (*DefaultSpec)(nil)

var _ attrSpec = (*AttrSpec)(nil)

var _ blockSpec = (*BlockSpec)(nil)
var _ blockSpec = (*BlockListSpec)(nil)
var _ blockSpec = (*BlockSetSpec)(nil)
var _ blockSpec = (*BlockMapSpec)(nil)
