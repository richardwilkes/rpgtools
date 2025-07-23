// Copyright (c) 2017-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package namesets

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xos"
)

// MustLoadFromReader loads a name set from the provided reader. The data should consist of lines of text, each of which
// contains a name and a count, separated by a comma.
func MustLoadFromReader(r io.Reader) map[string]int {
	m, err := LoadFromReader(r)
	xos.ExitIfErr(err)
	return m
}

// LoadFromReader loads a name set from the provided reader. The data should consist of lines of text, each of which
// contains a name and a count, separated by a comma. The trailing comma and count may be omitted, in which case a value
// of 1 is assumed.
func LoadFromReader(r io.Reader) (map[string]int, error) {
	m := make(map[string]int)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if parts[0] = strings.TrimSpace(parts[0]); parts[0] == "" {
			continue
		}
		count := int64(1)
		if len(parts) > 1 {
			var err error
			if count, err = strconv.ParseInt(parts[1], 10, 64); err != nil || count < 1 {
				count = 1
			}
		}
		m[parts[0]] += int(count)
	}
	return m, errs.Wrap(scanner.Err())
}
