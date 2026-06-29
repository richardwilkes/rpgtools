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
	"math/rand/v2"
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

// seededRand is a deterministic Randomizer backed by a fixed seed. Unlike constRand it produces a varied sequence, so
// the value returned for a given draw lands on different cumulative-weight boundaries and the generated names depend on
// the order of the transition and length tables. This lets a test detect non-reproducible table ordering.
type seededRand struct{ r *rand.Rand }

// A deterministic, reproducible sequence is exactly what this test helper needs, so the weak generator is intentional.
func newSeededRand(seed uint64) *seededRand {
	return &seededRand{r: rand.New(rand.NewPCG(seed, seed))} //nolint:gosec // deterministic sequence is the point
}

func (s *seededRand) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	return s.r.IntN(n)
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
	id := func(v int) int64 { return int64(v) }
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

// TestPickWeightedMatchesLinearScan pins the binary-search selection to the reference linear scan it replaced: for every
// draw value v in 1..total, pickWeighted must return the first entry whose cumulative weight reaches v. Any off-by-one
// in the binary search would surface here as a boundary mismatch.
func TestPickWeightedMatchesLinearScan(t *testing.T) {
	c := check.New(t)
	id := func(v int) int64 { return int64(v) }
	// Cumulative weights for per-item weights 3,1,4,1,5,9,2,6 (total 31), so every entry has a distinct, varying width.
	entries := []int{3, 4, 8, 9, 14, 23, 25, 31}
	total := entries[len(entries)-1]
	for j := 1; j <= total; j++ {
		// constRand(j-1) forces Intn(total) to j-1, so pickWeighted's draw v becomes j.
		got, ok := pickWeighted(entries, constRand(j-1), id)
		c.True(ok, "draw %d must select something", j)
		// Reference: the first entry whose cumulative weight is >= j.
		want := -1
		for _, e := range entries {
			if e >= j {
				want = e
				break
			}
		}
		c.Equal(want, got, "draw %d", j)
	}
}

// TestPickWeightedSkipsZeroWeightEntries verifies the predicate-based binary search returns the first entry that reaches
// the draw value, so a zero-weight entry (one sharing the previous entry's cumulative total) is never selected even when
// it sits between two real entries. The entries carry a name distinct from their cumulative weight so the test can tell
// which index was actually chosen.
func TestPickWeightedSkipsZeroWeightEntries(t *testing.T) {
	c := check.New(t)
	type entry struct {
		name string
		cum  int64
	}
	cumOf := func(e entry) int64 { return e.cum }
	// Per-item weights 2,0,3: the middle "zero" entry has zero weight, so it duplicates the first entry's cumulative
	// total (2) while remaining a separate index.
	entries := []entry{{"a", 2}, {"zero", 2}, {"b", 5}}
	picked := make(map[string]bool)
	for j := 1; j <= 5; j++ {
		got, ok := pickWeighted(entries, constRand(j-1), cumOf)
		c.True(ok)
		picked[got.name] = true
	}
	// Draws 1..2 land on "a", 3..5 on "b"; the zero-weight "zero" entry (index 1) is unreachable.
	c.Equal(map[string]bool{"a": true, "b": true}, picked)
}

func TestWeightsSaturateWithoutOverflow(t *testing.T) {
	c := check.New(t)
	const maxInt = int(^uint(0) >> 1) // the platform int maximum, far beyond the maxWeight ceiling

	// SimpleNamer: each per-name weight saturates at maxWeight, and the grand cumulative total is their int64 sum
	// (positive and well within int64 range). Before this, summing two int-max weights overflowed to a negative total,
	// which made pickWeighted give up and return "" for entirely valid data.
	simple := NewSimpleNamer(map[string]int{"aaa": maxInt, "bbb": maxInt}, false, false)
	c.Equal(int64(maxWeight), simple.data[0].cumulative)                    // first name's own weight, capped
	c.Equal(int64(maxWeight)*2, simple.data[len(simple.data)-1].cumulative) // grand total, summed as int64
	c.True(simple.GenerateNameWithRandomizer(constRand(0)) != "", "valid data must still produce a name")

	// MarkovLetterNamer: a transition weight built from an enormous count saturates the same way, and the int64
	// cumulative total stays positive so the chain still generates.
	letter := NewMarkovLetterNamer(1, map[string]int{"a": maxInt, "b": maxInt}, false, false)
	steps := letter.mapping[letter.stepper.initialKey()]
	c.Equal(int64(maxWeight)*2, steps[len(steps)-1].last)
	c.True(letter.GenerateNameWithRandomizer(constRand(0)) != "", "valid data must still produce a name")
}

//nolint:goconst // The tests are more readable without constants for duplicated string
func TestUnweightedConstructorsCountDuplicates(t *testing.T) {
	c := check.New(t)
	// A name repeated in an unweighted slice must count once per occurrence rather than being collapsed to a single
	// entry, which is what a naive []string -> map adapter would do.
	simple := NewSimpleUnweightedNamer([]string{"alice", "alice", "bob"}, false, false)
	// The last entry's cumulative weight is the grand total: 1 each for the two alices and bob.
	c.Equal(int64(3), simple.data[len(simple.data)-1].cumulative)
	// Each entry's own weight is its cumulative minus the previous one; the two "alice" occurrences sum to 2 rather
	// than collapsing to a single entry of weight 1.
	aliceCount := int64(0)
	prev := int64(0)
	for _, nc := range simple.data {
		if nc.name == "alice" {
			aliceCount += nc.cumulative - prev
		}
		prev = nc.cumulative
	}
	c.Equal(int64(2), aliceCount)

	// The Markov length distribution must likewise accumulate the duplicate: two 2-rune names give a cumulative
	// count of 2 for length 2.
	letter := NewMarkovLetterUnweightedNamer(1, []string{"ab", "ab"}, false, false)
	c.Equal([]weightedStep[int]{{step: 2, last: 2}}, letter.lengths)
	run := NewMarkovRunUnweightedNamer([]string{"ab", "ab"}, false, false)
	c.Equal([]weightedStep[int]{{step: 2, last: 2}}, run.lengths)
}
