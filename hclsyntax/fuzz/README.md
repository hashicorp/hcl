# hclsyntax fuzzing utilities

This directory contains helper functions and corpora that can be used to
fuzz-test the `hclsyntax` parsers using Go's native fuzz testing capabilities.

Please see https://go.dev/doc/fuzz/ for more information on fuzzing.

## Prerequisites
* Go 1.18

## Running the fuzzer

Each exported function in the `hclsyntax` package has a corresponding fuzz test.
These can be run one at a time via `go test`:

```
$ cd fuzz
$ go test -fuzz FuzzParseTemplate
$ go test -fuzz FuzzParseTraversalAbs
$ go test -fuzz FuzzParseExpression
$ go test -fuzz FuzzParseConfig
```

This command will exit only when a crasher is found (see "Understanding the 
result" below.)

## Seed corpus

The seed corpus for each fuzz test function is stored in the corresponding
directory under `hclsyntax/fuzz/testdata/fuzz/`. For example:

```
$ ls hclsyntax/fuzz/testdata/fuzz/FuzzParseTemplate
empty.tmpl
escape-dollar.tmpl
escape-newline-tmpl
...
```

Additional seed inputs can be added to this corpus. Each file must be in the Go 1.18 corpus file format. Files can be converted to this format using the `file2fuzz` tool. To install it:

```
$ go install golang.org/x/tools/cmd/file2fuzz@latest
$ file2fuzz -help
```

## Understanding the result

A small number of subdirectories will be created in the work directory.

If you let `go-fuzz` run for a few minutes (the more minutes the better) it
may detect "crashers", which are inputs that caused the parser to panic.
These are written to `hclsyntax/fuzz/testdata/fuzz/<fuzz test name>/`:

```
$ ls hclsyntax/fuzz/testdata/fuzz/FuzzParseTemplate
582528ddfad69eb57775199a43e0f9fd5c94bba343ce7bb6724d4ebafe311ed4
```

A good first step to fixing a detected crasher is to copy the failing input
into one of the unit tests in the `hclsyntax` package and see it crash there
too. After that, it's easy to re-run the test as you try to fix it. 
