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
	"encoding/binary"
	"hash"
	"math"
	"strconv"
	"strings"

	"github.com/richardwilkes/toolbox/xmath/rand"
)

// GURPSFormat determines whether GURPS dice formatting should be used. A value of true means the die count is always
// shown and the sides value is suppressed if it is a '6', while a value of false means the die count is suppressed if
// it is a '1' and the sides value is always shown.
var GURPSFormat = false

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
	spec = strings.TrimSpace(spec)
	var dice Dice
	var i int
	var ch byte
	dice.Count, i = extractValue(spec, 0)
	hadCount := i != 0
	ch, i = nextChar(spec, i)
	hadSides := false
	hadD := false
	if ch == 'd' || ch == 'D' {
		hadD = true
		j := i
		dice.Sides, i = extractValue(spec, i)
		hadSides = i != j
		ch, i = nextChar(spec, i)
	}
	if hadSides && !hadCount {
		dice.Count = 1
	} else if hadD && !hadSides && hadCount {
		dice.Sides = 6
	}
	if ch == '+' || ch == '-' {
		neg := ch == '-'
		dice.Modifier, i = extractValue(spec, i)
		if neg {
			dice.Modifier = -dice.Modifier
		}
		ch, i = nextChar(spec, i)
	}
	if !hadD {
		dice.Modifier += dice.Count
		dice.Count = 0
	}
	if ch == 'x' || ch == 'X' {
		dice.Multiplier, _ = extractValue(spec, i)
	}
	if dice.Multiplier == 0 {
		dice.Multiplier = 1
	}
	dice.Normalize()
	return &dice
}

func nextChar(in string, inPos int) (ch byte, outPos int) {
	if inPos < len(in) {
		return in[inPos], inPos + 1
	}
	return 0, inPos
}

func extractValue(in string, inPos int) (value, outPos int) {
	for inPos < len(in) {
		ch := in[inPos]
		if ch < '0' || ch > '9' {
			return value, inPos
		}
		value *= 10
		value += int(ch - '0')
		inPos++
	}
	return value, inPos
}

// ExtractDicePosition returns the start (inclusive) and end (exclusive) index of the Dice specification. If none can be found, -1, -1 will be returned.
func ExtractDicePosition(text string) (start, end int) {
	start = -1
	state := 0
	foundDigit := false
	maximum := len(text)
	for i, ch := range text {
		switch state {
		case 0: // Look for a leading number (with or without a sign) or a 'd'
			switch {
			case ch >= '0' && ch <= '9':
				foundDigit = true
				if start == -1 {
					start = i
				}
			case ch == 'd' || ch == 'D':
				if start == -1 {
					start = i
				}
				state = 1
			case ch == '+' || ch == '-':
				state = 2
			default:
				foundDigit = false
				start = -1
			}
		case 1: // Got 'd', but may not have found a digit yet; allow digits, sign or 'x'
			switch {
			case ch >= '0' && ch <= '9':
				foundDigit = true
			case !foundDigit:
				start = -1
				state = 0
			case ch == '+' || ch == '-':
				state = 2
			case ch == 'x' || ch == 'X':
				state = 3
			default:
				state = 4
			}
		case 2: // Found a sign; allow digits or 'x'
			if ch != ' ' && (ch < '0' || ch > '9') {
				if ch == 'x' || ch == 'X' {
					state = 3
				} else {
					state = 4
				}
			}
		case 3: // Found an 'x'; allow digits
			if ch != ' ' && (ch < '0' || ch > '9') {
				state = 4
			}
		}
		if state == 4 {
			maximum = i
			break
		}
	}
	if start != -1 {
		for start < maximum && text[start] == ' ' {
			start++
		}
		for maximum > start && text[maximum-1] == ' ' {
			maximum--
		}
		if start < maximum {
			return start, maximum
		}
	}
	return -1, -1
}

// Minimum returns the minimum result. 'extraDiceFromModifiers' determines if modifiers greater than or equal to the
// average result of the base die should be converted to extra dice for the purposes of this call. For example, 1d6+8
// will become 3d6+1.
func (dice *Dice) Minimum(extraDiceFromModifiers bool) int {
	count, result := dice.adjustedCountAndModifier(extraDiceFromModifiers)
	if dice.Sides > 0 {
		result += count
	}
	return result * dice.Multiplier
}

// Average returns the average result. 'extraDiceFromModifiers' determines if modifiers greater than or equal to the
// average result of the base die should be converted to extra dice for the purposes of this call. For example, 1d6+8
// will become 3d6+1.
func (dice *Dice) Average(extraDiceFromModifiers bool) int {
	count, result := dice.adjustedCountAndModifier(extraDiceFromModifiers)
	if count > 0 && dice.Sides > 0 {
		result += count * (dice.Sides + 1) / 2
	}
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
	switch {
	case dice.Sides > 1:
		for i := 0; i < count; i++ {
			result += 1 + rnd.Intn(dice.Sides)
		}
	case dice.Sides == 1:
		result = count
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
		if count != 0 && dice.Sides != 0 {
			buffer.WriteString("+")
		}
		buffer.WriteString(strconv.Itoa(modifier))
	} else if modifier < 0 {
		buffer.WriteString(strconv.Itoa(modifier))
	}
	if buffer.Len() == 0 {
		buffer.WriteString("0")
	}
	if dice.Multiplier != 1 {
		buffer.WriteString("x")
		buffer.WriteString(strconv.Itoa(dice.Multiplier))
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
	if dice.Sides == 0 {
		return dice.Count, dice.Modifier
	}
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
	if dice.Count < 0 {
		dice.Count = 0
	}
	if dice.Sides < 0 {
		dice.Sides = 0
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

// Hash writes this object's contents into the hasher.
//
//nolint:errcheck // Ignore failure to check error return on binary.Write
func (dice *Dice) Hash(h hash.Hash) {
	if dice == nil {
		return
	}
	_ = binary.Write(h, binary.LittleEndian, int64(dice.Count))
	_ = binary.Write(h, binary.LittleEndian, int64(dice.Sides))
	_ = binary.Write(h, binary.LittleEndian, int64(dice.Modifier))
	_ = binary.Write(h, binary.LittleEndian, int64(dice.Multiplier))
}
