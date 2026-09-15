# Security Model

HCL is a toolkit for building structured, expressive configuration languages. The
evaluator is powerful by design: it resolves variables, calls functions, and processes
complex expressions on behalf of the calling application. The security properties of an
HCL integration are determined primarily by how the library is used, not by the library
itself.

This document describes the integration security model for Go developers embedding HCL
in their own applications. It is not exhaustive. Integrators are responsible for assessing
their own risk posture and applying controls appropriate to their deployment context.

> **HCL configuration from untrusted sources must not be evaluated unless appropriate
> security controls are in place.**

## What HCL Guarantees

The HCL library itself makes the following guarantees in isolation:

- The expression evaluator does not make network calls, read environment variables, or
  execute system commands.
- HCL reads files from disk during the **parsing** step when loading configuration
  source, but the evaluator itself performs no I/O.
- HCL does not dynamically load code or plugins at runtime.
- The type system is strict. Evaluated values are confined to the types explicitly
  exposed by the integrator; the evaluator cannot reach outside that boundary on its
  own.

These guarantees apply to the library in isolation. They say nothing about what
integrator-supplied functions and variables are capable of doing.

## Integrator Responsibilities

The integrator controls what untrusted input HCL is asked to process: the source text
passed to the parser, the variables and functions made available during evaluation, and
the data flowing through both. HCL imposes no restrictions on any of these once they are
provided. The integrator is therefore the primary security control for any HCL-based
system.

### Principle of least privilege

Expose only the functions and variables needed for the configuration's intended purpose.
Do not expose general-purpose utilities that could be chained into unintended behavior.
The smaller the surface area presented to untrusted input, the smaller the impact of a
malicious or malformed configuration.

### Input validation

Untrusted input passed to HCL, whether as configuration source text or as values
supplied during evaluation, must be validated and sanitized before use. HCL operates on
what it is given without additional checks; ensuring that input is safe is the
integrator's responsibility.

HCL provides no stack overflow protection. Deeply nested expressions or structures in
untrusted configuration source can exhaust the call stack and crash the process.
Integrators that accept configuration from untrusted sources should enforce limits on
input size and structural depth before passing it to HCL.

### Privileged configuration and optional extensions

HCL includes optional extensions that significantly expand what a configuration can
express, including user-defined functions and dynamic block generation. These extensions
should only be enabled for **privileged configuration**: configuration source that is
fully controlled by the integrator or a trusted operator. They should not be enabled for
end-user-supplied input unless strong, independent controls are in place that bound what
a user can do.
