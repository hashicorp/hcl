// Package include implements a zcl extension that allows inclusion of
// one zcl body into another using blocks of type "include", with the following
// structure:
//
//     include {
//       path = "./foo.zcl"
//     }
//
// The processing of the given path is delegated to the calling application,
// allowing it to decide how to interpret the path and which syntaxes to
// support for referenced files.
package include
