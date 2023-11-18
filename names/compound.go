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

var _ Namer = &CompoundNamer{}

// CompoundNamer provides a name generator that combines multiple other name generators.
type CompoundNamer struct {
	namers       []Namer
	separator    string
	lowered      bool
	firstToUpper bool
}

// NewCompoundNamer creates a new CompoundNamer. The 'separator' will be placed between each name generated by the
// 'namers'. If 'lowered' is true, then the result will be forced to lowercase. If 'firstToUpper' is true, then the
// result will have its first letter capitalized.
func NewCompoundNamer(separator string, lowered, firstToUpper bool, namers ...Namer) *CompoundNamer {
	return &CompoundNamer{
		namers:       namers,
		separator:    separator,
		lowered:      lowered,
		firstToUpper: firstToUpper,
	}
}

// GenerateName generates a new random name.
func (n *CompoundNamer) GenerateName() string {
	return n.GenerateNameWithRandomizer(rand.NewCryptoRand())
}

// GenerateNameWithRandomizer generates a new random name using the specified randomizer.
func (n *CompoundNamer) GenerateNameWithRandomizer(rnd rand.Randomizer) string {
	var buffer strings.Builder
	for i, namer := range n.namers {
		if i != 0 {
			buffer.WriteString(n.separator)
		}
		buffer.WriteString(namer.GenerateNameWithRandomizer(rnd))
	}
	result := buffer.String()
	if n.lowered {
		result = strings.ToLower(result)
	}
	if n.firstToUpper {
		result = txt.FirstToUpper(result)
	}
	return result
}
