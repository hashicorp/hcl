// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

//go:build !go1.27
// +build !go1.27

package unicodeutil

import (
	v15 "github.com/apparentlymart/go-textseg/v15/textseg"
)

func ScanGraphemeClusters(data []byte, atEOF bool) (int, []byte, error) {
	return v15.ScanGraphemeClusters(data, atEOF)
}
