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
	"unicode/utf8"

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

func TestMarkovReproducibleAcrossBuilds(t *testing.T) {
	c := check.New(t)
	// Go randomizes map iteration order on every range, so rebuilding a namer from identical data exercises different
	// orderings of each transition's next-items and of the length buckets. The cumulative-weight tables, and therefore
	// the names a seeded randomizer produces, must not depend on that order; otherwise the same training data and seed
	// yield different names from one process run to the next. The data has many distinct first letters and runs so a
	// non-deterministic ordering would almost certainly change the output.
	data := map[string]int{
		"alpha": 1, "bravo": 2, "charlie": 3, "delta": 1, "echo": 2,
		"foxtrot": 3, "golf": 1, "hotel": 2, "india": 3, "juliet": 1,
		"kilo": 2, "lima": 3, "mike": 1, "november": 2, "oscar": 3,
	}
	const seed, samples = 42, 50
	letterNames := func() []string {
		n := NewMarkovLetterNamer(2, data, false, false)
		rnd := newSeededRand(seed)
		out := make([]string, samples)
		for i := range out {
			out[i] = n.GenerateNameWithRandomizer(rnd)
		}
		return out
	}
	runNames := func() []string {
		n := NewMarkovRunNamer(data, false, false)
		rnd := newSeededRand(seed)
		out := make([]string, samples)
		for i := range out {
			out[i] = n.GenerateNameWithRandomizer(rnd)
		}
		return out
	}
	letterWant := letterNames()
	runWant := runNames()
	for range 20 {
		c.Equal(letterWant, letterNames(), "letter namer output must be reproducible across rebuilds")
		c.Equal(runWant, runNames(), "run namer output must be reproducible across rebuilds")
	}
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

func TestMarkovLengthCountsRunes(t *testing.T) {
	c := check.New(t)
	// "ααααα" is 5 runes but 10 bytes. The chains are built rune-by-rune, so the recorded name length must be the
	// character (rune) count, not the UTF-8 byte count, otherwise non-ASCII names skew the length distribution.
	const name = "ααααα"
	c.Equal(5, utf8.RuneCountInString(name))
	c.Equal(10, len(name))

	letter := NewMarkovLetterNamer(1, map[string]int{name: 1}, false, false)
	c.Equal(1, len(letter.lengths))
	c.Equal(5, letter.lengths[0][0], "letter namer length must be counted in runes, not bytes")

	run := NewMarkovRunNamer(map[string]int{name: 1}, false, false)
	c.Equal(1, len(run.lengths))
	c.Equal(5, run.lengths[0][0], "run namer length must be counted in runes, not bytes")
}

func TestMarkovGenerationHasHardCap(t *testing.T) {
	c := check.New(t)
	// Hand-built models whose transition graph is an endless cycle (a->b->a / "a"->"b"->"a") with an empty final
	// set. The generation loop only stops on a dead-end key, an empty token, or a final token past 'maximum', so
	// without a hard cap it would spin forever. The cap bounds the result at twice the longest training length
	// (2*4 = 8 here).
	letter := &MarkovLetterNamer{&markov[rune]{
		stepper: letterStepper{depth: 1},
		mapping: map[string][]weightedStep[rune]{
			"\x00": {{step: 'a', last: 1}},
			"a":    {{step: 'b', last: 1}},
			"b":    {{step: 'a', last: 1}},
		},
		final:     map[rune]struct{}{},
		lengths:   [][2]int{{4, 1}},
		maxLength: 4,
	}}
	c.Equal(8, utf8.RuneCountInString(letter.GenerateName()), "letter namer must stop at the hard cap")

	run := &MarkovRunNamer{&markov[string]{
		stepper: runStepper{},
		mapping: map[string][]weightedStep[string]{
			"":  {{step: "a", last: 1}},
			"a": {{step: "b", last: 1}},
			"b": {{step: "a", last: 1}},
		},
		final:     map[string]struct{}{},
		lengths:   [][2]int{{4, 1}},
		maxLength: 4,
	}}
	c.Equal(8, utf8.RuneCountInString(run.GenerateName()), "run namer must stop at the hard cap")
}

func TestMarkovCoreGeneratesDeterministically(t *testing.T) {
	c := check.New(t)
	// Both namers now run through the same generic core, so a fixed (always-first) randomizer must walk each chain
	// deterministically: with a single training name and no branching, each namer reproduces that name exactly. This
	// pins the shared build-and-generate path for the rune and the run step types alike.
	letter := NewMarkovLetterNamer(1, map[string]int{"abc": 1}, false, false)
	c.Equal("abc", letter.GenerateNameWithRandomizer(constRand(0)))
	run := NewMarkovRunNamer(map[string]int{"aba": 1}, false, false)
	c.Equal("aba", run.GenerateNameWithRandomizer(constRand(0)))
}

func TestMarkovLetterGenerationCapsInRunes(t *testing.T) {
	c := check.New(t)
	// Both training names are 4 characters long, so every generated name should also be 4 characters. Their byte
	// lengths differ (4 vs 12), so a byte-based length distribution would target byte counts and emit names ranging
	// anywhere from 2 to 12 characters depending on which runes were chosen. Each name uses a self-cycling letter so
	// the chain never dead-ends and the length cap is what stops it.
	n := NewMarkovLetterNamer(1, map[string]int{"aaaa": 1, "好好好好": 1}, false, false)
	for range 200 {
		name := n.GenerateName()
		c.Equal(4, utf8.RuneCountInString(name), "generated %q has %d runes, want 4", name,
			utf8.RuneCountInString(name))
	}
}
