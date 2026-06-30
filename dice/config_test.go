// Copyright (c) 2017-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package dice_test

import (
	"math"
	"testing"

	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestConfig(t *testing.T) {
	c := check.New(t)
	opts := dice.DefaultConfig()
	c.NoError(opts.Valid())
}

func TestConfigValidatesMaxModifier(t *testing.T) {
	c := check.New(t)

	// A negative MaxModifier is rejected: Normalize clamps to [-MaxModifier, MaxModifier], so a negative MaxModifier
	// would force every modifier to a fixed (negative) value, silently corrupting parsed dice. Confirm both Valid and
	// NewRoller refuse it.
	cfg := dice.DefaultConfig()
	cfg.MaxModifier = -1
	c.HasError(cfg.Valid())
	r, err := dice.NewRoller(cfg)
	c.HasError(err)
	c.True(r == nil)

	// Zero is permitted: a config that simply disallows modifiers. A parsed modifier is then clamped away to 0.
	cfg = dice.DefaultConfig()
	cfg.MaxModifier = 0
	c.NoError(cfg.Valid())
	r, err = dice.NewRoller(cfg)
	c.NoError(err)
	c.Equal(0, r.Parse("3d6+5").Modifier)
}

func TestConfigGuardsAverageIntermediateOverflow(t *testing.T) {
	c := check.New(t)

	// Regression: the overflow guard must bound Average's intermediate, which forms count*(sides+1) before halving --
	// one step larger than the count*sides product that Maximum and Roll use. With these limits 2*MaxSides == MaxInt-1
	// clears the count*sides check, but the count*(sides+1) that Average evaluates would wrap to a negative result.
	// Valid (and therefore NewRoller) must reject the config rather than let Average return garbage.
	overflowing := dice.DefaultConfig()
	overflowing.MaxCount = 2
	overflowing.MaxSides = math.MaxInt / 2 // 2*(MaxInt/2) == MaxInt-1; adding count back overflows
	overflowing.MaxModifier = 0
	overflowing.MaxMultiplier = 1
	c.HasError(overflowing.Valid())
	r, err := dice.NewRoller(overflowing)
	c.HasError(err)
	c.True(r == nil)

	// One step below that boundary is still valid, and Average must compute a sane, positive value for the most extreme
	// dice the config permits -- proving the new guard is not off by one and does not reject safe configs. The average
	// of two N-sided dice is N+1, here MaxSides+1 == math.MaxInt/2; an overflowing intermediate would instead go
	// negative.
	safe := dice.DefaultConfig()
	safe.MaxCount = 2
	safe.MaxSides = math.MaxInt/2 - 1 // 2*MaxSides + 2 == MaxInt, so count*(sides+1) stays in range
	safe.MaxModifier = 0
	safe.MaxMultiplier = 1
	c.NoError(safe.Valid())
	r, err = dice.NewRoller(safe)
	c.NoError(err)
	c.Equal(math.MaxInt/2, r.Average(dice.Dice{Count: 2, Sides: math.MaxInt/2 - 1, Multiplier: 1}))
}
