// Copyright (c) 2017-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calendar

import (
	"strconv"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

func TestWidthNeeded(t *testing.T) {
	c := check.New(t)
	// widthNeeded must equal the decimal digit count for every non-negative input (the only kind the callers pass),
	// including the powers-of-ten boundaries where the digit count rolls over.
	for _, count := range []int{0, 1, 9, 10, 11, 99, 100, 101, 999, 1000, 123456} {
		c.Equal(len(strconv.Itoa(count)), widthNeeded(count), "count %d", count)
	}
}
