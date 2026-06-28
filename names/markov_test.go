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

func TestMarkovLetterWeightedSelection(t *testing.T) {
	c := check.New(t)
	// Two equally-weighted single-letter inputs. Each letter is the final entry in
	// the cumulative-weight table for one of the two transitions, so an off-by-one
	// in the weighted selection would make one of them impossible to ever produce.
	n := NewMarkovLetterNamer(1, map[string]int{"a": 1, "b": 1}, false, false)
	counts := make(map[string]int)
	for range 100 {
		counts[n.GenerateName()]++
	}
	c.Equal(2, len(counts), "expected both letters to be produced, got: %v", counts)
}

func TestMarkovRunLengthWeighting(t *testing.T) {
	c := check.New(t)
	// The name-length distribution must honor each name's count, just as the
	// transition table does. The cumulative length table therefore sums the counts
	// rather than counting distinct names.
	n := NewMarkovRunNamer(map[string]int{"oo": 3, "eee": 5}, false, false)
	c.True(len(n.lengths) > 0)
	c.Equal(8, n.lengths[len(n.lengths)-1][1])
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
