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
	"math"
	"testing"
	"unicode/utf8"

	"github.com/richardwilkes/toolbox/v2/check"
)

func TestMarkovEmptyData(t *testing.T) {
	c := check.New(t)
	// Training data that is empty or contains only blank entries leaves the namer
	// with nothing to generate. This must yield an empty name rather than panic.
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

func TestMarkovLetterDepthClampsToOne(t *testing.T) {
	c := check.New(t)
	// A depth below 1 would leave the sliding window empty and growing rather than sliding, so the constructors clamp
	// it to 1. Namers built with a zero or negative depth must therefore be indistinguishable from one built with depth
	// 1: same stepper, same transition table (including the initial key), and the same names from the same seed.
	weighted := map[string]int{"quill": 1, "raven": 2, "sable": 3}
	unweighted := []string{"quill", "raven", "sable"}
	const seed, samples = 7, 30
	generate := func(n *MarkovLetterNamer) []string {
		rnd := newSeededRand(seed)
		out := make([]string, samples)
		for i := range out {
			out[i] = n.GenerateNameWithRandomizer(rnd)
		}
		return out
	}
	wantWeighted := NewMarkovLetterNamer(1, weighted, false, false)
	wantUnweighted := NewMarkovLetterUnweightedNamer(1, unweighted, false, false)
	for _, depth := range []int{0, -3} {
		got := NewMarkovLetterNamer(depth, weighted, false, false)
		c.Equal(letterStepper{depth: 1}, got.stepper, "weighted depth %d", depth)
		c.Equal(wantWeighted.mapping, got.mapping, "weighted depth %d", depth)
		c.Equal(generate(wantWeighted), generate(got), "weighted depth %d", depth)

		got = NewMarkovLetterUnweightedNamer(depth, unweighted, false, false)
		c.Equal(letterStepper{depth: 1}, got.stepper, "unweighted depth %d", depth)
		c.Equal(wantUnweighted.mapping, got.mapping, "unweighted depth %d", depth)
		c.Equal(generate(wantUnweighted), generate(got), "unweighted depth %d", depth)
	}
}

func TestMarkovRunLengthWeighting(t *testing.T) {
	c := check.New(t)
	// The name-length distribution must honor each name's count, just as the
	// transition table does. The cumulative length table therefore sums the counts
	// rather than counting distinct names.
	n := NewMarkovRunNamer(map[string]int{"oo": 3, "eee": 5}, false, false)
	c.True(len(n.lengths) > 0)
	c.Equal(int64(8), n.lengths[len(n.lengths)-1].last)
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
	// Sanity check that, given real data, the namers actually produce non-empty names made up only of the letters
	// present in the training set. The rune check is what catches a chain emitting something it was never trained on,
	// such as a start-key sentinel leaking into the output.
	trained := make(map[rune]bool)
	for name := range data {
		for _, ch := range name {
			trained[ch] = true
		}
	}
	letter := NewMarkovLetterNamer(2, data, false, false)
	run := NewMarkovRunNamer(data, false, false)
	for range 25 {
		for _, name := range []string{letter.GenerateName(), run.GenerateName()} {
			c.True(name != "", "generated name must not be empty")
			for _, ch := range name {
				c.True(trained[ch], "generated %q contains %q, which is not in the training data", name, ch)
			}
		}
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
	c.Equal(5, letter.lengths[0].step, "letter namer length must be counted in runes, not bytes")

	run := NewMarkovRunNamer(map[string]int{name: 1}, false, false)
	c.Equal(1, len(run.lengths))
	c.Equal(5, run.lengths[0].step, "run namer length must be counted in runes, not bytes")
}

func TestMarkovGenerationHasHardCap(t *testing.T) {
	c := check.New(t)
	// Hand-built models whose transition graph is an endless cycle (a->b->a / "a"->"b"->"a") with an empty final
	// set. The generation loop only stops on a dead-end key, an empty token, or a final token past 'maximum', so
	// without a hard cap it would spin forever. The cap bounds the result at twice the longest training length
	// (2*4 = 8 here).
	letter := &MarkovLetterNamer{markov[rune]{
		stepper: letterStepper{depth: 1},
		mapping: map[string][]weightedStep[rune]{
			startMarker: {{step: 'a', last: 1}},
			"a":         {{step: 'b', last: 1}},
			"b":         {{step: 'a', last: 1}},
		},
		final:     map[rune]struct{}{},
		lengths:   []weightedStep[int]{{step: 4, last: 1}},
		maxLength: 4,
	}}
	c.Equal(8, utf8.RuneCountInString(letter.GenerateName()), "letter namer must stop at the hard cap")

	run := &MarkovRunNamer{markov[string]{
		stepper: runStepper{},
		mapping: map[string][]weightedStep[string]{
			"":  {{step: "a", last: 1}},
			"a": {{step: "b", last: 1}},
			"b": {{step: "a", last: 1}},
		},
		final:     map[string]struct{}{},
		lengths:   []weightedStep[int]{{step: 4, last: 1}},
		maxLength: 4,
	}}
	c.Equal(8, utf8.RuneCountInString(run.GenerateName()), "run namer must stop at the hard cap")

	// A run is appended whole, so the cap has to be checked before a run is taken rather than after: this cycle of
	// 3-rune runs would otherwise overshoot the 8-rune cap and produce 9 runes. It must stop at 6 instead.
	const wideRun = "abc"
	wide := &MarkovRunNamer{markov[string]{
		stepper: runStepper{},
		mapping: map[string][]weightedStep[string]{
			"":      {{step: wideRun, last: 1}},
			wideRun: {{step: wideRun, last: 1}},
		},
		final:     map[string]struct{}{},
		lengths:   []weightedStep[int]{{step: 4, last: 1}},
		maxLength: 4,
	}}
	c.Equal(wideRun+wideRun, wide.GenerateName(), "a run that would carry the name past the hard cap must not be taken")
}

func TestMarkovEndingsTable(t *testing.T) {
	c := check.New(t)
	// Training on "kf" and "knkf" (depth 1) gives k the successors f (twice, final) and n (once, not final), n the sole
	// successor k, and the start key the sole successor k. The endings table must hold, for every key, only the
	// successors that reach a final step soonest: f for k (a final step itself), and k for both n and the start key
	// (one step from k's final successor). The full transition table keeps every successor.
	n := NewMarkovLetterNamer(1, map[string]int{"kf": 1, "knkf": 1}, false, false)
	c.Equal(map[string][]weightedStep[rune]{
		startMarker: {{step: 'k', last: 2}},
		"k":         {{step: 'f', last: 2}, {step: 'n', last: 3}},
		"n":         {{step: 'k', last: 1}},
	}, n.mapping)
	c.Equal(map[string][]weightedStep[rune]{
		startMarker: {{step: 'k', last: 2}},
		"k":         {{step: 'f', last: 2}},
		"n":         {{step: 'k', last: 1}},
	}, n.endings)

	// The run namer's table is built the same way: "ka" and "kunkunka" give k and nk the successors a (final) and u, and
	// u the sole successor nk, so past the target length k and nk must head straight for a and u must go through nk.
	r := NewMarkovRunNamer(map[string]int{"ka": 1, "kunkunka": 1}, false, false)
	c.Equal(map[string][]weightedStep[string]{
		"":   {{step: "k", last: 2}},
		"k":  {{step: "a", last: 1}},
		"u":  {{step: "nk", last: 2}},
		"nk": {{step: "a", last: 1}},
	}, r.endings)
}

func TestMarkovEndsAtFirstOpportunityPastTarget(t *testing.T) {
	c := check.New(t)
	// A randomizer that always takes the last entry of every table picks the longest length and, among a key's
	// successors, the last in sorted order. Trained on "kf" and "knkf", that walks k, n, k, n and reaches the 4-rune
	// target on n, which is not final. From there the chain must follow the endings table (n -> k -> f) and stop at
	// "knknkf" rather than keep taking n, the last of k's successors, until the 8-rune hard cap cuts it off mid-word.
	last := constRand(math.MaxInt)
	letter := NewMarkovLetterNamer(1, map[string]int{"kf": 1, "knkf": 1}, false, false)
	c.Equal("knknkf", letter.GenerateNameWithRandomizer(last))

	// Likewise for runs: "ka" and "kunkunka" target 8 runes, reached on u (not final) after k, u, nk, u, nk, u. The
	// endings table (u -> nk -> a) must end the name at "kunkunkunka" instead of cycling u, nk up to the 16-rune cap.
	run := NewMarkovRunNamer(map[string]int{"ka": 1, "kunkunka": 1}, false, false)
	c.Equal("kunkunkunka", run.GenerateNameWithRandomizer(last))
}

func TestMarkovRealDataNeverReachesHardCap(t *testing.T) {
	c := check.New(t)
	// Every key a trained table contains can reach a final step, so once a name has reached its target length the
	// endings table brings it to a natural end within one training name's worth of further steps and the hard cap
	// (twice the longest training name) is never what stops it. These small corpora used to hit the cap in a measurable
	// share of draws (the letter chain has an l -> l self-loop and the run chain cycles through n, iyo) and emitted
	// truncations such as "cetttettttttttte" and "verryaniyoniyoniya"; the generated name must now always end on a
	// final step and stay strictly under the cap.
	letter := NewMarkovLetterUnweightedNamer(1, []string{"allisan", "kimette", "kaiyanna", "brandace", "cherisse", "morene"},
		false, false)
	run := NewMarkovRunUnweightedNamer([]string{
		"veronice", "niyona", "aulona", "jaelei", "terryana", "shley", "jaliana",
		"niyari", "aminaa", "mber",
	}, false, false)
	rnd := newSeededRand(99)
	for range 20_000 {
		name := letter.GenerateNameWithRandomizer(rnd)
		c.True(utf8.RuneCountInString(name) < 2*letter.maxLength, "letter namer hit the hard cap with %q", name)
		lastRune, _ := utf8.DecodeLastRuneInString(name)
		_, final := letter.final[lastRune]
		c.True(final, "letter namer produced %q, which does not end on a final step", name)

		name = run.GenerateNameWithRandomizer(rnd)
		c.True(utf8.RuneCountInString(name) < 2*run.maxLength, "run namer hit the hard cap with %q", name)
		runs := decompose(name)
		_, final = run.final[runs[len(runs)-1]]
		c.True(final, "run namer produced %q, which does not end on a final run", name)
	}
}

func TestMarkovLetterStartKeyCannotAlias(t *testing.T) {
	c := check.New(t)
	// The start key is an invalid UTF-8 byte rather than a run of NUL runes, so a training name that contains U+0000 no
	// longer aliases the start state. With a NUL sentinel this name taught the chain a NUL self-transition on the start
	// key and generated runs of NULs up to the hard cap; now the only path is the name itself.
	const name = "\x00zzz"
	n := NewMarkovLetterNamer(1, map[string]int{name: 1}, false, false)
	for range 50 {
		c.Equal(name, n.GenerateName())
	}
	c.Equal(map[string][]weightedStep[rune]{
		startMarker: {{step: 0, last: 1}},
		"\x00":      {{step: 'z', last: 1}},
		"z":         {{step: 'z', last: 2}},
	}, n.mapping)
}

func TestMarkovLetterKeyWindow(t *testing.T) {
	c := check.New(t)
	// The initial key is the constant start marker, which carries no depth-sized padding to allocate. While the window
	// is filling, the key is the marker plus the runes taken so far; once depth runes have been taken the marker drops
	// off and the key slides one rune at a time, in runes rather than bytes.
	s := letterStepper{depth: 3}
	key := s.initialKey()
	c.Equal(startMarker, key)
	key = s.advance(key, 'a')
	c.Equal(startMarker+"a", key)
	key = s.advance(key, 'α')
	c.Equal(startMarker+"aα", key)
	key = s.advance(key, 'c')
	c.Equal("aαc", key)
	key = s.advance(key, 'd')
	c.Equal("αcd", key)
	key = s.advance(key, 'e')
	c.Equal("cde", key)

	// With depth 1 the marker drops off on the very first step.
	s = letterStepper{depth: 1}
	c.Equal("a", s.advance(s.initialKey(), 'a'))
}

func TestMarkovLetterHugeDepth(t *testing.T) {
	c := check.New(t)
	// A depth of at least the longest training name reproduces training names verbatim, and a huge one must behave the
	// same rather than allocating a depth-sized key or panicking (math.MaxInt used to fail inside make([]rune, depth)).
	const name = "abc"
	for _, depth := range []int{3, 64, math.MaxInt} {
		var n *MarkovLetterNamer
		c.NotPanics(func() {
			n = NewMarkovLetterNamer(depth, map[string]int{name: 1}, false, false)
		}, "depth %d", depth)
		c.Equal(name, n.GenerateNameWithRandomizer(constRand(0)), "depth %d", depth)
		c.Equal(name, n.GenerateName(), "depth %d", depth)
	}
}

func TestDecompose(t *testing.T) {
	c := check.New(t)
	// Pins the run decomposition the whole MarkovRunNamer rests on: maximal runs of vowels (y counts as one), of
	// consonants, or of non-letters, with the very first rune setting the class of the first run.
	for _, tc := range []struct {
		in   string
		want []string
	}{
		{"a", []string{"a"}},
		{"b", []string{"b"}},
		{"aba", []string{"a", "b", "a"}},
		{"bab", []string{"b", "a", "b"}},
		{"strength", []string{"str", "e", "ngth"}},
		{"AEIOU", []string{"AEIOU"}},
		{"Mary", []string{"M", "a", "r", "y"}},
		{"Zoë", []string{"Z", "oë"}},
		// Non-letters form runs of their own instead of merging into the adjacent consonant run.
		{"O'Brien", []string{"O", "'", "Br", "ie", "n"}},
		{"Mary Ann", []string{"M", "a", "r", "y", " ", "A", "nn"}},
		{"Jean-Luc", []string{"J", "ea", "n", "-", "L", "u", "c"}},
		{"R2D2", []string{"R", "2", "D", "2"}},
		{"--", []string{"--"}},
		// A combining mark stays with the rune it modifies, so a decomposed accent behaves like a precomposed one.
		{"e\u0301l", []string{"e\u0301", "l"}},
		{"n\u0303a", []string{"n\u0303", "a"}},
		// A mark with nothing before it has no class to inherit and is a non-letter run.
		{"\u0301a", []string{"\u0301", "a"}},
	} {
		c.Equal(tc.want, decompose(tc.in), "decompose(%q)", tc.in)
	}
	c.Equal(0, len(decompose("")))
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
