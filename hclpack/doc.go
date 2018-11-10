// Package hclpack provides a straightforward representation of HCL block/body
// structure that can be easily serialized and deserialized for compact
// transmission (e.g. over a network) without transmitting the full source code.
//
// Expressions are retained in native syntax source form so that their
// evaluation can be delayed until a package structure is decoded by some
// other system that has enough information to populate the evaluation context.
//
// Packed structures retain source location information but do not retain
// actual source code. To make sense of source locations returned in diagnostics
// and via other APIs the caller must somehow gain access to the original source
// code that the packed representation was built from, which is a problem that
// must be solved somehow by the calling application.
package hclpack
