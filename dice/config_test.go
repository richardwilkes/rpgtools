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

func TestSetDefaultConfig(t *testing.T) {
	c := check.New(t)
	original := dice.DefaultConfig()
	defer func() { c.NoError(dice.SetDefaultConfig(original)) }()

	// An invalid Config is rejected and the default left untouched.
	cfg := dice.DefaultConfig()
	cfg.MaxCount = -5
	c.HasError(dice.SetDefaultConfig(cfg))
	c.Equal(original.MaxCount, dice.DefaultConfig().MaxCount)

	// A nil Config is likewise an error.
	c.HasError(dice.SetDefaultConfig(nil))
	c.Equal(original.MaxCount, dice.DefaultConfig().MaxCount)

	// A valid Config is observed both by DefaultConfig and by a Roller with no Config of its own.
	cfg = dice.DefaultConfig()
	cfg.MaxCount = 42
	cfg.GURPSFormat = !original.GURPSFormat
	c.NoError(dice.SetDefaultConfig(cfg))
	got := dice.DefaultConfig()
	c.Equal(42, got.MaxCount)
	c.Equal(cfg.GURPSFormat, got.GURPSFormat)
	var r dice.Roller
	c.Equal(42, r.Parse("100d6").Count)

	// The default holds a copy.
	cfg.MaxCount = 7
	c.Equal(42, dice.DefaultConfig().MaxCount)
}

func TestConfigValidatesMaxModifier(t *testing.T) {
	c := check.New(t)

	// A negative MaxModifier would force every modifier to a fixed value, so it is rejected.
	cfg := dice.DefaultConfig()
	cfg.MaxModifier = -1
	c.HasError(cfg.Valid())
	r, err := dice.NewRoller(cfg)
	c.HasError(err)
	c.True(r == nil)

	// Zero simply disallows modifiers.
	cfg = dice.DefaultConfig()
	cfg.MaxModifier = 0
	c.NoError(cfg.Valid())
	r, err = dice.NewRoller(cfg)
	c.NoError(err)
	c.Equal(0, r.Parse("3d6+5").Modifier)
}

func TestConfigRejectsFieldsAtMaxInt(t *testing.T) {
	c := check.New(t)

	// With every Max* field capped just below math.MaxInt, computeExtraDice's sides+1 intermediate cannot wrap and let
	// an overflowing config slip past Valid.
	cfg := dice.DefaultConfig()
	cfg.MaxCount = 1
	cfg.MaxSides = math.MaxInt
	cfg.MaxModifier = math.MaxInt
	cfg.MaxMultiplier = 1
	cfg.ExtraDiceFromModifiers = true
	c.HasError(cfg.Valid())
	r, err := dice.NewRoller(cfg)
	c.HasError(err)
	c.True(r == nil)

	// Each field at math.MaxInt is rejected on its own.
	for i, set := range []func(*dice.Config){
		func(o *dice.Config) { o.MaxCount = math.MaxInt },
		func(o *dice.Config) { o.MaxSides = math.MaxInt },
		func(o *dice.Config) { o.MaxModifier = math.MaxInt },
		func(o *dice.Config) { o.MaxMultiplier = math.MaxInt },
	} {
		cfg = dice.DefaultConfig()
		set(cfg)
		c.HasError(cfg.Valid(), "field %d at math.MaxInt", i)
	}

	// The cap itself (math.MaxInt-1) is still accepted.
	cfg = dice.DefaultConfig()
	cfg.MaxCount = 1
	cfg.MaxSides = math.MaxInt - 1
	cfg.MaxModifier = 0
	cfg.MaxMultiplier = 1
	cfg.ExtraDiceFromModifiers = false
	c.NoError(cfg.Valid())
}

// TestUnmarshalDiceStaysWithinFieldCap pins that UnmarshalText caps each field at maxFieldValue, never math.MaxInt.
func TestUnmarshalDiceStaysWithinFieldCap(t *testing.T) {
	c := check.New(t)
	var d dice.Dice
	c.NoError(d.UnmarshalText([]byte("99999999999999999999d99999999999999999999+99999999999999999999x99999999999999999999")))
	c.Equal(math.MaxInt-1, d.Count)
	c.Equal(math.MaxInt-1, d.Sides)
	c.Equal(math.MaxInt-1, d.Modifier)
	c.Equal(math.MaxInt-1, d.Multiplier)
}

