# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

go test fuzz v1
[]byte("a = \"hi $${var.foo}\"\n")