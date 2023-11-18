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

func TestSimple(t *testing.T) {
	s := NewSimpleNamer(map[string]int{
		"a": 1,
		"b": 1,
	})
	counts := make(map[string]int)
	for i := 0; i < 100; i++ {
		counts[s.GenerateName()]++
	}
	check.Equal(t, 2, len(counts))
	_, exists := counts["A"]
	check.True(t, exists, "a")
	_, exists = counts["B"]
	check.True(t, exists, "b")
}
