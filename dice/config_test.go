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
