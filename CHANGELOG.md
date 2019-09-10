# HCL Changelog

## v2.0.0 (Unreleased)

Initial release of HCL 2, which is a new implementating combining the HCL 1
language with the HIL expression language to produce a single language
supporting both nested configuration structures and arbitrary expressions.

HCL 2 has an entirely new Go library API and so is not a drop-in upgrade
relative to HCL 1. It's possible to import both versions of HCL into a single
program using Go's _semantic import versioning_ mechanism:

```
import (
    hcl1 "github.com/hashicorp/hcl"
    hcl2 "github.com/hashicorp/hcl/v2"
)
```

---

Prior to v2.0.0 there was not a curated changelog. Consult the git history for
information on the changes to HCL 1.
