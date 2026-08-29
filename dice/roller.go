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
)

// Roller provides the ability to parse, roll, and manipulate dice.
type Roller struct {
	cfg *Config
}

// NewRoller creates a new Roller from the given Config.
func NewRoller(cfg *Config) (*Roller, error) {
	if err := cfg.Valid(); err != nil {
		return nil, err
	}
	return &Roller{cfg: cfg.Clone()}, nil
}

// Config returns a clone of this Roller's Config.
func (r *Roller) Config() *Config {
	if r != nil && r.cfg != nil {
		return r.cfg.Clone()
	}
	return DefaultConfig() // already a clone
}

// config returns the Config for a single operation. Each exported method fetches it once so the whole operation sees
// one consistent Config even if SetDefaultConfig runs concurrently.
func (r *Roller) config() *Config {
	if r != nil && r.cfg != nil {
		return r.cfg
	}
	return DefaultConfig()
}

// Format a Dice for display.
func (r *Roller) Format(dice Dice) string {
	cfg := r.config()
	return cfg.prepare(dice).format(cfg.GURPSFormat)
}

// Parse a dice string in the form 3d6+1x2 and turn it into a Dice, with each field clamped to the range this Roller's
// Config permits. Parsing is lenient: surrounding whitespace is ignored and the first character that cannot continue
// the specification ends it ("3d6 damage" is 3d6; text with no notation is the zero Dice). A bare number is a modifier
// ("5+3" is 8). A sign before a dice spec, or a bare number joined to one by a sign, is dropped ("-3d6" and "12+3d6"
// are both 3d6), matching the span ExtractDicePosition reports. Use Dice.UnmarshalText to reject malformed text
// instead.
func (r *Roller) Parse(spec string) Dice {
	d, _ := parseDice(spec)
	return r.config().normalize(d)
}

func nextChar(in string, inPos int) (ch byte, outPos int) {
	if inPos < len(in) {
		return in[inPos], inPos + 1
	}
	return 0, inPos
}

func extractValue(in string, inPos, maxValue int) (value, outPos int) {
	for inPos < len(in) {
		ch := in[inPos]
		if !isDigit(rune(ch)) {
			return value, inPos
		}
		if value < maxValue {
			digit := int(ch - '0')
			if value > (math.MaxInt-digit)/10 { // value*10+digit would overflow, so it exceeds maxValue
				value = maxValue
			} else {
				value = min(value*10+digit, maxValue)
			}
		}
		inPos++
	}
	return value, inPos
}

// Roll the dice.
func (r *Roller) Roll(dice Dice) int {
	cfg := r.config()
	dice = cfg.prepare(dice)
	result := dice.Modifier
	switch {
	case dice.Sides > 1:
		for range dice.Count {
			result += 1 + cfg.Randomizer.Intn(dice.Sides)
		}
	case dice.Sides == 1:
		result += dice.Count
	}
	return result * dice.Multiplier
}

// Normalize the provided Dice, ensuring all values are within permitted ranges, and return the modified copy.
func (r *Roller) Normalize(dice Dice) Dice {
	return r.config().normalize(dice)
}

// ApplyExtraDiceFromModifiers returns the Dice as if the ExtraDiceFromModifiers configuration option had been applied
// to its components. No dice are added past the configured MaxCount; the unconverted modifier is retained instead.
func (r *Roller) ApplyExtraDiceFromModifiers(dice Dice) Dice {
	return r.config().applyExtraDiceFromModifiers(dice)
}

// normalize clamps each field of the Dice into the range this Config permits.
func (c *Config) normalize(dice Dice) Dice {
	dice.Count = min(max(dice.Count, 0), c.MaxCount)
	dice.Sides = min(max(dice.Sides, 0), c.MaxSides)
	dice.Modifier = min(max(dice.Modifier, -c.MaxModifier), c.MaxModifier)
	dice.Multiplier = min(max(dice.Multiplier, 1), c.MaxMultiplier)
	return dice.normalize()
}

// applyExtraDiceFromModifiers implements Roller.ApplyExtraDiceFromModifiers for this Config.
func (c *Config) applyExtraDiceFromModifiers(dice Dice) Dice {
	dice = c.normalize(dice)
	var adjustment int
	adjustment, dice.Modifier = computeExtraDice(dice.Sides, dice.Modifier, c.MaxCount-dice.Count)
	dice.Count += adjustment
	return dice
}

// prepare normalizes the Dice and, if this Config asks for it, converts its modifier into extra dice.
func (c *Config) prepare(dice Dice) Dice {
	if c.ExtraDiceFromModifiers {
		return c.applyExtraDiceFromModifiers(dice)
	}
	return c.normalize(dice)
}

// IsEquivalent returns true if the two Dice are equivalent.
func (r *Roller) IsEquivalent(d1, d2 Dice) bool {
	cfg := r.config()
	return cfg.normalize(d1) == cfg.normalize(d2)
}

// Minimum returns the minimum result.
func (r *Roller) Minimum(dice Dice) int {
	dice = r.config().prepare(dice)
	result := dice.Modifier
	if dice.Sides > 0 {
		result += dice.Count
	}
	return result * dice.Multiplier
}

// Average returns the average result, rounded down to a whole number only after the multiplier has been applied, so
// 1d6x10 averages 35, not 30.
func (r *Roller) Average(dice Dice) int {
	dice = r.config().prepare(dice)
	result := dice.Modifier
	half := 0
	if dice.Count > 0 && dice.Sides > 0 {
		// count*(sides+1) is twice the average. Carry its odd half separately so the multiplier is applied before it is
		// rounded away. Config.Valid bounds every product here.
		twice := dice.Count * (dice.Sides + 1)
		result += twice / 2
		half = twice & 1
	}
	return result*dice.Multiplier + half*dice.Multiplier/2
}

// Maximum returns the maximum result.
func (r *Roller) Maximum(dice Dice) int {
	dice = r.config().prepare(dice)
	result := dice.Modifier
	result += dice.Count * dice.Sides
	return result * dice.Multiplier
}

// PoolProbability returns the probability that at least one die will be equal to or greater than the target value. The
// Dice are prepared as Roll prepares them, so with ExtraDiceFromModifiers set the pool is the one actually rolled.
func (r *Roller) PoolProbability(dice Dice, target int) float64 {
	dice = r.config().prepare(dice)
	if dice.Count < 1 || dice.Sides < 1 || dice.Sides < target {
		return 0
	}
	if target < 1 {
		return 1
	}
	return 1 - math.Pow(1-float64(1+dice.Sides-target)/float64(dice.Sides), float64(dice.Count))
}
