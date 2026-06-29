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
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

func TestLoadFromReader(t *testing.T) {
	c := check.New(t)
	m, err := LoadFromReader(strings.NewReader(strings.Join([]string{
		"Alice, 100",  // space after the comma
		"Bob,50",      // no space
		"Carol ,  7",  // space around both name and count
		"Dave",        // no count, defaults to 1
		"Eve, ",       // empty count, defaults to 1
		"Frank, oops", // unparseable count, defaults to 1
		"   ",         // blank line, skipped
		"",            // empty line, skipped
		"Alice, 5",    // duplicate name, counts accumulate
	}, "\n")))
	c.NoError(err)
	c.Equal(105, m["Alice"]) // 100 + 5
	c.Equal(50, m["Bob"])
	c.Equal(7, m["Carol"])
	c.Equal(1, m["Dave"])
	c.Equal(1, m["Eve"])
	c.Equal(1, m["Frank"])
	c.Equal(6, len(m)) // no entry created for the blank/empty lines
}

func TestLoadFromReaderSuppressesNonPositiveCounts(t *testing.T) {
	c := check.New(t)
	// The namer constructors document that "any count less than 1 effectively removes the name from the set", so the
	// loader must honor an explicit non-positive count rather than silently treating it as 1.
	m, err := LoadFromReader(strings.NewReader(strings.Join([]string{
		"Keep, 3",      // positive count is retained
		"Zero, 0",      // explicit 0 suppresses the name
		"Neg, -5",      // negative count suppresses the name
		"Canceled, 5",  // accumulates to 0 with the next line, so it is suppressed
		"Canceled, -5", // cancels the previous line
		"Survivor, 5",  // accumulates above 0 with the next line, so it survives
		"Survivor, -2",
	}, "\n")))
	c.NoError(err)
	c.Equal(3, m["Keep"])
	c.Equal(3, m["Survivor"]) // 5 + (-2)
	_, ok := m["Zero"]
	c.False(ok, "an explicit count of 0 must remove the name")
	_, ok = m["Neg"]
	c.False(ok, "a negative count must remove the name")
	_, ok = m["Canceled"]
	c.False(ok, "a name whose counts cancel to 0 must be removed")
	c.Equal(2, len(m)) // only Keep and Survivor remain
}
