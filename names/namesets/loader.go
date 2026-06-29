// Copyright (c) 2017-2026 by Richard A. Wilkes. All rights reserved.
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
// contains a name and a count, separated by a comma. The trailing comma and count may be omitted, or the count may be
// unparseable, in which case a value of 1 is assumed. An explicit count of less than 1 removes the name from the
// returned set (matching the namer constructors), so a data author can suppress a name by giving it a count of 0.
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
			// Honor an explicit count exactly, including a value less than 1: per the namer constructors such a
			// count removes the name from the set, so a data author can suppress a name with a count of 0. Only a
			// malformed (unparseable) count falls back to the default of 1.
			if parsed, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64); err == nil {
				count = parsed
			}
		}
		m[parts[0]] += int(count)
	}
	// Drop names whose accumulated count is less than 1 so the returned set never contains a name that was suppressed
	// with a count of 0 (or one whose positive and negative counts canceled out).
	for name, count := range m {
		if count < 1 {
			delete(m, name)
		}
	}
	return m, errs.Wrap(scanner.Err())
}
