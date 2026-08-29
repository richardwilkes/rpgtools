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
	"maps"
	"strings"
	"unicode/utf8"
)

var (
	_ Namer               = &MarkovLetterNamer{}
	_ markovStepper[rune] = letterStepper{}
)

// MarkovLetterNamer provides a name generator that creates a name based on markov chains of individual letter
// sequences.
type MarkovLetterNamer struct {
	markov[rune]
}

// startMarker leads a lookup key while the sliding window is still filling, i.e. while fewer than depth runes have
// been taken; once the window is full the marker is dropped and the key is exactly the last depth runes. The marker is
// a byte that is not valid UTF-8, and the rest of every key is built solely from string(rune) conversions, which never
// yield an invalid byte, so no rune in a training name can alias the start state. (A NUL-rune sentinel used to, so a
// name containing U+0000 taught the chain a self-transition on the start key.) Being a constant, it also costs nothing
// to return, and a key never grows beyond the runes actually taken, so a huge depth neither allocates a huge key nor
// panics in the way a depth-sized make([]rune, depth) did.
const startMarker = "\xff"

// letterStepper generates one rune at a time, keyed on a sliding window of the previous 'depth' runes.
type letterStepper struct {
	depth int
}

func (s letterStepper) initialKey() string {
	return startMarker
}

func (s letterStepper) steps(name string) []rune {
	return []rune(name)
}

func (s letterStepper) advance(key string, step rune) string {
	if strings.HasPrefix(key, startMarker) {
		// The window is still filling: the key is the marker followed by every rune taken so far. Keep the marker until
		// this step brings the window up to depth runes.
		if utf8.RuneCountInString(key[len(startMarker):])+1 < s.depth {
			return key + string(step)
		}
		return key[len(startMarker):] + string(step)
	}
	// Slide the depth-rune window: drop its first rune and append the new one. The key holds exactly depth runes here,
	// so it is never empty.
	_, width := utf8.DecodeRuneInString(key)
	return key[width:] + string(step)
}

func (s letterStepper) length(rune) int {
	return 1
}

func (s letterStepper) write(b *strings.Builder, step rune) {
	b.WriteRune(step)
}

// NewMarkovLetterNamer creates a new MarkovLetterNamer. The depth is the number of letters to consider within a run at
// a time; a depth less than 1 is treated as 1, and a depth of at least the length of the longest training name makes
// the generator reproduce training names verbatim, so larger depths gain nothing. The data should be a map of names to
// a count which indicates how common the name is relative to others in the set. Any count less than 1 effectively
// removes the name from the set. If 'lowered' is true, then the result will be forced to lowercase. If 'firstToUpper'
// is true, then the result will have its first letter capitalized.
func NewMarkovLetterNamer(depth int, data map[string]int, lowered, firstToUpper bool) *MarkovLetterNamer {
	return newMarkovLetterNamer(depth, maps.All(data), lowered, firstToUpper)
}

// NewMarkovLetterUnweightedNamer creates a new MarkovLetterNamer. The depth is the number of letters to consider within
// a run at a time; a depth less than 1 is treated as 1, and a depth of at least the length of the longest training
// name makes the generator reproduce training names verbatim, so larger depths gain nothing. The data should be a list
// of names to train the model with; each occurrence counts once, so a name that appears more than once is weighted
// accordingly rather than collapsed to a single entry. If 'lowered' is true, then the result will be forced to
// lowercase. If 'firstToUpper' is true, then the result will have its first letter capitalized.
func NewMarkovLetterUnweightedNamer(depth int, data []string, lowered, firstToUpper bool) *MarkovLetterNamer {
	return newMarkovLetterNamer(depth, unweighted(data), lowered, firstToUpper)
}

func newMarkovLetterNamer(depth int, data iter.Seq2[string, int], lowered, firstToUpper bool) *MarkovLetterNamer {
	// A depth below 1 has no sensible meaning: the sliding window would start empty and grow with every step instead
	// of sliding, so clamp to 1, the smallest useful depth. Both exported constructors document this.
	if depth < 1 {
		depth = 1
	}
	return &MarkovLetterNamer{newMarkov[rune](letterStepper{depth: depth}, data, lowered, firstToUpper)}
}
