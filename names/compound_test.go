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
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xrand"
)

// fixedNamer is a Namer that always returns the same text, making compound output deterministic.
type fixedNamer string

func (f fixedNamer) GenerateName() string                                 { return string(f) }
func (f fixedNamer) GenerateNameWithRandomizer(_ xrand.Randomizer) string { return string(f) }

//nolint:goconst // The tests are more readable without constants for duplicated string
func TestCompoundSkipsNilNamers(t *testing.T) {
	c := check.New(t)
	a := fixedNamer("A")
	b := fixedNamer("B")

	// A nil namer anywhere in the list must be dropped rather than dereferenced at generation time, and it must
	// not leave a spurious separator behind.
	for i, one := range []struct {
		expected string
		namers   []Namer
	}{
		{"A-B", []Namer{a, nil, b}}, // 0 - nil in the middle
		{"A-B", []Namer{nil, a, b}}, // 1 - nil at the start
		{"A-B", []Namer{a, b, nil}}, // 2 - nil at the end
		{"A", []Namer{nil, a, nil}}, // 3 - only one survivor
		{"", []Namer{nil, nil}},     // 4 - nothing survives
	} {
		var result string
		c.NotPanics(func() {
			result = NewCompoundNamer("-", false, false, one.namers...).GenerateName()
		}, "table index %d", i)
		c.Equal(one.expected, result, "table index %d", i)
	}
}
