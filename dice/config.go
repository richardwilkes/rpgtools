// Copyright ©2017-2022 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package dice

import "github.com/richardwilkes/toolbox/xmath/rand"

// Config holds a configuration for Dice.
type Config struct {
	// Randomizer holds the randomizer that will be used.
	Randomizer rand.Randomizer
	// GURPSFormat determines whether GURPS dice formatting should be used. A
	// value of true means the die count is always shown and the sides value
	// is suppressed if it is a '6', while a value of false means the die
	// count is suppressed if it is a '1' and the sides value is always shown.
	GURPSFormat bool
	// ExtraDiceFromModifiers determines if modifiers greater than or equal to
	// the average result of the base die should be converted to extra dice.
	// For example, 1d6+8 will become 3d6+1.
	ExtraDiceFromModifiers bool
}

// DefaultGURPSConfig creates a Dice config suited for GURPS.
func DefaultGURPSConfig() *Config {
	return &Config{
		Randomizer:  rand.NewCryptoRand(),
		GURPSFormat: true,
	}
}
