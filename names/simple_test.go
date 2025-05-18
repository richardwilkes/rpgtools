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
	"testing"

	"github.com/richardwilkes/toolbox/check"
)

var data = map[string]int{
	"aA": 1,
	"bB": 1,
}

func TestSimple(t *testing.T) {
	s := NewSimpleNamer(data, false, false)
	counts := make(map[string]int)
	for range 25 {
		counts[s.GenerateName()]++
	}
	check.Equal(t, 2, len(counts))
	_, exists := counts["aA"]
	check.True(t, exists, "expecting to find 'aA' in: %v", counts)
	_, exists = counts["bB"]
	check.True(t, exists, "expecting to find 'bB' in: %v", counts)
}

func TestSimpleLowered(t *testing.T) {
	s := NewSimpleNamer(data, true, false)
	counts := make(map[string]int)
	for range 25 {
		counts[s.GenerateName()]++
	}
	check.Equal(t, 2, len(counts))
	_, exists := counts["aa"]
	check.True(t, exists, "expecting to find 'aa' in: %v", counts)
	_, exists = counts["bb"]
	check.True(t, exists, "expecting to find 'bb' in: %v", counts)
}

func TestSimpleFirstUpper(t *testing.T) {
	s := NewSimpleNamer(data, false, true)
	counts := make(map[string]int)
	for range 25 {
		counts[s.GenerateName()]++
	}
	check.Equal(t, 2, len(counts))
	_, exists := counts["AA"]
	check.True(t, exists, "expecting to find 'AA' in: %v", counts)
	_, exists = counts["BB"]
	check.True(t, exists, "expecting to find 'BB' in: %v", counts)
}

func TestSimpleLoweredAndFirstUpper(t *testing.T) {
	s := NewSimpleNamer(data, true, true)
	counts := make(map[string]int)
	for range 25 {
		counts[s.GenerateName()]++
	}
	check.Equal(t, 2, len(counts))
	_, exists := counts["Aa"]
	check.True(t, exists, "expecting to find 'Aa' in: %v", counts)
	_, exists = counts["Bb"]
	check.True(t, exists, "expecting to find 'Bb' in: %v", counts)
}
