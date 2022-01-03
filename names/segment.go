// Copyright ©2017-2022 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package names

import "github.com/richardwilkes/toolbox/xmath/rand"

// Segment holds string segment and its frequency of occurrence.
type Segment struct {
	Value string `json:"value"`
	Freq  int    `json:"freq"`
}

// PickSegmentValue picks a value from the segment slice.
func PickSegmentValue(rnd rand.Randomizer, total int, segments []Segment) string {
	r := rnd.Intn(total)
	for i := range segments {
		if r < segments[i].Freq {
			return segments[i].Value
		}
		r -= segments[i].Freq
	}
	return segments[0].Value // Should never reach this line
}
