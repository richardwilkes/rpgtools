// Copyright (c) 2017-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

// Package dice simulates dice using standard roleplaying game notation.
package dice

import (
	"math"
	"sync"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xrand"
	"github.com/richardwilkes/toolbox/v2/xreflect"
)

// maxFieldValue is the largest value permitted for any of a Config's Max* fields. It sits one below math.MaxInt so the
// overflow checks -- and computeExtraDice, which forms a sides+1 intermediate from MaxSides -- can add 1 to a field
// value without that addition itself wrapping. With the fields bounded this way, equationOverflows always sees the true
// values and reliably detects a real overflow instead of being handed one that has already wrapped negative.
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

// Config holds various configuration options Dice.
type Config struct {
	// Randomizer is the source of randomness to use when rolling dice. When Clone() is called, it is copied by
	// reference, so the same Randomizer is used in both the original and the clone.
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
	// be converted to extra dice for the purposes of display. For example, 1d6+8 will display as 3d6+1.
	ExtraDiceFromModifiers bool
}

// DefaultConfig returns a copy of the default Config that will be used if one isn't explicitly set on a Roller.
func DefaultConfig() *Config {
	defaultConfigLock.RLock()
	defer defaultConfigLock.RUnlock()
	return defaultConfig.Clone()
}

// SetDefaultConfig sets the default Config to use when one isn't explicitly set on a Roller. A copy will be made.
func SetDefaultConfig(opts *Config) {
	if opts.Valid() == nil {
		defaultConfigLock.Lock()
		defaultConfig = opts.Clone()
		defaultConfigLock.Unlock()
	}
}

// Clone this configuration. Currently, this is a simple copy, but provided so that if the options become more complex
// in the future, there is one canonical way to clone them.
func (c *Config) Clone() *Config {
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
// would overflow an int (while also accounting for the c.ExtraDiceFromModifiers flag). It assumes the ranges Valid()
// enforces before calling it: every Max* field is <= maxFieldValue, c.MaxCount >= 1, c.MaxSides is >= 2, and
// c.MaxMultiplier is >= 1, so c.MaxCount*c.MaxSides is non-negative and can only overflow past math.MaxInt, and the
// sides+1 intermediates computeExtraDice and Average form cannot themselves wrap. Adding c.MaxModifier can only
// overflow on the high side, since the product is non-negative and so the sum stays >= c.MaxModifier >= math.MinInt.
// The resulting sum may be negative, so the final multiply by c.MaxMultiplier is checked against both math.MaxInt and
// math.MinInt; because c.MaxMultiplier is >= 1, the bound on the wrong side of zero is never crossed. Each step is
// checked in Go's evaluation order, so an intermediate result overflowing is caught even when the final value would
// have been representable. Average forms a slightly larger intermediate than the equation above -- count*(sides+1)
// rather than count*sides -- so that product+count step is bounded as well, keeping every Roller computation safe and
// not just the one shown.
func (c *Config) equationOverflows() bool {
	var count, modifier int
	if c.ExtraDiceFromModifiers {
		// Measure the full, uncapped conversion here -- the worst case for overflow -- even though
		// ApplyExtraDiceFromModifiers caps the dice it actually adds at MaxCount at runtime.
		count, modifier = computeExtraDice(c.MaxSides, c.MaxModifier, math.MaxInt)
		count += c.MaxCount
		if count < 0 {
			return true
		}
	} else {
		count = c.MaxCount
		modifier = c.MaxModifier
	}
	if mulOverflows(count, c.MaxSides) {
		return true
	}
	product := count * c.MaxSides
	// Average evaluates count*(sides+1) -- that is, product + count -- before halving, so the product must leave room
	// to add count back without overflowing. count is >= 1 here (MaxCount is >= 1 and the ExtraDiceFromModifiers
	// adjustment only adds to it). Valid() caps every Max* field at maxFieldValue before this runs, so the sides+1 that
	// computeExtraDice forms above -- and count*(sides+1) here -- cannot themselves wrap.
	if product > math.MaxInt-count {
		return true
	}
	if modifier > 0 && product > math.MaxInt-modifier {
		return true
	}
	sum := product + modifier
	return sum > math.MaxInt/c.MaxMultiplier || sum < math.MinInt/c.MaxMultiplier
}

func mulOverflows(a, b int) bool {
	return a != 0 && b > math.MaxInt/a
}

// computeExtraDice converts as much of modifier as possible into extra dice of the given number of sides, returning the
// number of dice to add and the modifier left over after the conversion. No more than maxAdjustment dice are produced;
// any modifier that would have converted into further dice beyond that limit is left in the returned modifier instead.
// Pass a maxAdjustment of math.MaxInt (or any value at least as large as the uncapped count) to convert without a cap.
func computeExtraDice(sides, modifier, maxAdjustment int) (dieCountAdjustment, adjustedModifier int) {
	if sides < 2 || modifier < sides/2 || maxAdjustment < 1 {
		return 0, modifier
	}
	average := (sides + 1) / 2
	if sides&1 == 1 {
		// Odd number of sides, so average is a whole number and every die consumes exactly average of the modifier.
		dieCountAdjustment = min(modifier/average, maxAdjustment)
		return dieCountAdjustment, modifier - dieCountAdjustment*average
	}
	// Even number of sides, so each die's true average is average+0.5. A pair of dice therefore consumes exactly
	// 2*average+1 of the modifier and a lone die consumes average+1 (rounding the trailing half up).
	perPair := 2*average + 1
	dieCountAdjustment = 2 * (modifier / perPair)
	adjustedModifier = modifier % perPair
	if adjustedModifier >= average+1 {
		dieCountAdjustment++
		adjustedModifier -= average + 1
	}
	if dieCountAdjustment > maxAdjustment {
		// The full conversion overshoots the cap, so keep only maxAdjustment dice and leave the rest of the modifier
		// untouched. maxAdjustment/2 whole pairs each consume perPair; an odd cap adds one lone die consuming
		// average+1, charging exactly what the greedy conversion above would have charged those same dice.
		dieCountAdjustment = maxAdjustment
		consumed := (maxAdjustment / 2) * perPair
		if maxAdjustment&1 == 1 {
			consumed += average + 1
		}
		adjustedModifier = modifier - consumed
	}
	return dieCountAdjustment, adjustedModifier
}
