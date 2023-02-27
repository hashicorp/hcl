# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

a { b = "foo", c = "bar" }
a { b = "foo"
}
a { b = "foo"
  c = "bar" }
a { b = "foo"
  c = "bar"
}
a { d {} }
