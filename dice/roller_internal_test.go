// Copyright (c) 2017-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package dice

import (
	"math"
	"strconv"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

func TestExtractValueCapsWithoutOverflow(t *testing.T) {
	c := check.New(t)
	const twentyNines = "99999999999999999999" // larger than math.MaxInt
	maxIntDigits := strconv.Itoa(math.MaxInt)
	for i, one := range []struct {
		in       string
		maxValue int
		want     int
		wantPos  int
	}{
		{"123", math.MaxInt, 123, 3},                              // 0 - ordinary value, no capping
		{twentyNines, math.MaxInt, math.MaxInt, len(twentyNines)}, // 1 - regression: cap, never wrap to garbage
		{twentyNines, 999_999, 999_999, len(twentyNines)},         // 2 - a small cap still applies
		{"9", 1, 1, 1},           // 3 - a single digit past a tiny cap
		{"0", math.MaxInt, 0, 1}, // 4 - zero
		{maxIntDigits, math.MaxInt, math.MaxInt, len(maxIntDigits)}, // 5 - exactly math.MaxInt, no spurious cap
		{"18446744073709551616", math.MaxInt, math.MaxInt, 20},      // 6 - 2^64, exercises the overflow branch
		{"7d6", math.MaxInt, 7, 1},                                  // 7 - stops at the first non-digit
	} {
		got, pos := extractValue(one.in, 0, one.maxValue)
		c.Equal(one.want, got, "case %d value", i)
		c.Equal(one.wantPos, pos, "case %d position", i)
		c.True(got >= 0, "case %d produced a negative value %d", i, got)
	}
}
