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
	"iter"
	"strings"

	"github.com/richardwilkes/toolbox/v2/xrand"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

// Namer defines the methods required of a name generator.
type Namer interface {
	// GenerateName generates a new random name.
	GenerateName() string
	// GenerateNameWithRandomizer generates a new random name using the specified randomizer.
	GenerateNameWithRandomizer(rnd xrand.Randomizer) string
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
func pickWeighted[T any](entries []T, rnd xrand.Randomizer, cumulativeOf func(T) int) (T, bool) {
	var zero T
	if len(entries) == 0 {
		return zero, false
	}
	total := cumulativeOf(entries[len(entries)-1])
	if total < 1 {
		return zero, false
	}
	v := 1 + rnd.Intn(total)
	for _, entry := range entries {
		if cumulativeOf(entry) >= v {
			return entry, true
		}
	}
	return zero, false
}

// cumulativePairs converts a transition table of per-item counts into one of per-item cumulative weights suitable for
// pickWeighted. makePair builds the stored pair from an item and its running cumulative total.
func cumulativePairs[K, V comparable, P any](source map[K]map[V]int, makePair func(item V, cumulative int) P) map[K][]P {
	result := make(map[K][]P, len(source))
	for key, counts := range source {
		total := 0
		pairs := make([]P, 0, len(counts))
		for item, count := range counts {
			total += count
			pairs = append(pairs, makePair(item, total))
		}
		result[key] = pairs
	}
	return result
}
