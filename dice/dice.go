// Copyright ©2017-2022 by Richard A. Wilkes. All rights reserved.
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
	"bytes"
	"math"
	"regexp"
	"strconv"

	"github.com/richardwilkes/toolbox/xmath/rand"
)

const diceRegexStr = `(((\d+)?[dD](\d+))|((\d+)[dD](\d+)?)){1}([-+]\d+)?([xX](\d+))?`

var (
	// GURPSFormat determines whether GURPS dice formatting should be used. A value of true means the die count is
	// always shown and the sides value is suppressed if it is a '6', while a value of false means the die count is
	// suppressed if it is a '1' and the sides value is always shown.
	GURPSFormat   = false
	diceRegex     = regexp.MustCompile(diceRegexStr)
	diceRegexOnly = regexp.MustCompile(`^\s*` + diceRegexStr + `\s*$`)
)

// Dice holds the dice information.
type Dice struct {
	Count      int
	Sides      int
	Modifier   int
	Multiplier int
}

// Roll the dice. This is short-hand for New(spec).Roll(extraDiceFromModifiers).
func Roll(spec string, extraDiceFromModifiers bool) int {
	return New(spec).Roll(extraDiceFromModifiers)
}

// New creates a new Dice based on the given spec.
func New(spec string) *Dice {
	dice := &Dice{}
	match := diceRegexOnly.FindStringSubmatch(spec)
	if match != nil {
		if match[2] != "" {
			dice.Count = atoi(match[3])
			dice.Sides = atoi(match[4])
		} else {
			dice.Count = atoi(match[6])
			dice.Sides = atoi(match[7])
		}
		dice.Modifier = atoi(match[8])
		dice.Multiplier = atoi(match[10])
	}
	dice.Normalize()
	return dice
}

func atoi(text string) int {
	if value, err := strconv.Atoi(text); err == nil {
		return value
	}
	return 0
}

// Minimum returns the minimum result. 'extraDiceFromModifiers' determines if modifiers greater than or equal to the
// average result of the base die should be converted to extra dice for the purposes of this call. For example, 1d6+8
// will become 3d6+1.
func (dice *Dice) Minimum(extraDiceFromModifiers bool) int {
	count, result := dice.adjustedCountAndModifier(extraDiceFromModifiers)
	result += count
	return result * dice.Multiplier
}

// Average returns the average result. 'extraDiceFromModifiers' determines if modifiers greater than or equal to the
// average result of the base die should be converted to extra dice for the purposes of this call. For example, 1d6+8
// will become 3d6+1.
func (dice *Dice) Average(extraDiceFromModifiers bool) int {
	count, result := dice.adjustedCountAndModifier(extraDiceFromModifiers)
	result += count * (dice.Sides + 1) / 2
	return result * dice.Multiplier
}

// Maximum returns the maximum result. 'extraDiceFromModifiers' determines if modifiers greater than or equal to the
// average result of the base die should be converted to extra dice for the purposes of this call. For example, 1d6+8
// will become 3d6+1.
func (dice *Dice) Maximum(extraDiceFromModifiers bool) int {
	count, result := dice.adjustedCountAndModifier(extraDiceFromModifiers)
	result += count * dice.Sides
	return result * dice.Multiplier
}

// Roll returns the result of rolling the dice. 'extraDiceFromModifiers' determines if modifiers greater than or equal
// to the average result of the base die should be converted to extra dice for the purposes of this call. For example,
// 1d6+8 will become 3d6+1.
func (dice *Dice) Roll(extraDiceFromModifiers bool) int {
	return dice.RollWithRandomizer(nil, extraDiceFromModifiers)
}

// RollWithRandomizer returns the result of rolling the dice. If 'rnd' is nil, rand.NewCryptoRand() will be used.
// 'extraDiceFromModifiers' determines if modifiers greater than or equal to the average result of the base die should
// be converted to extra dice for the purposes of this call. For example, 1d6+8 will become 3d6+1.
func (dice *Dice) RollWithRandomizer(rnd rand.Randomizer, extraDiceFromModifiers bool) int {
	count, result := dice.adjustedCountAndModifier(extraDiceFromModifiers)
	if rnd == nil {
		rnd = rand.NewCryptoRand()
	}
	for i := 0; i < count; i++ {
		result += 1 + rnd.Intn(dice.Sides)
	}
	return result * dice.Multiplier
}

