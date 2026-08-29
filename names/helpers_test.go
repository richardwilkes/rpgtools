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

// constRand is a Randomizer whose Intn always returns the same value, clamped to a valid index.
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

// seededRand is a deterministic Randomizer that, unlike constRand, produces a varied sequence, so the generated names
// depend on table ordering.
type seededRand struct{ r *rand.Rand }

func newSeededRand(seed uint64) *seededRand {
	return &seededRand{r: rand.New(rand.NewPCG(seed, seed))} //nolint:gosec // deterministic sequence is the point
}

func (s *seededRand) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	return s.r.IntN(n)
}

// blankWeighted and blankUnweighted contain only names that trim to nothing.
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

// recordingNamer captures the randomizer handed to GenerateNameWithRandomizer.
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
	// generateName must pass the shared default randomizer (xrand.New returns a singleton) and return the result.
	r := &recordingNamer{}
	c.Equal("recorded", r.GenerateName())
	c.True(r.got != nil, "generateName must pass a non-nil randomizer")
	c.True(r.got == xrand.New(), "generateName must pass the shared default randomizer")
}

func TestDefaultRandomizerWeightsUniformly(t *testing.T) {
	c := check.New(t)
	// Weighted selection is only as fair as the default randomizer's Intn. toolbox before v2.17.0 reduced a byte-width
	// draw with v % n, so with 200 equally weighted names the first 56 in table order were picked twice as often
	// (43.75% of draws instead of 28%).
	const total, block, draws = 200, 56, 200_000
	list := make([]string, total)
	for i := range list {
		list[i] = fmt.Sprintf("n%03d", i)
	}
	n := NewSimpleUnweightedNamer(list, false, false)
	// The table is sorted, so every name below n056 is in the first block.
	boundary := fmt.Sprintf("n%03d", block)
	inBlock := 0
	for range draws {
		if n.GenerateName() < boundary {
			inBlock++
		}
	}
	share := float64(inBlock) / draws
	// The fair share is 0.28 with a standard deviation of about 0.001, so 0.35 cannot flake in either direction.
	c.True(share < 0.35, "first %d of %d names received %.4f of draws; expected about 0.28", block, total, share)
}

func TestRandomBelowWithLimit(t *testing.T) {
	c := check.New(t)
	// A total that fits the Intn limit is a single draw.
	c.Equal(int64(3), randomBelowWithLimit(constRand(3), 7, 10))
	c.Equal(int64(0), randomBelowWithLimit(constRand(0), 10, 10))

	// Past the limit the draw is hi*limit+lo: with limit 10 and n 25, a randomizer fixed at 2 yields 22. One fixed at 9
	// yields 29 on every attempt, so the bounded rejection loop must fold it into range.
	c.Equal(int64(22), randomBelowWithLimit(constRand(2), 25, 10))
	c.Equal(int64(29%25), randomBelowWithLimit(constRand(9), 25, 10))

	// Each of the 25 values is expected 1,000 times with a standard deviation of about 31.
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
	// Every total fits a single Intn draw on this platform, so randomBelow must return Intn's result unchanged.
	for _, v := range []int{0, 1, 41} {
		c.Equal(int64(v), randomBelow(constRand(v), 42))
	}
}

func TestZeroValueNamersGenerateEmpty(t *testing.T) {
	c := check.New(t)
	// A zero-value namer has no data and must return "" rather than panic.
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
	// With a single source name, SimpleNamer and CompoundNamer are deterministic; the Markov namers must at least
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
	// Per-item weights 1,2,3; the draw v ranges over 1..6.
	entries := []int{1, 3, 6}

	got, ok := pickWeighted(entries, constRand(0), id) // v=1 -> first entry
	c.True(ok)
	c.Equal(1, got)
	got, ok = pickWeighted(entries, constRand(2), id) // v=3 -> second entry
	c.True(ok)
	c.Equal(3, got)
	got, ok = pickWeighted(entries, constRand(5), id) // v=6 -> last entry
	c.True(ok)
	c.Equal(6, got)

	_, ok = pickWeighted([]int{}, constRand(0), id)
	c.False(ok)
	_, ok = pickWeighted([]int{0}, constRand(0), id) // non-positive total
	c.False(ok)
}

