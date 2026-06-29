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

	"github.com/richardwilkes/toolbox/v2/xunicode"
)

var (
	_ Namer                 = &MarkovRunNamer{}
	_ markovStepper[string] = runStepper{}
)

// MarkovRunNamer provides a name generator that creates a name based on markov chains of runs of vowels or consonants.
type MarkovRunNamer struct {
	*markov[string]
}

// runStepper generates one vowel/consonant run at a time, keyed on the previous run.
type runStepper struct{}

func (runStepper) initialKey() string {
	return ""
}

func (runStepper) steps(name string) []string {
	return decompose(name)
}

func (runStepper) advance(_, step string) string {
	return step
}

func (runStepper) length(step string) int {
	return utf8.RuneCountInString(step)
}

func (runStepper) write(b *strings.Builder, step string) {
	b.WriteString(step)
}

// decompose splits s into a sequence of maximal runs, each consisting entirely of vowels or entirely of consonants.
func decompose(s string) []string {
	var runs []string
	var buffer strings.Builder
	state := -1
	for _, ch := range s {
		isVowel := xunicode.IsVowely(ch)
		switch state {
		case 0:
			if isVowel {
				runs = append(runs, buffer.String())
				buffer.Reset()
				buffer.WriteRune(ch)
				state = 1
			} else {
				buffer.WriteRune(ch)
			}
		case 1:
			if isVowel {
				buffer.WriteRune(ch)
			} else {
				runs = append(runs, buffer.String())
				buffer.Reset()
				buffer.WriteRune(ch)
				state = 0
			}
		default:
			if isVowel {
				state = 1
			} else {
				state = 0
			}
			buffer.WriteRune(ch)
		}
	}
	if buffer.Len() != 0 {
		runs = append(runs, buffer.String())
	}
	return runs
}

// NewMarkovRunNamer creates a new MarkovRunNamer. The data should be a map of names to a count which indicates how
// common the name is relative to others in the set. Any count less than 1 effectively removes the name from the set. If
// 'lowered' is true, then the result will be forced to lowercase. If 'firstToUpper' is true, then the result will have
// its first letter capitalized.
func NewMarkovRunNamer(data map[string]int, lowered, firstToUpper bool) *MarkovRunNamer {
	return newMarkovRunNamer(maps.All(data), lowered, firstToUpper)
}

// NewMarkovRunUnweightedNamer creates a new MarkovRunNamer. The data should be a set of names to train the model with.
// If 'lowered' is true, then the result will be forced to lowercase. If 'firstToUpper' is true, then the result will
// have its first letter capitalized.
func NewMarkovRunUnweightedNamer(data []string, lowered, firstToUpper bool) *MarkovRunNamer {
	return newMarkovRunNamer(unweighted(data), lowered, firstToUpper)
}

func newMarkovRunNamer(data iter.Seq2[string, int], lowered, firstToUpper bool) *MarkovRunNamer {
	return &MarkovRunNamer{newMarkov[string](runStepper{}, data, lowered, firstToUpper)}
}
