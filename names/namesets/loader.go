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
	"math"
	"strconv"
	"strings"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xos"
)

// MustLoadFromReader loads a name set as LoadFromReader does, but any error terminates the process, so it is intended
// for known-good data such as an embedded corpus.
func MustLoadFromReader(r io.Reader) map[string]int {
	m, err := LoadFromReader(r)
	xos.ExitIfErr(err)
	return m
}

// LoadFromReader loads a name set from the provided reader. The data should consist of lines of text, each of which
// contains a name optionally followed by a comma and a count. A count is recognized only when the text after the final
// comma parses as an integer; a dangling trailing comma is dropped, and any other comma is part of the name ("Smith,
// Jr."). When no count is given, 1 is assumed. Counts for a name that appears on more than one line accumulate, and a
// name whose total is less than 1 is removed from the returned set, so a count of 0 suppresses a name that appears only
// once and a negative count offsets earlier lines.
//
// Lines are read with the default bufio.Scanner buffer, so a single line longer than roughly 64KB stops the scan: the
// names accumulated so far are returned along with a non-nil error.
func LoadFromReader(r io.Reader) (map[string]int, error) {
	m := make(map[string]int)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		name := line
		count := int64(1)
		if idx := strings.LastIndex(line, ","); idx >= 0 {
			if suffix := strings.TrimSpace(line[idx+1:]); suffix == "" {
				name = line[:idx]
			} else if parsed, err := strconv.ParseInt(suffix, 10, 64); err == nil {
				name = line[:idx]
				count = parsed
			}
		}
		if name = strings.TrimSpace(name); name == "" {
			continue
		}
		m[name] = saturatingAddInt32(m[name], count)
	}
	for name, count := range m {
		if count < 1 {
			delete(m, name)
		}
	}
	return m, errs.Wrap(scanner.Err())
}

// saturatingAddInt32 returns sum + delta with both delta and the total clamped to the int32 range, so huge counts can
// neither wrap a platform int nor exceed the maxWeight ceiling the namers apply, while a total below 1 still suppresses
// the name.
func saturatingAddInt32(sum int, delta int64) int {
	delta = min(max(delta, math.MinInt32), math.MaxInt32)
	total := min(max(int64(sum)+delta, math.MinInt32), math.MaxInt32)
	return int(total)
}
