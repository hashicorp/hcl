// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

//go:build go1.27

package unicodeutil

import (
	v17 "github.com/apparentlymart/go-textseg/v17/textseg"
)

func ScanGraphemeClusters(data []byte, atEOF bool) (int, []byte, error) {
	return v17.ScanGraphemeClusters(data, atEOF)
}
