# HCL Dynamic Blocks Extension

This HCL extension implements a special block type named "dynamic" that can
be used to dynamically generate blocks of other types by iterating over
collection values.

Normally the block structure in an HCL configuration file is rigid, even
though dynamic expressions can be used within attribute values. This is
convenient for most applications since it allows the overall structure of
the document to be decoded easily, but in some applications it is desirable
to allow dynamic block generation within certain portions of the configuration.

Dynamic block generation is performed using the `dynamic` block type:

```hcl
toplevel {
  nested {
    foo = "static block 1"
  }

  dynamic "nested" {
    for_each = ["a", "b", "c"]
    iterator = nested
    content {
      foo = "dynamic block ${nested.value}"
    }
  }

  nested {
    foo = "static block 2"
  }
}
```

The above is interpreted as if it were written as follows:

```hcl
toplevel {
  nested {
    foo = "static block 1"
  }

  nested {
    foo = "dynamic block a"
  }

  nested {
    foo = "dynamic block b"
  }

  nested {
    foo = "dynamic block c"
  }

  nested {
    foo = "static block 2"
  }
}
```

Since HCL block syntax is not normally exposed to the possibility of unknown
values, this extension must make some compromises when asked to iterate over
an unknown collection. If the length of the collection cannot be statically
recognized (because it is an unknown value of list, map, or set type) then
the `dynamic` construct will generate a _single_ dynamic block whose iterator
key and value are both unknown values of the dynamic pseudo-type, thus causing
any attribute values derived from iteration to appear as unknown values. There
is no explicit representation of the fact that the length of the collection may
eventually be different than one.

## Usage

Pass a body to function `Expand` to obtain a new body that will, on access
to its content, evaluate and expand any nested `dynamic` blocks.
Dynamic block processing is also automatically propagated into any nested
blocks that are returned, allowing users to nest dynamic blocks inside
one another and to nest dynamic blocks inside other static blocks.

HCL structural decoding does not normally have access to an `EvalContext`, so
any variables and functions that should be available to the `for_each`
and `labels` expressions must be passed in when calling `Expand`. Expressions
within the `content` block are evaluated separately and so can be passed a
separate `EvalContext` if desired, during normal attribute expression
evaluation.

Some applications dynamically generate an `EvalContext` by analyzing which
variables are referenced by an expression before evaluating it. This can be
achieved for a block that might contain `dynamic` blocks by calling
`ForEachVariables`, which returns the variables required by the `for_each`
and `labels` attributes in all `dynamic` blocks within the given body,
including any nested `dynamic` blocks.

# Performance

This extension is going quite harshly against the grain of the HCL API, and
so it uses lots of wrapping objects and temporary data structures to get its
work done. HCL in general is not suitable for use in high-performance situations
or situations sensitive to memory pressure, but that is _especially_ true for
this extension.
