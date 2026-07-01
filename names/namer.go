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

// maxWeight caps an individual name or transition weight. Bounding each weight at math.MaxInt32 keeps the int64
// cumulative totals (see cumulativeWeights and pickWeighted) safe from overflow no matter how large the supplied counts
// are or how many entries are summed: even four billion maximum-weight entries stay within int64.
const maxWeight = math.MaxInt32

// addWeight returns sum + delta saturated at maxWeight. delta is capped at maxWeight first (a negative delta passes
// through unchanged, since callers drop non-positive counts before they reach a weighted table), so even a
// pathologically large count can neither push an accumulated weight past the ceiling nor overflow a platform int.
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

// generateName produces a name from n using a fresh default randomizer. Every Namer implementation's GenerateName
// delegates here so the choice of default randomizer (xrand.New) lives in one place instead of being repeated in each.
func generateName(n Namer) string {
	return n.GenerateNameWithRandomizer(xrand.New())
}

// applyCase applies the optional case transformations shared by every namer: lower-casing the whole result and/or
// upper-casing its first letter.
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
// when there is nothing to pick. Centralizing the selection keeps the off-by-one prone arithmetic in a single place.
func pickWeighted[T any](entries []T, rnd xrand.Randomizer, cumulativeOf func(T) int64) (T, bool) {
	var zero T
	if len(entries) == 0 {
		return zero, false
	}
	total := cumulativeOf(entries[len(entries)-1])
	if total < 1 {
		return zero, false
	}
	// The cumulative total is an int64 so summing the weights cannot overflow; Randomizer.Intn works in int, but the
	// total never exceeds the entry count times the maxWeight ceiling, which stays within int on the 64-bit platforms
	// this package targets.
	v := 1 + int64(rnd.Intn(int(total)))
	// The cumulative weights are non-decreasing, so the entry to pick is the first one whose running total reaches v.
	// Binary search for it rather than scanning every entry: a SimpleNamer or transition key built from many entries
	// would otherwise pay an O(n) walk on every draw, the same linear-scan trap Date.Year once had.
	if i, _ := slices.BinarySearchFunc(entries, v, func(e T, target int64) int {
		return cmp.Compare(cumulativeOf(e), target)
	}); i < len(entries) {
		return entries[i], true
	}
	return zero, false
}

// cumulativePairs converts a transition table of per-item counts into one of per-item cumulative weights suitable for
// pickWeighted. makePair builds the stored pair from an item and its running cumulative total. Each table is built by
// cumulativeWeights, so the accumulation is defined in exactly one place.
func cumulativePairs[K comparable, V cmp.Ordered, P any](source map[K]map[V]int, makePair func(item V, cumulative int64) P) map[K][]P {
	result := make(map[K][]P, len(source))
	for key, counts := range source {
		result[key] = cumulativeWeights(counts, makePair)
	}
	return result
}

// cumulativeWeights converts a map of per-item counts into a slice pairing each item with the running cumulative weight
// total through it (so the last entry's total is the grand total), the form pickWeighted consumes. The total is
// accumulated as an int64 so that summing many weights (each capped at maxWeight by the callers) cannot overflow. Items
// are accumulated in sorted order so that a given seeded randomizer reproduces the same selections across process runs;
// iterating the map in Go's randomized order would otherwise vary how the cumulative weights line up with the draws,
// and thus the generated names, from one run to the next.
func cumulativeWeights[V cmp.Ordered, P any](counts map[V]int, makePair func(item V, cumulative int64) P) []P {
	var total int64
	pairs := make([]P, 0, len(counts))
	for _, item := range slices.Sorted(maps.Keys(counts)) {
		total += int64(counts[item])
		pairs = append(pairs, makePair(item, total))
	}
	return pairs
}
