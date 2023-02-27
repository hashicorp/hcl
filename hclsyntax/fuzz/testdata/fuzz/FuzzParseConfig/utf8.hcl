# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

go test fuzz v1
[]byte("foo = \"föo ${föo(\"föo\")}\"\n")