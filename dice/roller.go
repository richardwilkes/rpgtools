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
	return r.cfg.Clone()
}

// Format a Dice for display.
func (r *Roller) Format(dice Dice) string {
	return r.prepare(dice).format(r.cfg.GURPSFormat)
}

// Parse a dice string in the form 3d6+1x2 and turns it into a Dice.
func (r *Roller) Parse(spec string) Dice {
	return r.Normalize(parseDice(spec, r.cfg.MaxCount, r.cfg.MaxSides, r.cfg.MaxModifier, r.cfg.MaxMultiplier))
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
			if value > (math.MaxInt-digit)/10 {
				// value*10 + digit would overflow an int, so it necessarily exceeds maxValue (which is itself at most
				// math.MaxInt). Cap rather than wrap to a garbage value on an absurdly long number.
				value = maxValue
			} else {
				value = min(value*10+digit, maxValue) // Cap rather than overflow on an absurdly long number.
			}
		}
		inPos++
	}
	return value, inPos
}

// Roll the dice.
func (r *Roller) Roll(dice Dice) int {
	dice = r.prepare(dice)
	result := dice.Modifier
	switch {
	case dice.Sides > 1:
		for range dice.Count {
			result += 1 + r.cfg.Randomizer.Intn(dice.Sides)
		}
	case dice.Sides == 1:
		result += dice.Count
	}
	return result * dice.Multiplier
}

// Normalize the provided Dice, ensuring all values are within permitted ranges, and return the modified copy.
func (r *Roller) Normalize(dice Dice) Dice {
	dice.Count = min(max(dice.Count, 0), r.cfg.MaxCount)
	dice.Sides = min(max(dice.Sides, 0), r.cfg.MaxSides)
	dice.Modifier = min(max(dice.Modifier, -r.cfg.MaxModifier), r.cfg.MaxModifier)
	dice.Multiplier = min(max(dice.Multiplier, 1), r.cfg.MaxMultiplier)
	if dice.Count == 0 || dice.Sides == 0 {
		dice.Count = 0
		dice.Sides = 0
	}
	if dice.Count == 0 && dice.Modifier == 0 {
		dice.Multiplier = 1
	}
	return dice
}

// ApplyExtraDiceFromModifiers returns the Dice as if the ExtraDiceFromModifiers configuration option had been applied
// to its components. No more dice are added than the configured MaxCount allows: once the count would reach MaxCount,
// any modifier that would have converted into further dice is left in the modifier instead.
func (r *Roller) ApplyExtraDiceFromModifiers(dice Dice) Dice {
	dice = r.Normalize(dice)
	var adjustment int
	adjustment, dice.Modifier = computeExtraDice(dice.Sides, dice.Modifier, r.cfg.MaxCount-dice.Count)
	dice.Count += adjustment
	return dice
}

func (r *Roller) prepare(dice Dice) Dice {
	if r.cfg.ExtraDiceFromModifiers {
		return r.ApplyExtraDiceFromModifiers(dice)
	}
	return r.Normalize(dice)
}

// IsEquivalent returns true if the two Dice are equivalent.
func (r *Roller) IsEquivalent(d1, d2 Dice) bool {
	return r.Normalize(d1) == r.Normalize(d2)
}

// Minimum returns the minimum result.
func (r *Roller) Minimum(dice Dice) int {
	dice = r.prepare(dice)
	result := dice.Modifier
	if dice.Sides > 0 {
		result += dice.Count
	}
	return result * dice.Multiplier
}

// Average returns the average result.
func (r *Roller) Average(dice Dice) int {
	dice = r.prepare(dice)
	result := dice.Modifier
	if dice.Count > 0 && dice.Sides > 0 {
		result += dice.Count * (dice.Sides + 1) / 2
	}
	return result * dice.Multiplier
}

// Maximum returns the maximum result.
func (r *Roller) Maximum(dice Dice) int {
	dice = r.prepare(dice)
	result := dice.Modifier
	result += dice.Count * dice.Sides
	return result * dice.Multiplier
}

// PoolProbability return the probability that at least one die will be equal to or greater than the target value.
func (r *Roller) PoolProbability(dice Dice, target int) float64 {
	dice = r.Normalize(dice)
	if dice.Count < 1 || dice.Sides < 1 || dice.Sides < target {
		return 0
	}
	if target < 1 {
		// Every die rolls at least 1, so a non-positive target is always met.
		return 1
	}
	return 1 - math.Pow(1-float64(1+dice.Sides-target)/float64(dice.Sides), float64(dice.Count))
}
