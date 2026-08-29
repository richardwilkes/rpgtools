// Copyright (c) 2017-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package dice

import (
	"math"
	"sync"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xrand"
	"github.com/richardwilkes/toolbox/v2/xreflect"
)

// maxFieldValue is the largest value permitted for any of a Config's Max* fields. It sits one below math.MaxInt so that
// the overflow checks and computeExtraDice's sides+1 intermediate can add 1 to a field without wrapping.
const maxFieldValue = math.MaxInt - 1

var (
	defaultConfigLock sync.RWMutex
	defaultConfig     = &Config{
		Randomizer:             xrand.New(),
		MaxCount:               999_999,
		MaxSides:               999_999,
		MaxModifier:            999_999,
		MaxMultiplier:          999_999,
		GURPSFormat:            false,
		ExtraDiceFromModifiers: false,
	}
)

// Config holds the options for parsing, formatting and rolling Dice.
type Config struct {
	// Randomizer is the source of randomness to use when rolling dice. Clone copies it by reference, and nothing in
	// this package synchronizes calls to it, so it must be safe for concurrent use if the Roller holding it is shared
	// between goroutines, and always if installed via SetDefaultConfig. The Randomizer returned by xrand.New is; a
	// math/rand-backed one is not.
	Randomizer    xrand.Randomizer
	MaxCount      int
	MaxSides      int
	MaxModifier   int
	MaxMultiplier int
	// GURPSFormat determines whether GURPS dice formatting should be used. A value of true means the die count is
	// always shown and the sides value is suppressed if it is a '6', while a value of false means the die count is
	// suppressed if it is a '1' and the sides value is always shown.
	GURPSFormat bool
	// ExtraDiceFromModifiers determines if modifiers greater than or equal to the average result of the base die should
	// be converted to extra dice. The conversion is applied by Roll, Minimum, Average, Maximum and PoolProbability as
	// well as Format, so it changes the distribution of results: with it set, 1d6+8 is treated as 3d6+1.
	ExtraDiceFromModifiers bool
}

// DefaultConfig returns a copy of the default Config that will be used if one isn't explicitly set on a Roller.
func DefaultConfig() *Config {
	defaultConfigLock.RLock()
	defer defaultConfigLock.RUnlock()
	return defaultConfig.Clone()
}

// SetDefaultConfig sets the default Config to use when one isn't explicitly set on a Roller. A copy will be made. If
// the Config is not Valid, an error is returned and the default is left unchanged. Every zero-value Roller shares the
// installed Randomizer, so it must be safe for concurrent use.
func SetDefaultConfig(cfg *Config) error {
	if err := cfg.Valid(); err != nil {
		return err
	}
	defaultConfigLock.Lock()
	defaultConfig = cfg.Clone()
	defaultConfigLock.Unlock()
	return nil
}

// Clone this configuration. A nil Config clones to nil.
func (c *Config) Clone() *Config {
	if c == nil {
		return nil
	}
	other := *c
	return &other
}

// Valid returns nil if the data is usable.
func (c *Config) Valid() error {
	if c == nil {
		return errs.New("may not be nil")
	}
	if xreflect.IsNil(c.Randomizer) {
		return errs.New("Randomizer may not be nil")
	}
	if c.MaxCount < 1 {
		return errs.New("MaxCount may not be less than 1")
	}
	if c.MaxCount > maxFieldValue {
		return errs.Newf("MaxCount may not be greater than %d", maxFieldValue)
	}
	if c.MaxSides < 2 {
		return errs.New("MaxSides may not be less than 2")
	}
	if c.MaxSides > maxFieldValue {
		return errs.Newf("MaxSides may not be greater than %d", maxFieldValue)
	}
	if c.MaxModifier < 0 {
		return errs.New("MaxModifier may not be less than 0")
	}
	if c.MaxModifier > maxFieldValue {
		return errs.Newf("MaxModifier may not be greater than %d", maxFieldValue)
	}
	if c.MaxMultiplier < 1 {
		return errs.New("MaxMultiplier may not be less than 1")
	}
	if c.MaxMultiplier > maxFieldValue {
		return errs.Newf("MaxMultiplier may not be greater than %d", maxFieldValue)
	}
	if c.equationOverflows() {
		return errs.New("max values may cause an overflow")
	}
	return nil
}

// equationOverflows reports whether evaluating
//
//	value = (c.MaxCount*c.MaxSides + c.MaxModifier) * c.MaxMultiplier
//
// or Average's larger count*(sides+1) intermediate would overflow an int. It assumes the ranges Valid enforces before
// calling it, so every term is non-negative. Each step is checked in evaluation order so an overflowing intermediate is
// caught even when the final value would fit. ExtraDiceFromModifiers needs no separate treatment: the conversion never
// adds dice past MaxCount and only reduces the modifier.
func (c *Config) equationOverflows() bool {
	if mulOverflows(c.MaxCount, c.MaxSides) {
		return true
	}
	product := c.MaxCount * c.MaxSides
	// Average evaluates count*(sides+1), which is product+count.
	if product > math.MaxInt-c.MaxCount {
		return true
	}
	if product > math.MaxInt-c.MaxModifier {
		return true
	}
	return product+c.MaxModifier > math.MaxInt/c.MaxMultiplier
}

func mulOverflows(a, b int) bool {
	return a != 0 && b > math.MaxInt/a
}

// computeExtraDice converts as much of modifier as possible into at most maxAdjustment extra dice of the given number
// of sides, returning the number of dice to add and the modifier left over.
func computeExtraDice(sides, modifier, maxAdjustment int) (dieCountAdjustment, adjustedModifier int) {
	if sides < 2 || modifier < sides/2 || maxAdjustment < 1 {
		return 0, modifier
	}
	average := (sides + 1) / 2
	if sides&1 == 1 {
		// Odd sides: average is a whole number, so every die consumes exactly average.
		dieCountAdjustment = min(modifier/average, maxAdjustment)
		return dieCountAdjustment, modifier - dieCountAdjustment*average
	}
	// Even sides: the true average is average+0.5, so a pair of dice consumes 2*average+1 and a lone die average+1.
	perPair := 2*average + 1
	dieCountAdjustment = 2 * (modifier / perPair)
	adjustedModifier = modifier % perPair
	if adjustedModifier >= average+1 {
		dieCountAdjustment++
		adjustedModifier -= average + 1
	}
	if dieCountAdjustment > maxAdjustment {
		// Keep only maxAdjustment dice, charging exactly what the greedy conversion above charged them.
		dieCountAdjustment = maxAdjustment
		consumed := (maxAdjustment / 2) * perPair
		if maxAdjustment&1 == 1 {
			consumed += average + 1
		}
		adjustedModifier = modifier - consumed
	}
	return dieCountAdjustment, adjustedModifier
}