func TestConfigGuardsAverageIntermediateOverflow(t *testing.T) {
	c := check.New(t)

	// The overflow guard must bound Average's count*(sides+1) intermediate, one step larger than the count*sides
	// product Maximum and Roll use.
	overflowing := dice.DefaultConfig()
	overflowing.MaxCount = 2
	overflowing.MaxSides = math.MaxInt / 2 // 2*(MaxInt/2) == MaxInt-1; adding count back overflows
	overflowing.MaxModifier = 0
	overflowing.MaxMultiplier = 1
	c.HasError(overflowing.Valid())
	r, err := dice.NewRoller(overflowing)
	c.HasError(err)
	c.True(r == nil)

	// One step below the boundary is valid, and Average must compute a sane value for the most extreme dice: two
	// N-sided dice average N+1.
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

func TestConfigCloneNilReceiver(t *testing.T) {
	c := check.New(t)
	var nilCfg *dice.Config
	c.NotPanics(func() { nilCfg = nilCfg.Clone() })
	c.True(nilCfg == nil)

	// A real clone is an independent copy.
	cfg := dice.DefaultConfig()
	clone := cfg.Clone()
	c.True(clone != cfg)
	c.Equal(*cfg, *clone)
	clone.MaxCount = cfg.MaxCount + 1
	c.NotEqual(cfg.MaxCount, clone.MaxCount)
}

// nilRandomizer exists only so a typed nil pointer can be stored in a Config's Randomizer field.
type nilRandomizer struct{}

func (*nilRandomizer) Intn(_ int) int { return 0 }

// TestConfigValidRejectsEachBound covers every check in Valid, each with a sibling one step inside the bound.
func TestConfigValidRejectsEachBound(t *testing.T) {
	c := check.New(t)
	var typedNil *nilRandomizer
	for _, one := range []struct {
		set   func(*dice.Config)
		name  string
		valid bool
	}{
		{func(o *dice.Config) { o.Randomizer = nil }, "nil Randomizer", false},
		{func(o *dice.Config) { o.Randomizer = typedNil }, "typed-nil Randomizer", false},
		{func(o *dice.Config) { o.MaxSides = 1 }, "MaxSides of 1", false},
		{func(o *dice.Config) { o.MaxSides = 0 }, "MaxSides of 0", false},
		{func(o *dice.Config) { o.MaxSides = 2 }, "MaxSides of 2", true},
		{func(o *dice.Config) { o.MaxMultiplier = 0 }, "MaxMultiplier of 0", false},
		{func(o *dice.Config) { o.MaxMultiplier = -1 }, "MaxMultiplier of -1", false},
		{func(o *dice.Config) { o.MaxMultiplier = 1 }, "MaxMultiplier of 1", true},
		// MaxCount*MaxSides itself overflows: 2*(MaxInt/2+1) is one past MaxInt.
		{func(o *dice.Config) {
			o.MaxCount = 2
			o.MaxSides = math.MaxInt/2 + 1
			o.MaxModifier = 0
			o.MaxMultiplier = 1
		}, "count*sides overflows", false},
		// Adding MaxModifier to the product overflows by exactly one...
		{func(o *dice.Config) {
			o.MaxCount = 1
			o.MaxSides = math.MaxInt / 2
			o.MaxModifier = math.MaxInt - math.MaxInt/2 + 1
			o.MaxMultiplier = 1
		}, "product+modifier overflows", false},
		// ...while one less lands exactly on MaxInt, which fits.
		{func(o *dice.Config) {
			o.MaxCount = 1
			o.MaxSides = math.MaxInt / 2
			o.MaxModifier = math.MaxInt - math.MaxInt/2
			o.MaxMultiplier = 1
		}, "product+modifier is exactly MaxInt", true},
		// The final multiply overflows: 2*(MaxInt/2+1) is one past MaxInt...
		{func(o *dice.Config) {
			o.MaxCount = 1
			o.MaxSides = 2
			o.MaxModifier = 0
			o.MaxMultiplier = math.MaxInt/2 + 1
		}, "sum*multiplier overflows", false},
		// ...while 2*(MaxInt/2) fits.
		{func(o *dice.Config) {
			o.MaxCount = 1
			o.MaxSides = 2
			o.MaxModifier = 0
			o.MaxMultiplier = math.MaxInt / 2
		}, "sum*multiplier fits", true},
	} {
		cfg := dice.DefaultConfig()
		one.set(cfg)
		err := cfg.Valid()
		if one.valid {
			c.NoError(err, one.name)
		} else {
			c.HasError(err, one.name)
		}
		r, err := dice.NewRoller(cfg)
		c.Equal(one.valid, err == nil, one.name)
		c.Equal(one.valid, r != nil, one.name)
	}
}

// TestConfigValidIgnoresExtraDiceFlagForOverflow pins that the overflow check measures what ApplyExtraDiceFromModifiers
// can actually produce, so setting the flag cannot make a safe config unsafe.
func TestConfigValidIgnoresExtraDiceFlagForOverflow(t *testing.T) {
	c := check.New(t)
	cfg := dice.DefaultConfig()
	cfg.MaxCount = 1
	cfg.MaxSides = 6
	cfg.MaxModifier = math.MaxInt - 10
	cfg.MaxMultiplier = 1
	cfg.ExtraDiceFromModifiers = false
	c.NoError(cfg.Valid())
	cfg.ExtraDiceFromModifiers = true
	c.NoError(cfg.Valid())

	// Nothing converts (the count is already at MaxCount), so the most extreme dice stay within the bounds Valid
	// measured.
	r, err := dice.NewRoller(cfg)
	c.NoError(err)
	d := dice.Dice{Count: 1, Sides: 6, Modifier: math.MaxInt - 10, Multiplier: 1}
	c.Equal(d, r.ApplyExtraDiceFromModifiers(d))
	c.Equal(math.MaxInt-9, r.Minimum(d))
	c.Equal(math.MaxInt-7, r.Average(d))
	c.Equal(math.MaxInt-4, r.Maximum(d))
	got := r.Roll(d)
	c.True(got >= math.MaxInt-9 && got <= math.MaxInt-4, "roll %d out of range", got)
}
