# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

go test fuzz v1
[]byte("a = foo.bar[1].baz[\"foo\"].pizza\n")