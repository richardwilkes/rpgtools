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

// MustLoadFromReader loads a name set from the provided reader. The data should consist of lines of text, each of which
// contains a name and a count, separated by a comma. As with LoadFromReader, a single line longer than roughly 64KB is
// rejected; here that error terminates the process, so this is intended for known-good data such as an embedded corpus.
func MustLoadFromReader(r io.Reader) map[string]int {
	m, err := LoadFromReader(r)
	xos.ExitIfErr(err)
	return m
}

// LoadFromReader loads a name set from the provided reader. The data should consist of lines of text, each of which
// contains a name optionally followed by a comma and a count. A count is recognized only when the text after the final
// comma parses as an integer; a dangling trailing comma (with nothing after it) is dropped, and any other comma is part
// of the name, so a name that itself contains a comma, such as "Smith, Jr.", is kept intact rather than truncated. When
// no count is given a value of 1 is assumed. An explicit count of less than 1 removes the name from the returned set
// (matching the namer constructors), so a data author can suppress a name by giving it a count of 0.
//
// Lines are read with the default bufio.Scanner buffer, so a single line longer than roughly 64KB is not split or
// truncated: the scan stops at that line and a non-nil error is returned along with the names accumulated so far. A
// well-formed name set, with one short name per line, never approaches this limit.
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
		// Accumulate in int64 and saturate the per-name total at the int32 range. Without this, two very large counts
		// for the same name could wrap a platform int negative, and the "< 1" filter below would then delete the name
		// entirely. The math.MaxInt32 ceiling matches the weight cap the namer constructors apply, while a net total
		// below 1 is left as-is so it still suppresses the name.
		m[name] = saturatingAddInt32(m[name], count)
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

// saturatingAddInt32 returns sum + delta with both the addend and the running total clamped to the int32 range, so a
// pathologically large or repeated count can neither wrap a platform int nor exceed the math.MaxInt32 weight ceiling
// the namers use. A net total below 1 is left negative or zero so the caller's suppression filter still removes the
// name.
func saturatingAddInt32(sum int, delta int64) int {
	delta = min(max(delta, math.MinInt32), math.MaxInt32)
	total := min(max(int64(sum)+delta, math.MinInt32), math.MaxInt32)
	return int(total)
}
