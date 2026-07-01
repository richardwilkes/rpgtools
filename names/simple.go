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

	"github.com/richardwilkes/toolbox/v2/xrand"
)

var _ Namer = &SimpleNamer{}

// SimpleNamer provides a name generator that selects a name from the weighted set of names provided to it.
type SimpleNamer struct {
	data         []weightedStep[string]
	lowered      bool
	firstToUpper bool
}

// NewSimpleNamer creates a new SimpleNamer. The data should be a map of names to a count which indicates how common the
// name is relative to others in the set. Any count less than 1 effectively removes the name from the set. If 'lowered'
// is true, then the result will be forced to lowercase. If 'firstToUpper' is true, then the result will have its first
// letter capitalized.
func NewSimpleNamer(data map[string]int, lowered, firstToUpper bool) *SimpleNamer {
	return newSimpleNamer(maps.All(data), lowered, firstToUpper)
}

// NewSimpleUnweightedNamer creates a new SimpleNamer. The data should be a set of names to choose from. If 'lowered' is
// true, then the result will be forced to lowercase. If 'firstToUpper' is true, then the result will have its first
// letter capitalized.
func NewSimpleUnweightedNamer(data []string, lowered, firstToUpper bool) *SimpleNamer {
	return newSimpleNamer(unweighted(data), lowered, firstToUpper)
}

func newSimpleNamer(data iter.Seq2[string, int], lowered, firstToUpper bool) *SimpleNamer {
	counts := make(map[string]int)
	for name, count := range data {
		if count > 0 {
			if name = strings.TrimSpace(name); name != "" {
				counts[name] = addWeight(counts[name], count)
			}
		}
	}
	return &SimpleNamer{
		data: cumulativeWeights(counts, func(name string, cumulative int64) weightedStep[string] {
			return weightedStep[string]{step: name, last: cumulative}
		}),
		lowered:      lowered,
		firstToUpper: firstToUpper,
	}
}

// GenerateName generates a new random name.
func (n *SimpleNamer) GenerateName() string {
	return generateName(n)
}

// GenerateNameWithRandomizer generates a new random name using the specified randomizer.
func (n *SimpleNamer) GenerateNameWithRandomizer(rnd xrand.Randomizer) string {
	if e, ok := pickWeighted(n.data, rnd, func(e weightedStep[string]) int64 { return e.last }); ok {
		return applyCase(e.step, n.lowered, n.firstToUpper)
	}
	return ""
}
