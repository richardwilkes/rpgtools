// Copyright (c) 2017-2026 by Richard A. Wilkes. All rights reserved.
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

// constRand is a deterministic Randomizer whose Intn always returns the same value (clamped to a valid index), letting
// the weighted-pick boundaries be exercised exactly.
type constRand int

func (c constRand) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	if v := int(c); v < n {
		return v
	}
	return n - 1
}

func TestApplyCase(t *testing.T) {
	c := check.New(t)
	c.Equal("hELLo", applyCase("hELLo", false, false)) // unchanged
	c.Equal("hello", applyCase("hELLo", true, false))  // lower-cased
	c.Equal("HELLo", applyCase("hELLo", false, true))  // first letter upper-cased
	c.Equal("Hello", applyCase("hELLo", true, true))   // lowered, then first letter upper-cased
}

func TestPickWeighted(t *testing.T) {
	c := check.New(t)
	id := func(v int) int { return v }
	// Cumulative weights [1,3,6] (per-item 1,2,3; total 6). 'v' ranges over 1..6.
	entries := []int{1, 3, 6}

	got, ok := pickWeighted(entries, constRand(0), id) // v=1 -> first entry
	c.True(ok)
	c.Equal(1, got)
	got, ok = pickWeighted(entries, constRand(2), id) // v=3 -> second entry
	c.True(ok)
	c.Equal(3, got)
	got, ok = pickWeighted(entries, constRand(5), id) // v=6 -> last entry (the off-by-one boundary)
	c.True(ok)
	c.Equal(6, got)

	// Nothing to pick from.
	_, ok = pickWeighted([]int{}, constRand(0), id)
	c.False(ok)
	// A non-positive grand total cannot be selected from.
	_, ok = pickWeighted([]int{0}, constRand(0), id)
	c.False(ok)
}

//nolint:goconst // The tests are more readable without constants for duplicated string
func TestUnweightedConstructorsCountDuplicates(t *testing.T) {
	c := check.New(t)
	// A name repeated in an unweighted slice must count once per occurrence rather than being collapsed to a single
	// entry, which is what a naive []string -> map adapter would do.
	simple := NewSimpleUnweightedNamer([]string{"alice", "alice", "bob"}, false, false)
	c.Equal(3, simple.total)
	aliceCount := 0
	for _, nc := range simple.data {
		if nc.name == "alice" {
			aliceCount += nc.count
		}
	}
	c.Equal(2, aliceCount)

	// The Markov length distribution must likewise accumulate the duplicate: two 2-rune names give a cumulative
	// count of 2 for length 2.
	letter := NewMarkovLetterUnweightedNamer(1, []string{"ab", "ab"}, false, false)
	c.Equal([][2]int{{2, 2}}, letter.lengths)
	run := NewMarkovRunUnweightedNamer([]string{"ab", "ab"}, false, false)
	c.Equal([][2]int{{2, 2}}, run.lengths)
}
