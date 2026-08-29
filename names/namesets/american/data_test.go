// Copyright (c) 2017-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package american_test

import (
	"maps"
	"testing"

	"github.com/richardwilkes/rpgtools/names/namesets/american"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestNameSetsReturnIndependentEqualCopies(t *testing.T) {
	c := check.New(t)
	// Each call must return an independent map: repeated calls agree and a mutation never leaks into a later call.
	for name, fn := range map[string]func() map[string]int{
		"Female": american.Female,
		"Male":   american.Male,
		"Last":   american.Last,
	} {
		first := fn()
		c.True(len(first) > 0, "%s must be non-empty", name)
		for _, count := range first {
			c.True(count > 0, "%s counts must all be positive, got %d", name, count)
		}

		second := fn()
		c.True(maps.Equal(first, second), "%s: repeated calls must return equal content", name)

		// Bump an existing count and add a key that cannot occur in the data.
		var existing string
		for k := range first {
			existing = k
			break
		}
		const sentinel = "\x00 sentinel - not a real name \x00"
		originalCount := first[existing]
		first[existing] = originalCount + 1_000_000
		first[sentinel] = 7

		third := fn()
		_, leaked := third[sentinel]
		c.False(leaked, "%s: a key added to a returned map leaked into a later call", name)
		c.Equal(originalCount, third[existing], "%s: a count bumped in a returned map leaked into a later call", name)
	}
}
