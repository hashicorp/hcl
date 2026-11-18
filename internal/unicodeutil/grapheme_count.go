// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package unicodeutil

import (
	"bufio"
	"bytes"
)

func GraphemeCount(buf []byte) (int, error) {
	scanner := bufio.NewScanner(bytes.NewReader(buf))
	scanner.Buffer(make([]byte, 0, 2), bufio.MaxScanTokenSize)
	scanner.Split(ScanGraphemeClusters)
	var ret int
	for scanner.Scan() {
		ret++
	}
	return ret, scanner.Err()
}
