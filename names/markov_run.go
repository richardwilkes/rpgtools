// Copyright ©2017-2023 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package names

import (
	"strings"

	"github.com/richardwilkes/toolbox/v2/xrand"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/toolbox/v2/xunicode"
)

type stringLast struct {
	s    string
	last int
}

// MarkovRunNamer provides a name generator that creates a name based on markov chains of runs of vowels or consonants.
type MarkovRunNamer struct {
	mapping      map[string][]stringLast
	final        map[string]struct{}
	lengths      [][2]int
	lowered      bool
	firstToUpper bool
}

// NewMarkovRunNamer creates a new MarkovRunNamer. The data should be a map of names to a count which indicates how
// common the name is relative to others in the set. Any count less than 1 effectively removes the name from the set. If
// 'lowered' is true, then the result will be forced to lowercase. If 'firstToUpper' is true, then the result will have
// its first letter capitalized.
func NewMarkovRunNamer(data map[string]int, lowered, firstToUpper bool) *MarkovRunNamer {
	n := &MarkovRunNamer{
		final:        make(map[string]struct{}),
		lowered:      lowered,
		firstToUpper: firstToUpper,
	}
	mapping := make(map[string]map[string]int)
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

// NewMarkovRunUnweightedNamer creates a new MarkovRunNamer. The data should be a set of names to train the model with.
// If 'lowered' is true, then the result will be forced to lowercase. If 'firstToUpper' is true, then the result will
// have its first letter capitalized.
func NewMarkovRunUnweightedNamer(data []string, lowered, firstToUpper bool) *MarkovRunNamer {
	n := &MarkovRunNamer{
		final:        make(map[string]struct{}),
		lowered:      lowered,
		firstToUpper: firstToUpper,
	}
	mapping := make(map[string]map[string]int)
	lengths := make(map[int]int)
	for _, name := range data {
		if name = strings.TrimSpace(name); name != "" {
			n.add(name, 1, mapping, lengths)
		}
	}
	n.finish(mapping, lengths)
	return n
}

func (n *MarkovRunNamer) add(name string, count int, mapping map[string]map[string]int, lengths map[int]int) {
	last := ""
	for _, next := range n.decompose(name) {
		m, ok := mapping[last]
		if !ok {
			m = make(map[string]int)
			mapping[last] = m
		}
		m[next] += count
		last = next
	}
	n.final[last] = struct{}{}
	lengths[len(name)]++
}

func (n *MarkovRunNamer) decompose(s string) []string {
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

func (n *MarkovRunNamer) finish(mapping map[string]map[string]int, lengths map[int]int) {
	n.lengths = computeLengths(lengths)
	n.mapping = make(map[string][]stringLast)
	for k, v := range mapping {
		total := 0
		pairs := make([]stringLast, 0, len(v))
		for s, count := range v {
			total += count
			pairs = append(pairs, stringLast{s: s, last: total})
		}
		n.mapping[k] = pairs
	}
}

// GenerateName generates a new random name.
func (n *MarkovRunNamer) GenerateName() string {
	return n.GenerateNameWithRandomizer(xrand.New())
}

// GenerateNameWithRandomizer generates a new random name using the specified randomizer.
func (n *MarkovRunNamer) GenerateNameWithRandomizer(rnd xrand.Randomizer) string {
	var buffer strings.Builder
	maximum := selectMax(n.lengths, rnd)
	last := ""
	for {
		m, ok := n.mapping[last]
		if !ok {
			break
		}
		next := n.nextPart(m, rnd)
		if next == "" {
			break
		}
		last = next
		buffer.WriteString(next)
		if buffer.Len() >= maximum {
			if _, final := n.final[next]; final {
				break
			}
		}
	}
	result := buffer.String()
	if n.lowered {
		result = strings.ToLower(result)
	}
	if n.firstToUpper {
		result = xstrings.FirstToUpper(result)
	}
	return result
}

func (n *MarkovRunNamer) nextPart(m []stringLast, rnd xrand.Randomizer) string {
	v := rnd.Intn(m[len(m)-1].last)
	for i := range m {
		if v <= m[i].last {
			return m[i].s
		}
	}
	// Should not be reachable
	return ""
}
