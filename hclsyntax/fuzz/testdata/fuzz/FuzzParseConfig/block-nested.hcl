# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

go test fuzz v1
[]byte("block {\n  another_block {\n    foo = bar\n  }\n}\n")