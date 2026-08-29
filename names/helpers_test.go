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
	"fmt"
	"math"
	"math/rand/v2"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xrand"
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

// blankWeighted and blankUnweighted contain only names that trim to nothing, so a namer built from either has no usable
// data. Shared by the empty-data tests of every namer.
var (
	blankWeighted   = map[string]int{"": 5, "   ": 2}
	blankUnweighted = []string{"", "   "}
)

func TestApplyCase(t *testing.T) {
	c := check.New(t)
	c.Equal("hELLo", applyCase("hELLo", false, false)) // unchanged
	c.Equal("hello", applyCase("hELLo", true, false))  // lower-cased
	c.Equal("HELLo", applyCase("hELLo", false, true))  // first letter upper-cased
	c.Equal("Hello", applyCase("hELLo", true, true))   // lowered, then first letter upper-cased
}

// recordingNamer captures the randomizer handed to GenerateNameWithRandomizer so a test can confirm the shared
// generateName helper supplies a real one. Its GenerateName routes through generateName exactly as the real namers do.
type recordingNamer struct {
	got xrand.Randomizer
}

func (r *recordingNamer) GenerateName() string { return generateName(r) }

func (r *recordingNamer) GenerateNameWithRandomizer(rnd xrand.Randomizer) string {
	r.got = rnd
	return "recorded"
}

func TestGenerateNameSuppliesDefaultRandomizer(t *testing.T) {
	c := check.New(t)
	// generateName is the single place every Namer.GenerateName routes through; it must delegate to
	// GenerateNameWithRandomizer with the shared default randomizer (xrand.New returns a stateless singleton, so every
	// call is handed the same instance) and return that method's result.
	r := &recordingNamer{}
	c.Equal("recorded", r.GenerateName())
	c.True(r.got != nil, "generateName must pass a non-nil randomizer")
	c.True(r.got == xrand.New(), "generateName must pass the shared default randomizer")
}

func TestDefaultRandomizerWeightsUniformly(t *testing.T) {
	c := check.New(t)
	// Weighted selection is only as fair as the default randomizer's Intn. The toolbox xrand implementation before
	// v2.17.0 drew the smallest byte width that could hold n and reduced it with v % n, so with 200 equally weighted
	// names (a one-byte draw: 256 = 200 + 56) the first 56 names in table order were each picked twice as often as the
	// rest, landing 43.75% of draws in that block instead of 28%. Rejection sampling fixed that upstream; this pins the
	// dependency so a regression (or a downgrade) shows up here as a visibly skewed block share.
	const total, block, draws = 200, 56, 200_000
	list := make([]string, total)
	for i := range list {
		list[i] = fmt.Sprintf("n%03d", i)
	}
	n := NewSimpleUnweightedNamer(list, false, false)
	// The table is sorted, so n000..n055 occupy the first 'block' slots and every name below n056 is one of them.
	boundary := fmt.Sprintf("n%03d", block)
	inBlock := 0
	for range draws {
		if n.GenerateName() < boundary {
			inBlock++
		}
	}
	share := float64(inBlock) / draws
	// The fair share is 0.28 with a standard deviation of about 0.001 over this many draws, so a threshold of 0.35 sits
	// some 70 standard deviations from a fair generator and 80 from the biased one: it cannot flake in either direction.
	c.True(share < 0.35, "first %d of %d names received %.4f of draws; expected about 0.28", block, total, share)
}

func TestRandomBelowWithLimit(t *testing.T) {
	c := check.New(t)
	// A total that fits the Intn limit is a single draw.
	c.Equal(int64(3), randomBelowWithLimit(constRand(3), 7, 10))
	c.Equal(int64(0), randomBelowWithLimit(constRand(0), 10, 10))

	// Past the limit the draw is composed as hi*limit+lo: with limit 10 and n 25 there are 3 hi buckets, so a
	// randomizer fixed at 2 yields 2*10+2 = 22. One fixed at 9 is clamped to bucket 2 for hi and yields 29, which is out
	// of range on every attempt; the bounded rejection loop must then fold it into range rather than spin forever.
	c.Equal(int64(22), randomBelowWithLimit(constRand(2), 25, 10))
	c.Equal(int64(29%25), randomBelowWithLimit(constRand(9), 25, 10))

	// Every value below n must be reachable and roughly equally likely: over 25,000 draws each of the 25 values is
	// expected 1,000 times with a standard deviation of about 31, so bounds of 800 and 1,200 sit over six deviations out.
	const n, draws = 25, 25_000
	counts := make([]int, n)
	rnd := newSeededRand(5)
	for range draws {
		v := randomBelowWithLimit(rnd, n, 10)
		c.True(v >= 0 && v < n, "draw %d is out of range", v)
		counts[v]++
	}
	for v, count := range counts {
		c.True(count > 800 && count < 1200, "value %d drawn %d times, expected about 1000", v, count)
	}
}