// TestPickWeightedMatchesLinearScan pins the binary search to a reference linear scan for every draw value.
func TestPickWeightedMatchesLinearScan(t *testing.T) {
	c := check.New(t)
	id := func(v int) int64 { return int64(v) }
	// Per-item weights 3,1,4,1,5,9,2,6.
	entries := []int{3, 4, 8, 9, 14, 23, 25, 31}
	total := entries[len(entries)-1]
	for j := 1; j <= total; j++ {
		got, ok := pickWeighted(entries, constRand(j-1), id) // draw v = j
		c.True(ok, "draw %d must select something", j)
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

// TestPickWeightedSkipsZeroWeightEntries pins that a zero-weight entry, whose cumulative total duplicates the previous
// entry's, is never selected.
func TestPickWeightedSkipsZeroWeightEntries(t *testing.T) {
	c := check.New(t)
	type entry struct {
		name string
		cum  int64
	}
	cumOf := func(e entry) int64 { return e.cum }
	// Per-item weights 2,0,3.
	entries := []entry{{"a", 2}, {"zero", 2}, {"b", 5}}
	picked := make(map[string]bool)
	for j := 1; j <= 5; j++ {
		got, ok := pickWeighted(entries, constRand(j-1), cumOf)
		c.True(ok)
		picked[got.name] = true
	}
	// Draws 1..2 land on "a", 3..5 on "b".
	c.Equal(map[string]bool{"a": true, "b": true}, picked)
}

func TestAddWeightSaturates(t *testing.T) {
	c := check.New(t)
	c.Equal(5, addWeight(2, 3))
	c.Equal(maxWeight, addWeight(0, math.MaxInt))       // the delta is capped
	c.Equal(maxWeight, addWeight(maxWeight, 1))         // the sum saturates
	c.Equal(maxWeight, addWeight(maxWeight, maxWeight)) // without overflowing
	c.Equal(maxWeight-1, addWeight(maxWeight, -1))      // a negative delta passes through
}

func TestWeightsSaturateWithoutOverflow(t *testing.T) {
	c := check.New(t)
	const maxInt = int(^uint(0) >> 1)

	// Each per-name weight saturates at maxWeight and the grand total is their int64 sum, so pickWeighted still works.
	simple := NewSimpleNamer(map[string]int{"aaa": maxInt, "bbb": maxInt}, false, false)
	c.Equal(int64(maxWeight), simple.data[0].last)
	c.Equal(int64(maxWeight)*2, simple.data[len(simple.data)-1].last)
	c.True(simple.GenerateNameWithRandomizer(constRand(0)) != "", "valid data must still produce a name")

	// A transition weight saturates the same way.
	letter := NewMarkovLetterNamer(1, map[string]int{"a": maxInt, "b": maxInt}, false, false)
	steps := letter.mapping[letter.stepper.initialKey()]
	c.Equal(int64(maxWeight)*2, steps[len(steps)-1].last)
	c.True(letter.GenerateNameWithRandomizer(constRand(0)) != "", "valid data must still produce a name")
}

//nolint:goconst // The tests are more readable without constants for duplicated string
func TestUnweightedConstructorsCountDuplicates(t *testing.T) {
	c := check.New(t)
	// A repeated name must count once per occurrence.
	simple := NewSimpleUnweightedNamer([]string{"alice", "alice", "bob"}, false, false)
	c.Equal(int64(3), simple.data[len(simple.data)-1].last)
	aliceCount := int64(0)
	prev := int64(0)
	for _, ws := range simple.data {
		if ws.step == "alice" {
			aliceCount += ws.last - prev
		}
		prev = ws.last
	}
	c.Equal(int64(2), aliceCount)

	// The Markov length distribution must likewise accumulate the duplicate.
	letter := NewMarkovLetterUnweightedNamer(1, []string{"ab", "ab"}, false, false)
	c.Equal([]weightedStep[int]{{step: 2, last: 2}}, letter.lengths)
	run := NewMarkovRunUnweightedNamer([]string{"ab", "ab"}, false, false)
	c.Equal([]weightedStep[int]{{step: 2, last: 2}}, run.lengths)
}
