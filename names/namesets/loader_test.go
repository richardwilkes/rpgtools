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
