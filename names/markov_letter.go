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
	*markov[rune]
}

// letterStepper generates one rune at a time, keyed on a sliding window of the previous 'depth' runes.
type letterStepper struct {
	depth int
}

func (s letterStepper) initialKey() string {
	return string(make([]rune, s.depth))
}

func (s letterStepper) steps(name string) []rune {
	return []rune(name)
}

func (s letterStepper) advance(key string, step rune) string {
	// Slide the depth-rune window: drop its first rune and append the new one. The key always holds exactly depth
	// runes, so it is never empty.
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
// a time. The data should be a map of names to a count which indicates how common the name is relative to others in the
// set. Any count less than 1 effectively removes the name from the set. If 'lowered' is true, then the result will be
// forced to lowercase. If 'firstToUpper' is true, then the result will have its first letter capitalized.
func NewMarkovLetterNamer(depth int, data map[string]int, lowered, firstToUpper bool) *MarkovLetterNamer {
	return newMarkovLetterNamer(depth, maps.All(data), lowered, firstToUpper)
}

// NewMarkovLetterUnweightedNamer creates a new MarkovLetterNamer. The depth is the number of letters to consider within
// a run at a time. The data should be a set of names to train the model with. If 'lowered' is true, then the result
// will be forced to lowercase. If 'firstToUpper' is true, then the result will have its first letter capitalized.
func NewMarkovLetterUnweightedNamer(depth int, data []string, lowered, firstToUpper bool) *MarkovLetterNamer {
	return newMarkovLetterNamer(depth, unweighted(data), lowered, firstToUpper)
}

func newMarkovLetterNamer(depth int, data iter.Seq2[string, int], lowered, firstToUpper bool) *MarkovLetterNamer {
	if depth < 1 {
		depth = 1
	}
	return &MarkovLetterNamer{newMarkov[rune](letterStepper{depth: depth}, data, lowered, firstToUpper)}
}
