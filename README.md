# hcl [![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/fatih/hcl) [![Build Status](http://img.shields.io/travis/fatih/hcl.svg?style=flat-square)](https://travis-ci.org/fatih/hcl)

HCL is a lexer and parser family written in Go for
[HCL](https://github.com/hashicorp/hcl) (Hashicorp Configuration Language). It
has several components, similar to Go's own parser family. It provides a set of
packages to write tools and customize files written in HCL. For example both
[`hclfmt`](https://github.com/fatih/hclfmt) and `hcl2json` (coming soon) is
written based on these tools. 

## API

If you are already familiar with Go's own parser family it's really easy to
dive. It basically resembles the same logic. However there are several differences
and the implementation is completely different. Right now it contains the
following packages:

* `token`: defines constants representing the lexical tokens for a scanned HCL file.
* `scanner`: scanner is a lexical scanner. It scans a given HCL file and
  returns a stream of tokens.
* `ast`: declares the types used to represent the syntax tree for parsed HCL files.
* `parser`:  parses a given HCL file and creates a AST representation
* `printer`: prints any given AST node and formats

## Why 

The whole parser family was created because I wanted a `hclfmt` command, which
like `gofmt` would format a HCL file. I didn't want to use the package
[github/hashicorp/hcl](https://github.com/hashicorp/hcl) in the first place,
because the lexer and parser is generated and it doesn't expose any kind of
flexibility.

Another reason was that I wanted to learn and experience how to implement a
proper lexer and parser in Go. It was really fun and I think it was worth it.

## License

The BSD 3-Clause License - see
[`LICENSE`](https://github.com/fatih/hcl/blob/master/LICENSE) for more
details

