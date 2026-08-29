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
	"cmp"
	"iter"
	"maps"
	"math"
	"slices"
	"strings"

	"github.com/richardwilkes/toolbox/v2/xrand"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

// maxWeight caps an individual name or transition weight, so the int64 cumulative totals cannot overflow no matter how
// many entries are summed.
const maxWeight = math.MaxInt32

// addWeight returns sum + delta saturated at maxWeight, capping delta first so the addition cannot overflow. A negative
// delta passes through unchanged; callers drop non-positive counts before they reach a weighted table.
func addWeight(sum, delta int) int {
	delta = min(delta, maxWeight)
	if sum > maxWeight-delta {
		return maxWeight
	}
	return sum + delta
}

// Namer defines the methods required of a name generator.
type Namer interface {
	// GenerateName generates a new random name.
	GenerateName() string
	// GenerateNameWithRandomizer generates a new random name using the specified randomizer.
	GenerateNameWithRandomizer(rnd xrand.Randomizer) string
}

// generateName produces a name from n using the default randomizer, which every GenerateName delegates to. xrand.New
// returns a stateless singleton that is safe for concurrent use, so nothing is allocated per call.
func generateName(n Namer) string {
	return n.GenerateNameWithRandomizer(xrand.New())
}

// applyCase applies the case transformations shared by every namer.
func applyCase(in string, lowered, firstToUpper bool) string {
	if lowered {
		in = strings.ToLower(in)
	}
	if firstToUpper {
		in = xstrings.FirstToUpper(in)
	}
	return in
}

// unweighted adapts a slice of names into a (name, count) sequence in which each occurrence counts once. This lets the
// unweighted constructors share the weighted build path without collapsing repeated names into a single entry.
func unweighted(names []string) iter.Seq2[string, int] {
	return func(yield func(string, int) bool) {
		for _, name := range names {
			if !yield(name, 1) {
				return
			}
		}
	}
}

// pickWeighted selects an entry at random from a slice ordered by ascending cumulative weight, where cumulativeOf
// reports the running weight total through that entry (so the last entry's value is the grand total). It reports false
// when there is nothing to pick.
func pickWeighted[T any](entries []T, rnd xrand.Randomizer, cumulativeOf func(T) int64) (T, bool) {
	var zero T
	if len(entries) == 0 {
		return zero, false
	}
	total := cumulativeOf(entries[len(entries)-1])
	if total < 1 {
		return zero, false
	}
	v := 1 + randomBelow(rnd, total)
	// The entry to pick is the first whose cumulative total reaches v.
	if i, _ := slices.BinarySearchFunc(entries, v, func(e T, target int64) int {
		return cmp.Compare(cumulativeOf(e), target)
	}); i < len(entries) {
		return entries[i], true
	}
	return zero, false
}

// randomBelow returns a uniformly distributed value in [0, n) for a positive n. Randomizer.Intn works in the platform
// int, so on a 32-bit platform a total past math.MaxInt (two saturated weights suffice) is drawn in two parts.
func randomBelow(rnd xrand.Randomizer, n int64) int64 {
	return randomBelowWithLimit(rnd, n, math.MaxInt)
}

// randomBelowWithLimit is randomBelow with the largest argument Intn accepts made explicit, so the two-part path can be
// tested on any platform. When n exceeds limit the draw is hi*limit+lo, with hi in [0, ceil(n/limit)) and lo in
// [0, limit); rejecting values at or above n keeps the result uniform. The rejection loop is bounded so a degenerate
// Randomizer cannot hang it; after the last attempt the value is folded into range, which is biased but finite.
func randomBelowWithLimit(rnd xrand.Randomizer, n, limit int64) int64 {
	if n <= limit {
		return int64(rnd.Intn(int(n)))
	}
	buckets := (n + limit - 1) / limit
	const maxAttempts = 64
	var v int64
	for range maxAttempts {
		v = int64(rnd.Intn(int(buckets)))*limit + int64(rnd.Intn(int(limit)))
		if v < n {
			return v
		}
	}
	return v % n
}

// cumulativePairs applies cumulativeWeights to every entry of a transition table.
func cumulativePairs[K comparable, V cmp.Ordered, P any](source map[K]map[V]int, makePair func(item V, cumulative int64) P) map[K][]P {
	result := make(map[K][]P, len(source))
	for key, counts := range source {
		result[key] = cumulativeWeights(counts, makePair)
	}
	return result
}

// cumulativeWeights converts a map of per-item counts into the slice of cumulative weights pickWeighted consumes, built
// with makePair. Items are taken in sorted order so a seeded randomizer reproduces the same selections across runs.
func cumulativeWeights[V cmp.Ordered, P any](counts map[V]int, makePair func(item V, cumulative int64) P) []P {
	var total int64
	pairs := make([]P, 0, len(counts))
	for _, item := range slices.Sorted(maps.Keys(counts)) {
		total += int64(counts[item])
		pairs = append(pairs, makePair(item, total))
	}
	return pairs
}
