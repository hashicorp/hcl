# hcldec

`hcldec` is a command line tool that transforms HCL input into JSON output
using a decoding specification given by the user.

This tool is intended as a "glue" tool, with use-cases like the following:

* Define a HCL-based configuration format for a third-party tool that takes
  JSON as input, and then translate the HCL configuration into JSON before
  running the tool. (See [the `npm-package` example](examples/npm-package).)

* Use HCL from languages where a HCL parser/decoder is not yet available.
  At the time of writing, that's any language other than Go.

* In particular, define a HCL-based configuration format for a shell script
  and then use `jq` to load the result into environment variables for
  further processing. (See [the `sh-config-file` example](examples/sh-config-file).)

## Installation

If you have a working Go development environment, you can install this tool
with `go get` in the usual way:

```
$ go get -u github.com/hashicorp/hcl/v2/cmd/hcldec
```

This will install `hcldec` in `$GOPATH/bin`, which usually places it into
your shell `PATH` so you can then run it as `hcldec`.

## Usage

```
usage: hcldec --spec=<spec-file> [options] [hcl-file ...]
  -o, --out string          write to the given file, instead of stdout
  -s, --spec string         path to spec file (required)
  -V, --vars json-or-file   provide variables to the given configuration file(s)
  -v, --version             show the version number and immediately exit
```

The most important step in using `hcldec` is to write the specification that
defines how to interpret the given configuration files and translate them
into JSON. The following is a simple specification that creates a JSON
object from two top-level attributes in the input configuration:

```hcl
object {
  attr "name" {
    type     = string
    required = true
  }
  attr "is_member" {
    type = bool
  }
}
```

Specification files are conventionally kept in files with a `.hcldec`
extension. We'll call this one `example.hcldec`.

With the above specification, the following input file `example.conf` is
valid:

```hcl
name = "Raul"
```

The spec and the input file can then be provided to `hcldec` to extract a
JSON representation:

```
$ hcldec --spec=example.hcldec example.conf
{"name": "Raul"}
```

The specification defines both how to map the input into a JSON data structure
and what input is valid. The `required = true` specified for the `name`
allows `hcldec` to detect and raise an error when an attribute of that name
is not provided:

```
$ hcldec --spec=example.hcldec typo.conf
Error: Unsupported attribute

  on example.conf line 1:
   1: namme = "Juan"

An attribute named "namme" is not expected here. Did you mean "name"?

Error: Missing required attribute

  on example.conf line 2:

The attribute "name" is required, but no definition was found.
```

## Further Reading

For more details on the `.hcldec` specification file format, see
[the spec file documentation](spec-format.md).
