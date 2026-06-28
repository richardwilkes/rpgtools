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

	"github.com/richardwilkes/toolbox/v2/xrand"
)

var _ Namer = &MarkovLetterNamer{}

type runeLast struct {
	ch   rune
	last int
}

// MarkovLetterNamer provides a name generator that creates a name based on markov chains of individual letter
// sequences.
type MarkovLetterNamer struct {
	mapping      map[string][]runeLast
	final        map[rune]struct{}
	lengths      [][2]int
	depth        int
	maxLength    int
	lowered      bool
	firstToUpper bool
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
	n := &MarkovLetterNamer{
		depth:        depth,
		final:        make(map[rune]struct{}),
		lowered:      lowered,
		firstToUpper: firstToUpper,
	}
	mapping := make(map[string]map[rune]int)
	lengths := make(map[int]int)
	for name, count := range data {
		if count > 0 {
			if name = strings.TrimSpace(name); name != "" {
				n.add(name, count, mapping, lengths)
			}
		}
	}
	n.finish(mapping, lengths)
	return n
}

func (n *MarkovLetterNamer) add(name string, count int, mapping map[string]map[rune]int, lengths map[int]int) {
	ch := make([]rune, n.depth)
	for _, next := range name {
		key := string(ch)
		m, ok := mapping[key]
		if !ok {
			m = make(map[rune]int)
			mapping[key] = m
		}
		m[next] += count
		for i := range n.depth - 1 {
			ch[i] = ch[i+1]
		}
		ch[n.depth-1] = next
	}
	n.final[ch[len(ch)-1]] = struct{}{}
	lengths[utf8.RuneCountInString(name)] += count
}

func (n *MarkovLetterNamer) finish(mapping map[string]map[rune]int, lengths map[int]int) {
	n.lengths, n.maxLength = computeLengths(lengths)
	n.mapping = cumulativePairs(mapping, func(ch rune, cumulative int) runeLast {
		return runeLast{ch: ch, last: cumulative}
	})
}

// GenerateName generates a new random name.
func (n *MarkovLetterNamer) GenerateName() string {
	return n.GenerateNameWithRandomizer(xrand.New())
}

// GenerateNameWithRandomizer generates a new random name using the specified randomizer.
func (n *MarkovLetterNamer) GenerateNameWithRandomizer(rnd xrand.Randomizer) string {
	var buffer strings.Builder
	maximum := selectMax(n.lengths, rnd)
	// Past 'maximum' the loop keeps going only to end on a natural (final) letter. Training data whose transition graph
	// cycles without a reachable final letter would otherwise loop forever, so cap the length at twice the longest
	// training name as a safety valve; legitimate data never reaches it.
	hardCap := 2 * n.maxLength
	ch := make([]rune, n.depth)
	count := 0
	for {
		m, ok := n.mapping[string(ch)]
		if !ok {
			break
		}
		next := n.nextRune(m, rnd)
		if next == 0 {
			break
		}
		for i := range n.depth - 1 {
			ch[i] = ch[i+1]
		}
		ch[n.depth-1] = next
		buffer.WriteRune(next)
		count++
		if count >= maximum {
			if _, final := n.final[next]; final {
				break
			}
		}
		if count >= hardCap {
			break
		}
	}
	return applyCase(buffer.String(), n.lowered, n.firstToUpper)
}

func computeLengths(lengths map[int]int) (result [][2]int, maxLength int) {
	result = make([][2]int, 0, len(lengths))
	total := 0
	for length, count := range lengths {
		total += count
		result = append(result, [2]int{length, total})
		maxLength = max(maxLength, length)
	}
	return result, maxLength
}

func selectMax(lengths [][2]int, rnd xrand.Randomizer) int {
	if p, ok := pickWeighted(lengths, rnd, func(p [2]int) int { return p[1] }); ok {
		return p[0]
	}
	return 0
}

func (n *MarkovLetterNamer) nextRune(m []runeLast, rnd xrand.Randomizer) rune {
	if e, ok := pickWeighted(m, rnd, func(e runeLast) int { return e.last }); ok {
		return e.ch
	}
	return 0
}
