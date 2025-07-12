// Copyright ©2017-2023 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package names

import "github.com/richardwilkes/toolbox/v2/xrand"

// Namer defines the methods required of a name generator.
type Namer interface {
	// GenerateName generates a new random name.
	GenerateName() string
	// GenerateNameWithRandomizer generates a new random name using the specified randomizer.
	GenerateNameWithRandomizer(rnd xrand.Randomizer) string
}
