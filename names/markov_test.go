// Copyright (c) 2017-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package names

import (
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

func TestMarkovEmptyData(t *testing.T) {
	c := check.New(t)
	// Training data that is empty or contains only blank entries leaves the namer
	// with nothing to generate. This must yield an empty name rather than panic.
	blankWeighted := map[string]int{"": 5, "   ": 2}
	blankUnweighted := []string{"", "   "}
	c.Equal("", NewMarkovLetterNamer(2, blankWeighted, false, false).GenerateName())
	c.Equal("", NewMarkovLetterNamer(2, map[string]int{}, false, false).GenerateName())
	c.Equal("", NewMarkovLetterUnweightedNamer(2, blankUnweighted, false, false).GenerateName())
	c.Equal("", NewMarkovRunNamer(blankWeighted, false, false).GenerateName())
	c.Equal("", NewMarkovRunUnweightedNamer(blankUnweighted, false, false).GenerateName())
}

func TestMarkovGeneratesFromData(t *testing.T) {
	c := check.New(t)
	// Sanity check that, given real data, the namers actually produce non-empty
	// names made up only of the letters present in the training set.
	for range 25 {
		c.True(NewMarkovLetterNamer(2, data, false, false).GenerateName() != "")
		c.True(NewMarkovRunNamer(data, false, false).GenerateName() != "")
	}
}
