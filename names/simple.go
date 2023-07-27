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
	"golang.org/x/exp/slices"
)

var _ Namer = &SimpleNamer{}

type nameCount struct {
	name  string
	count int
}

// SimpleNamer provides a name generator that selects a name from the weighted set of names provided to it.
type SimpleNamer struct {
	data  []nameCount
	total int
}

// NewSimpleNamer creates a new SimpleNamer. The data should be a map of names to a count which indicates how common the
// name is relative to others in the set. Any count less than 1 effectively removes the name from the set.
func NewSimpleNamer(data map[string]int) *SimpleNamer {
	n := SimpleNamer{data: make([]nameCount, 0, len(data))}
	for name, count := range data {
		if count > 0 {
			if name = strings.TrimSpace(name); name != "" {
				n.data = append(n.data, nameCount{name: strings.ToLower(name), count: count})
				n.total += count
			}
		}
	}
	slices.SortFunc(n.data, func(a, b nameCount) int { return txt.NaturalCmp(a.name, b.name, false) })
	return &n
}

// NewSimpleUnweightedNamer creates a new SimpleNamer. The data should be a set of names to choose from.
func NewSimpleUnweightedNamer(data []string) *SimpleNamer {
	n := SimpleNamer{data: make([]nameCount, 0, len(data))}
	for _, name := range data {
		if name = strings.TrimSpace(name); name != "" {
			n.data = append(n.data, nameCount{name: strings.ToLower(name), count: 1})
			n.total++
		}
	}
	slices.SortFunc(n.data, func(a, b nameCount) int { return txt.NaturalCmp(a.name, b.name, false) })
	return &n
}

// GenerateName generates a new random name.
func (n *SimpleNamer) GenerateName() string {
	return n.GenerateNameWithRandomizer(rand.NewCryptoRand())
}

// GenerateNameWithRandomizer generates a new random name using the specified randomizer.
func (n *SimpleNamer) GenerateNameWithRandomizer(rnd rand.Randomizer) string {
	v := rnd.Intn(n.total)
	for i := range n.data {
		if v -= n.data[i].count; v < 1 {
			return txt.FirstToUpper(n.data[i].name)
		}
	}
	// Should not be reachable
	return txt.FirstToUpper(n.data[0].name)
}