func (dice *Dice) String() string {
	return dice.StringExtra(false)
}

// StringExtra returns the text representation of the Dice. 'extraDiceFromModifiers' determines if modifiers greater
// than or equal to the average result of the base die should be converted to extra dice for the purposes of this call.
// For example, 1d6+8 will become 3d6+1.
func (dice *Dice) StringExtra(extraDiceFromModifiers bool) string {
	count, modifier := dice.adjustedCountAndModifier(extraDiceFromModifiers)
	var buffer bytes.Buffer
	if count > 0 {
		if GURPSFormat || count > 1 {
			buffer.WriteString(strconv.Itoa(count))
		}
		buffer.WriteString("d")
		if !GURPSFormat || dice.Sides != 6 {
			buffer.WriteString(strconv.Itoa(dice.Sides))
		}
	}
	if modifier > 0 {
		buffer.WriteString("+")
		buffer.WriteString(strconv.Itoa(modifier))
	} else if modifier < 0 {
		buffer.WriteString(strconv.Itoa(modifier))
	}
	if dice.Multiplier != 1 {
		buffer.WriteString("x")
		buffer.WriteString(strconv.Itoa(dice.Multiplier))
	}
	if buffer.Len() == 0 {
		buffer.WriteString("0")
	}
	return buffer.String()
}

// ApplyExtraDiceFromModifiers alters the Dice, converting modifiers greater than or equal to the average result of the
// base die to extra dice. For example, 1d6+8 will become 3d6+1.
func (dice *Dice) ApplyExtraDiceFromModifiers() {
	dice.Count, dice.Modifier = dice.adjustedCountAndModifier(true)
}

func (dice *Dice) adjustedCountAndModifier(applyExtractDiceFromModifiers bool) (count, modifier int) {
	dice.Normalize()
	count = dice.Count
	modifier = dice.Modifier
	if applyExtractDiceFromModifiers && modifier > 0 {
		average := (dice.Sides + 1) / 2
		if dice.Sides&1 == 1 {
			// Odd number of sides, so average is a whole number
			count += modifier / average
			modifier %= average
		} else {
			// Even number of sides, so average has an extra half, which means
			// we alternate
			for modifier > average {
				if modifier > 2*average {
					modifier -= 2*average + 1
					count += 2
				} else {
					modifier -= average + 1
					count++
				}
			}
		}
	}
	if count < 0 {
		count = 0
	}
	return
}

// Normalize the internal state.
func (dice *Dice) Normalize() {
	if dice.Count < 1 {
		dice.Count = 1
	}
	if dice.Sides < 1 {
		dice.Sides = 6
	}
	if dice.Multiplier < 1 {
		dice.Multiplier = 1
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (dice Dice) MarshalText() (text []byte, err error) {
	return []byte(dice.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (dice *Dice) UnmarshalText(text []byte) error {
	*dice = *New(string(text))
	return nil
}

// IsEquivalent returns true if this Dice is equivalent to another Dice. Normalizes both Dice.
func (dice *Dice) IsEquivalent(other *Dice) bool {
	dice.Normalize()
	other.Normalize()
	return dice.Count == other.Count && dice.Sides == other.Sides && dice.Modifier == other.Modifier && dice.Multiplier == other.Multiplier
}

// PoolProbability return the probability that at least one die will be equal to or greater than the target value.
func (dice *Dice) PoolProbability(target int) float64 {
	dice.Normalize()
	if dice.Count < 1 || dice.Sides < target {
		return 0
	}
	return 1 - math.Pow(1-float64(1+dice.Sides-target)/float64(dice.Sides), float64(dice.Count))
}

// ExtractFirstPosition returns a two-element slice of integers defining the location of the first Dice specification in
// the input text it finds. The match itself is at text[loc[0]:loc[1]]. A return value of nil indicates no match was
// found.
func ExtractFirstPosition(text string) []int {
	return diceRegex.FindStringIndex(text)
}

// ExtractAllPositions is similar to ExtractFirstPosition, except it returns a slice of up to max matches. If max is
// less than 1, then all matches will be returned. A return value of nil indicates no matches were found.
func ExtractAllPositions(text string, max int) [][]int {
	if max < 1 {
		max = -1
	}
	return diceRegex.FindAllStringIndex(text, max)
}