func TestRandomBelowMatchesSingleDraw(t *testing.T) {
	c := check.New(t)
	// On this platform every cumulative total the namers can produce fits a single Intn draw, so randomBelow must hand
	// the total straight to Intn and return its result unchanged.
	for _, v := range []int{0, 1, 41} {
		c.Equal(int64(v), randomBelow(constRand(v), 42))
	}
}

func TestZeroValueNamersGenerateEmpty(t *testing.T) {
	c := check.New(t)
	// A namer that was never built by a constructor has no data to draw from. Every implementation must treat that as
	// "nothing to generate" and return "" rather than panic, so the zero value of each is usable just as a zero
	// dice.Roller is. The Markov namers embed their core by value for exactly this reason: a nil embedded pointer would
	// panic on the first field access.
	var (
		simple   SimpleNamer
		compound CompoundNamer
		letter   MarkovLetterNamer
		run      MarkovRunNamer
	)
	for i, namer := range []Namer{&simple, &compound, &letter, &run} {
		c.NotPanics(func() {
			c.Equal("", namer.GenerateName(), "namer index %d", i)
			c.Equal("", namer.GenerateNameWithRandomizer(constRand(0)), "namer index %d", i)
		}, "namer index %d", i)
	}
}

func TestGenerateNameAcrossNamers(t *testing.T) {
	c := check.New(t)
	// Every concrete Namer.GenerateName delegates through the shared generateName helper and must still produce a valid
	// name. With a single source name, SimpleNamer and CompoundNamer are deterministic; the Markov namers must at least
	// produce a non-empty name.
	const name = "solo"
	c.Equal(name, NewSimpleNamer(map[string]int{name: 1}, false, false).GenerateName())
	c.Equal(name, NewCompoundNamer(" ", false, false, NewSimpleNamer(map[string]int{name: 1}, false, false)).GenerateName())
	c.True(NewMarkovLetterNamer(2, map[string]int{name: 1}, false, false).GenerateName() != "",
		"MarkovLetterNamer.GenerateName must produce a name")
	c.True(NewMarkovRunNamer(map[string]int{name: 1}, false, false).GenerateName() != "",
		"MarkovRunNamer.GenerateName must produce a name")
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

func TestAddWeightSaturates(t *testing.T) {
	c := check.New(t)
	// addWeight caps each delta at maxWeight (via the builtin min) and saturates the running total there, so no
	// accumulation can exceed the ceiling or overflow a platform int. A negative delta passes through unchanged.
	c.Equal(5, addWeight(2, 3))                         // ordinary addition, well under the ceiling
	c.Equal(maxWeight, addWeight(0, math.MaxInt))       // the largest possible delta is capped to the ceiling
	c.Equal(maxWeight, addWeight(maxWeight, 1))         // a ceiling sum plus more stays at the ceiling
	c.Equal(maxWeight, addWeight(maxWeight, maxWeight)) // two ceiling weights saturate rather than overflow
	c.Equal(maxWeight-1, addWeight(maxWeight, -1))      // a negative delta is left as-is (min keeps it)
}

func TestWeightsSaturateWithoutOverflow(t *testing.T) {
	c := check.New(t)
	const maxInt = int(^uint(0) >> 1) // the platform int maximum, far beyond the maxWeight ceiling

	// SimpleNamer: each per-name weight saturates at maxWeight, and the grand cumulative total is their int64 sum
	// (positive and well within int64 range). Before this, summing two int-max weights overflowed to a negative total,
	// which made pickWeighted give up and return "" for entirely valid data.
	simple := NewSimpleNamer(map[string]int{"aaa": maxInt, "bbb": maxInt}, false, false)
	c.Equal(int64(maxWeight), simple.data[0].last)                    // first name's own weight, capped
	c.Equal(int64(maxWeight)*2, simple.data[len(simple.data)-1].last) // grand total, summed as int64
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
	c.Equal(int64(3), simple.data[len(simple.data)-1].last)
	// Each entry's own weight is its cumulative minus the previous one; the two "alice" occurrences sum to 2 rather
	// than collapsing to a single entry of weight 1.
	aliceCount := int64(0)
	prev := int64(0)
	for _, ws := range simple.data {
		if ws.step == "alice" {
			aliceCount += ws.last - prev
		}
		prev = ws.last
	}
	c.Equal(int64(2), aliceCount)

	// The Markov length distribution must likewise accumulate the duplicate: two 2-rune names give a cumulative
	// count of 2 for length 2.
	letter := NewMarkovLetterUnweightedNamer(1, []string{"ab", "ab"}, false, false)
	c.Equal([]weightedStep[int]{{step: 2, last: 2}}, letter.lengths)
	run := NewMarkovRunUnweightedNamer([]string{"ab", "ab"}, false, false)
	c.Equal([]weightedStep[int]{{step: 2, last: 2}}, run.lengths)
}
