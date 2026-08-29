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

	// An invalid Config is rejected with an error and the default is left untouched, rather than being silently
	// discarded.
	cfg := dice.DefaultConfig()
	cfg.MaxCount = -5
	c.HasError(dice.SetDefaultConfig(cfg))
	c.Equal(original.MaxCount, dice.DefaultConfig().MaxCount)

	// A nil Config is likewise an error, not a panic or a silent no-op.
	c.HasError(dice.SetDefaultConfig(nil))
	c.Equal(original.MaxCount, dice.DefaultConfig().MaxCount)

	// A valid Config is installed and observed both by DefaultConfig and by a Roller that has no Config of its own.
	cfg = dice.DefaultConfig()
	cfg.MaxCount = 42
	cfg.GURPSFormat = !original.GURPSFormat
	c.NoError(dice.SetDefaultConfig(cfg))
	got := dice.DefaultConfig()
	c.Equal(42, got.MaxCount)
	c.Equal(cfg.GURPSFormat, got.GURPSFormat)
	var r dice.Roller
	c.Equal(42, r.Parse("100d6").Count)

	// The default holds a copy: mutating the Config that was passed in afterward must not leak into the default.
	cfg.MaxCount = 7
	c.Equal(42, dice.DefaultConfig().MaxCount)
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

func TestConfigRejectsFieldsAtMaxInt(t *testing.T) {
	c := check.New(t)

	// Regression for the equationOverflows false negative: with ExtraDiceFromModifiers set and MaxSides at math.MaxInt,
	// computeExtraDice's internal sides+1 term wrapped to a negative value, yielding a negative die-count adjustment
	// that canceled MaxCount to exactly 0 and let the overflowing config slip past the count<0 guard. The same config
	// with ExtraDiceFromModifiers cleared was already rejected, so Valid() gave opposite answers based solely on that
	// flag. Every Max* field is now capped just below math.MaxInt, so the config is rejected before that arithmetic
	// runs.
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

	// Each individual field at math.MaxInt is rejected on its own, independent of the others.
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

	// The largest permitted value for a field (math.MaxInt-1, the cap itself) is still accepted with the other fields
	// minimal, proving the bound is inclusive and not an off-by-one that rejects safe configs.
	cfg = dice.DefaultConfig()
	cfg.MaxCount = 1
	cfg.MaxSides = math.MaxInt - 1
	cfg.MaxModifier = 0
	cfg.MaxMultiplier = 1
	cfg.ExtraDiceFromModifiers = false
	c.NoError(cfg.Valid())
}

// TestUnmarshalDiceStaysWithinFieldCap verifies UnmarshalText parses against maxFieldValue, so an unmarshaled Dice
// never holds a field at math.MaxInt (which would be outside the envelope Config enforces).
func TestUnmarshalDiceStaysWithinFieldCap(t *testing.T) {
	c := check.New(t)
	var d dice.Dice
	// A count of 20 nines far exceeds math.MaxInt; the parser must cap it at maxFieldValue (math.MaxInt-1), never
	// math.MaxInt and never a wrapped/garbage value.
	c.NoError(d.UnmarshalText([]byte("99999999999999999999d99999999999999999999+99999999999999999999x99999999999999999999")))
	c.Equal(math.MaxInt-1, d.Count)
	c.Equal(math.MaxInt-1, d.Sides)
	c.Equal(math.MaxInt-1, d.Modifier)
	c.Equal(math.MaxInt-1, d.Multiplier)
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

func TestConfigCloneNilReceiver(t *testing.T) {
	c := check.New(t)
	// Clone is nil-safe like its siblings Valid, Dice.Hash and Roller.Config, rather than panicking on a nil receiver.
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

// TestConfigValidRejectsEachBound covers every individual check in Valid, including the overflow guards, each with a
// sibling one step inside the bound to prove no guard is off by one.
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
// can actually produce. It never adds dice past MaxCount and only ever reduces the modifier, so setting
// ExtraDiceFromModifiers cannot make a safe config unsafe. Previously the check measured the uncapped conversion, which
// is never performed, and so rejected this config with the flag set while accepting it with the flag clear.
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

	// The most extreme dice the config permits are safe under every method: nothing converts (the count is already at
	// MaxCount), so the arithmetic stays within the bounds Valid measured.
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
