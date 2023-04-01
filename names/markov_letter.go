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

	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xmath/rand"
)

var _ Namer = &MarkovLetterNamer{}

type runeLast struct {
	ch   rune
	last int
}

// MarkovLetterNamer provides a name generator that creates a name based on markov chains of individual letter
// sequences.
type MarkovLetterNamer struct {
	depth   int
	mapping map[string][]runeLast
	final   map[rune]struct{}
	lengths [][2]int
}

// NewMarkovLetterNamer creates a new MarkovLetterNamer. The depth is the number of letters to consider within a run at
// a time. The data should be a map of names to a count which indicates how common the name is relative to others in the
// set. Any count less than 1 effectively removes the name from the set.
func NewMarkovLetterNamer(depth int, data map[string]int) *MarkovLetterNamer {
	if depth < 1 {
		depth = 1
	}
	n := &MarkovLetterNamer{
		depth: depth,
		final: make(map[rune]struct{}),
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

// NewMarkovLetterUnweightedNamer creates a new MarkovLetterNamer. The depth is the number of letters to consider within
// a run at a time. The data should be a set of names to train the model with.
func NewMarkovLetterUnweightedNamer(depth int, data []string) *MarkovLetterNamer {
	if depth < 1 {
		depth = 1
	}
	n := &MarkovLetterNamer{
		depth: depth,
		final: make(map[rune]struct{}),
	}
	mapping := make(map[string]map[rune]int)
	lengths := make(map[int]int)
	for _, name := range data {
		if name = strings.TrimSpace(name); name != "" {
			n.add(name, 1, mapping, lengths)
		}
	}
	n.finish(mapping, lengths)
	return n
}

func (n *MarkovLetterNamer) add(name string, count int, mapping map[string]map[rune]int, lengths map[int]int) {
	name = strings.ToLower(name)
	ch := make([]rune, n.depth)
	for _, next := range name {
		key := string(ch)
		m, ok := mapping[key]
		if !ok {
			m = make(map[rune]int)
			mapping[key] = m
		}
		m[next] += count
		for i := 0; i < n.depth-1; i++ {
			ch[i] = ch[i+1]
		}
		ch[n.depth-1] = next
	}
	n.final[ch[len(ch)-1]] = struct{}{}
	lengths[len(name)] += count
}

func (n *MarkovLetterNamer) finish(mapping map[string]map[rune]int, lengths map[int]int) {
	n.lengths = computeLengths(lengths)
	n.mapping = make(map[string][]runeLast)
	for k, v := range mapping {
		total := 0
		pairs := make([]runeLast, 0, len(v))
		for ch, count := range v {
			total += count
			pairs = append(pairs, runeLast{ch: ch, last: total})
		}
		n.mapping[k] = pairs
	}
}

// GenerateName generates a new random name.
func (n *MarkovLetterNamer) GenerateName() string {
	return n.GenerateNameWithRandomizer(rand.NewCryptoRand())
}

// GenerateNameWithRandomizer generates a new random name using the specified randomizer.
func (n *MarkovLetterNamer) GenerateNameWithRandomizer(rnd rand.Randomizer) string {
	var buffer strings.Builder
	max := selectMax(n.lengths, rnd)
	ch := make([]rune, n.depth)
	for {
		m, ok := n.mapping[string(ch)]
		if !ok {
			break
		}
		next := n.nextRune(m, rnd)
		if next == 0 {
			break
		}
		for i := 0; i < n.depth-1; i++ {
			ch[i] = ch[i+1]
		}
		ch[n.depth-1] = next
		buffer.WriteRune(next)
		if buffer.Len() >= max {
			if _, final := n.final[next]; final {
				break
			}
		}
	}
	return txt.FirstToUpper(buffer.String())
}

func computeLengths(lengths map[int]int) [][2]int {
	result := make([][2]int, 0, len(lengths))
	total := 0
	for length, count := range lengths {
		total += count
		result = append(result, [2]int{length, total})
	}
	return result
}

func selectMax(lengths [][2]int, rnd rand.Randomizer) int {
	max := rnd.Intn(lengths[len(lengths)-1][1])
	for _, p := range lengths {
		if p[1] >= max {
			return p[0]
		}
	}
	// Should not be reachable
	return 5
}

func (n *MarkovLetterNamer) nextRune(m []runeLast, rnd rand.Randomizer) rune {
	v := rnd.Intn(m[len(m)-1].last)
	for i := range m {
		if v <= m[i].last {
			return m[i].ch
		}
	}
	// Should not be reachable
	return 0
}
