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

var data = map[string]int{
	"aA": 1,
	"bB": 1,
}

func TestSimple(t *testing.T) {
	c := check.New(t)
	s := NewSimpleNamer(data, false, false)
	counts := make(map[string]int)
	for range 25 {
		counts[s.GenerateName()]++
	}
	c.Equal(2, len(counts))
	_, exists := counts["aA"]
	c.True(exists, "expecting to find 'aA' in: %v", counts)
	_, exists = counts["bB"]
	c.True(exists, "expecting to find 'bB' in: %v", counts)
}

func TestSimpleBuildsCumulativeWeightedSteps(t *testing.T) {
	c := check.New(t)
	// SimpleNamer stores its weighted names in the same weightedStep[string] the Markov namers use, rather than a
	// bespoke pair type. Building from a set with distinct counts must yield those names in sorted order, each paired
	// with the running cumulative weight (so the final entry holds the grand total) -- the exact shape pickWeighted
	// consumes.
	s := NewSimpleNamer(map[string]int{"coral": 3, "amber": 2, "ivory": 1}, false, false)
	c.Equal([]weightedStep[string]{
		{step: "amber", last: 2},
		{step: "coral", last: 5},
		{step: "ivory", last: 6},
	}, s.data)
}

func TestSimpleReproducibleAcrossBuilds(t *testing.T) {
	c := check.New(t)
	// Go randomizes map iteration order on every range, so rebuilding a SimpleNamer from identical data exercises
	// different insertion orders. The cumulative-weight table, and therefore the names a seeded randomizer produces,
	// must not depend on that order. Routing the build through the shared cumulativeWeights helper (which sorts the
	// names) preserves this guarantee; the data has many distinct names so a non-deterministic order would change it.
	reproData := map[string]int{
		"alpha": 1, "bravo": 2, "charlie": 3, "delta": 1, "echo": 2,
		"foxtrot": 3, "golf": 1, "hotel": 2, "india": 3, "juliet": 1,
	}
	const seed, samples = 42, 50
	names := func() []string {
		n := NewSimpleNamer(reproData, false, false)
		rnd := newSeededRand(seed)
		out := make([]string, samples)
		for i := range out {
			out[i] = n.GenerateNameWithRandomizer(rnd)
		}
		return out
	}
	want := names()
	for range 20 {
		c.Equal(want, names(), "SimpleNamer output must be reproducible across rebuilds")
	}
}

func TestSimpleLowered(t *testing.T) {
	c := check.New(t)
	s := NewSimpleNamer(data, true, false)
	counts := make(map[string]int)
	for range 25 {
		counts[s.GenerateName()]++
	}
	c.Equal(2, len(counts))
	_, exists := counts["aa"]
	c.True(exists, "expecting to find 'aa' in: %v", counts)
	_, exists = counts["bb"]
	c.True(exists, "expecting to find 'bb' in: %v", counts)
}

func TestSimpleFirstUpper(t *testing.T) {
	c := check.New(t)
	s := NewSimpleNamer(data, false, true)
	counts := make(map[string]int)
	for range 25 {
		counts[s.GenerateName()]++
	}
	c.Equal(2, len(counts))
	_, exists := counts["AA"]
	c.True(exists, "expecting to find 'AA' in: %v", counts)
	_, exists = counts["BB"]
	c.True(exists, "expecting to find 'BB' in: %v", counts)
}

func TestSimpleLoweredAndFirstUpper(t *testing.T) {
	c := check.New(t)
	s := NewSimpleNamer(data, true, true)
	counts := make(map[string]int)
	for range 25 {
		counts[s.GenerateName()]++
	}
	c.Equal(2, len(counts))
	_, exists := counts["Aa"]
	c.True(exists, "expecting to find 'Aa' in: %v", counts)
	_, exists = counts["Bb"]
	c.True(exists, "expecting to find 'Bb' in: %v", counts)
}
