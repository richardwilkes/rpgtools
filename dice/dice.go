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
	"bytes"
	"encoding/binary"
	"hash"
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/richardwilkes/toolbox/v2/xrand"
)

// GURPSFormat determines whether GURPS dice formatting should be used. A value of true means the die count is always
// shown and the sides value is suppressed if it is a '6', while a value of false means the die count is suppressed if
// it is a '1' and the sides value is always shown.
//
// If you modify this, you are responsible for ensuring it is done in a thread-safe context, as this code assumes it is
// effectively immutable when used.
var GURPSFormat = false

// MaxValue and MinValue bound each Dice field. The fields are kept within this range so that combinations of large
// values cannot saturate or overflow the derived calculations. MinValue only applies to Modifier, the sole field that
// may be negative; Count and Sides have a minimum of 0 and Multiplier a minimum of 1.
//
// These are variables so the limits may be tuned, but raise MaxValue with care: the calculations multiply up to three
// fields together (e.g. Maximum computes count*sides*multiplier), so a MaxValue much above ~2,000,000 risks overflowing
// the int results on a 64-bit platform.
//
// If you modify either of these variables, you are responsible for ensuring it is done in a thread-safe context, as
// this code assumes they are effectively immutable when used.
var (
	MaxValue = 2_000_000
	MinValue = -2_000_000
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
	spec = strings.TrimSpace(spec)
	var dice Dice
	var i int
	var ch byte
	dice.Count, i = extractValue(spec, 0)
	hadCount := i != 0
	ch, i = nextChar(spec, i)
	hadSides := false
	hadD := false
	if isDieMarker(rune(ch)) {
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
	if isSign(rune(ch)) {
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
	if isMultiplier(rune(ch)) {
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
		if !isDigit(rune(ch)) {
			return value, inPos
		}
		if value < MaxValue {
			value = value*10 + int(ch-'0')
			if value > MaxValue {
				value = MaxValue // Cap rather than overflow on an absurdly long number.
			}
		}
		inPos++
	}
	return value, inPos
}

// ExtractDicePosition returns the start (inclusive) and end (exclusive) index of the Dice specification. If none can be
// found, -1, -1 will be returned. The span never contains an internal space and always begins with a digit or a die
// marker, so re-parsing text[start:end] with New yields exactly the specification the span represents (New likewise
// stops at the first space and ignores any dangling operator).
func ExtractDicePosition(text string) (start, end int) {
	start = -1
	state := 0
	foundDigit := false   // The current candidate contains at least one digit (a count or a number of sides).
	hasD := false         // The current candidate contains a 'd'.
	droppedD := false     // A standalone (non-word) 'd' was discarded because no digit followed it.
	dInWord := false      // The 'd' starting the current candidate is adjacent to a prose letter, so it is part of a word.
	signHasDigit := false // A digit has followed the latest sign, so the sign has an operand and is not dangling.
	maximum := len(text)
	var prev rune
	for i, ch := range text {
		if state == 5 {
			// A bare number was found and we are skipping the spaces that follow it. It stays a valid result only if
			// the text ends here; any other character means the number was not the final token, so discard it and
			// rescan from this character.
			if ch == ' ' {
				prev = ch
				continue
			}
			start = -1
			foundDigit = false
			state = 0
		}
		switch state {
		case 0: // Look for a leading number (with or without a sign) or a 'd'
			switch {
			case isDigit(ch):
				foundDigit = true
				if start == -1 {
					start = i
				}
			case isDieMarker(ch):
				if start == -1 {
					start = i
				}
				hasD = true
				dInWord = isProseLetter(prev)
				state = 1
			case isSign(ch):
				signHasDigit = false
				state = 2
			case ch == ' ' && start != -1:
				// A space after a bare number may just be trailing whitespace; defer judging it until we learn whether
				// any non-space content follows (handled by the state == 5 branch above).
				state = 5
			default:
				foundDigit = false
				start = -1
				hasD = false
			}
		case 1: // Got 'd', but may not have found a digit yet; allow digits, sign or 'x'
			switch {
			case isDigit(ch):
				foundDigit = true
			case !foundDigit:
				// Discard the 'd': no digit followed it, so it is not a die marker. Only remember the discard when the
				// 'd' was standalone; a 'd' that is part of a word (adjacent to a prose letter before or after it, as
				// in "read 5" or "drum 5") must not suppress an unrelated bare number later in the text.
				if !dInWord && !isProseLetter(ch) {
					droppedD = true
				}
				start = -1
				hasD = false
				state = 0
			case isSign(ch):
				signHasDigit = false
				state = 2
			case isMultiplier(ch):
				state = 3
			default:
				state = 4
			}
		case 2: // Found a sign; take its digit operand, then a multiplier if present, as New does.
			switch {
			case isDigit(ch):
				signHasDigit = true
			case isMultiplier(ch) && signHasDigit:
				state = 3
			default:
				// A sign with no digit operand is dangling: New reads an empty operand and drops the sign, so it
				// cannot carry a following multiplier into the spec. End the spec here so the trailing-operator trim
				// drops the dangling sign, keeping the span canonical (e.g. "d6+x2" yields "d6").
				state = 4
			}
		case 3: // Found an 'x'; allow digits. A space ends the spec, just as it does in New.
			if !isDigit(ch) {
				state = 4
			}
		}
		if state == 4 {
			maximum = i
			break
		}
		prev = ch
	}
	// A real specification must contain a digit, so a lone 'd' (as in "d" or "roll d", even at the end of the text) is
	// rejected. Additionally, once a standalone 'd' has been discarded as non-dice notation, only a candidate that
	// itself contains a 'd' is a real dice spec; without this the discarded 'd' would let an unrelated trailing bare
	// number (the "5" in "d 5" or "d-5") be reported, which is inconsistent with bare numbers like "13 years" returning
	// none. A 'd' embedded in a word (as in "read 5") is not treated as discarded, so it leaves a trailing bare number
	// reportable just like prose without any 'd'.
	if start != -1 && foundDigit && (hasD || !droppedD) {
		// Trim a trailing operator ('+', '-' or 'x'/'X') left without an operand, plus any surrounding spaces, so the
		// span covers only the dice spec itself (e.g. "d6+" yields "d6" and "3d6x" yields "3d6"). Within the span such
		// an operator is always dangling, since an operand digit would otherwise follow it.
		for maximum > start {
			if c := text[maximum-1]; c != ' ' && c != '+' && c != '-' && c != 'x' && c != 'X' {
				break
			}
			maximum--
		}
		if start < maximum {
			return start, maximum
		}
	}
	return -1, -1
}

// isDigit reports whether ch is an ASCII decimal digit.
func isDigit(ch rune) bool { return ch >= '0' && ch <= '9' }

// isDieMarker reports whether ch is the 'd' that separates a die count from the number of sides.
func isDieMarker(ch rune) bool { return ch == 'd' || ch == 'D' }

// isMultiplier reports whether ch is the 'x' that introduces a result multiplier.
func isMultiplier(ch rune) bool { return ch == 'x' || ch == 'X' }

// isSign reports whether ch is a modifier sign.
func isSign(ch rune) bool { return ch == '+' || ch == '-' }

// isProseLetter reports whether r is an alphabetic letter that is not significant to dice notation (the 'd' die marker
// or the 'x' multiplier). A 'd' adjacent to such a letter belongs to an ordinary word rather than a dice specification,
// so ExtractDicePosition must not treat it as a discarded die marker.
func isProseLetter(r rune) bool {
	return unicode.IsLetter(r) && !isDieMarker(r) && !isMultiplier(r)
}

// Minimum returns the minimum result. 'extraDiceFromModifiers' determines if modifiers greater than or equal to the
// average result of the base die should be converted to extra dice for the purposes of this call. For example, 1d6+8
// will become 3d6+1.
func (dice *Dice) Minimum(extraDiceFromModifiers bool) int {
	count, sides, result, multiplier := dice.adjustedValues(extraDiceFromModifiers)
	if sides > 0 {
		result += count
	}
	return result * multiplier
}

// Average returns the average result. 'extraDiceFromModifiers' determines if modifiers greater than or equal to the
// average result of the base die should be converted to extra dice for the purposes of this call. For example, 1d6+8
// will become 3d6+1.
func (dice *Dice) Average(extraDiceFromModifiers bool) int {
	count, sides, result, multiplier := dice.adjustedValues(extraDiceFromModifiers)
	if count > 0 && sides > 0 {
		result += count * (sides + 1) / 2
	}
	return result * multiplier
}

// Maximum returns the maximum result. 'extraDiceFromModifiers' determines if modifiers greater than or equal to the
// average result of the base die should be converted to extra dice for the purposes of this call. For example, 1d6+8
// will become 3d6+1.
func (dice *Dice) Maximum(extraDiceFromModifiers bool) int {
	count, sides, result, multiplier := dice.adjustedValues(extraDiceFromModifiers)
	result += count * sides
	return result * multiplier
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
func (dice *Dice) RollWithRandomizer(rnd xrand.Randomizer, extraDiceFromModifiers bool) int {
	count, sides, result, multiplier := dice.adjustedValues(extraDiceFromModifiers)
	if rnd == nil {
		rnd = xrand.New()
	}
	switch {
	case sides > 1:
		for range count {
			result += 1 + rnd.Intn(sides)
		}
	case sides == 1:
		result += count
	}
	return result * multiplier
}

func (dice *Dice) String() string {
	return dice.StringExtra(false)
}

// StringExtra returns the text representation of the Dice. 'extraDiceFromModifiers' determines if modifiers greater
// than or equal to the average result of the base die should be converted to extra dice for the purposes of this call.
// For example, 1d6+8 will become 3d6+1.
func (dice *Dice) StringExtra(extraDiceFromModifiers bool) string {
	count, sides, modifier, multiplier := dice.adjustedValues(extraDiceFromModifiers)
	var buffer bytes.Buffer
	if count > 0 {
		if GURPSFormat || count > 1 {
			buffer.WriteString(strconv.Itoa(count))
		}
		buffer.WriteString("d")
		if !GURPSFormat || sides != 6 {
			buffer.WriteString(strconv.Itoa(sides))
		}
	}
	if modifier > 0 {
		if buffer.Len() != 0 {
			buffer.WriteString("+")
		}
		buffer.WriteString(strconv.Itoa(modifier))
	} else if modifier < 0 {
		buffer.WriteString(strconv.Itoa(modifier))
	}
	if buffer.Len() == 0 {
		buffer.WriteString("0")
	}
	if multiplier != 1 {
		buffer.WriteString("x")
		buffer.WriteString(strconv.Itoa(multiplier))
	}
	return buffer.String()
}

// ApplyExtraDiceFromModifiers alters the Dice, converting modifiers greater than or equal to the average result of the
// base die to extra dice. For example, 1d6+8 will become 3d6+1.
func (dice *Dice) ApplyExtraDiceFromModifiers() {
	dice.Count, dice.Sides, dice.Modifier, dice.Multiplier = dice.adjustedValues(true)
}

// adjustedValues returns the field values clamped to their permitted ranges, optionally converting modifiers to extra
// dice. It does not modify the receiver, so it is safe to call on Dice whose exported fields have been set directly to
// out-of-range values.
func (dice *Dice) adjustedValues(applyExtraDiceFromModifiers bool) (count, sides, modifier, multiplier int) {
	count, sides, modifier, multiplier = dice.clamped()
	if sides == 0 {
		return count, sides, modifier, multiplier
	}
	if applyExtraDiceFromModifiers && modifier > 0 {
		average := (sides + 1) / 2
		if sides&1 == 1 {
			// Odd number of sides, so average is a whole number
			count += modifier / average
			modifier %= average
		} else {
			// Even number of sides, so each die's true average is average+0.5. A pair of dice therefore consumes
			// exactly 2*average+1 of the modifier and a lone die consumes average+1 (rounding the trailing half up).
			// Compute the whole-pair and optional trailing-die counts directly; doing it by subtracting in a loop is
			// O(modifier), which is needlessly slow when the modifier is near its cap.
			perPair := 2*average + 1
			count += 2 * (modifier / perPair)
			modifier %= perPair
			if modifier >= average+1 {
				count++
				modifier -= average + 1
			}
		}
		count = min(count, MaxValue) // Converting a large modifier can push the count past its maximum.
	}
	return count, sides, modifier, multiplier
}

// clamped returns each field constrained to its permitted range without modifying the receiver. Because the fields are
// exported, callers may set them to any value, so every use must clamp rather than assume a valid range. A Dice that
// rolls no dice (a zero Count or a zero Sides) has no meaningful Count or Sides, so both are collapsed to 0. This gives
// every "no dice" spec one canonical form: it renders as just its modifier and compares (and hashes) equal to any other
// no-dice spec with the same modifier, so e.g. "0d6+2", "3d0+2" and "2" all behave identically rather than only some of
// them collapsing depending on which field happened to be zero.
func (dice *Dice) clamped() (count, sides, modifier, multiplier int) {
	count = min(max(dice.Count, 0), MaxValue)
	sides = min(max(dice.Sides, 0), MaxValue)
	modifier = min(max(dice.Modifier, MinValue), MaxValue)
	multiplier = min(max(dice.Multiplier, 1), MaxValue)
	if count == 0 || sides == 0 {
		count = 0
		sides = 0
	}
	return count, sides, modifier, multiplier
}

// Normalize the internal state, clamping each field to its permitted range.
func (dice *Dice) Normalize() {
	dice.Count, dice.Sides, dice.Modifier, dice.Multiplier = dice.clamped()
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

// IsEquivalent returns true if this Dice is equivalent to another Dice. Both are compared as if normalized, without
// modifying either Dice.
func (dice *Dice) IsEquivalent(other *Dice) bool {
	left := *dice
	right := *other
	left.Normalize()
	right.Normalize()
	return left == right
}

// PoolProbability return the probability that at least one die will be equal to or greater than the target value.
func (dice *Dice) PoolProbability(target int) float64 {
	count, sides, _, _ := dice.clamped()
	if count < 1 || sides < 1 || sides < target {
		return 0
	}
	if target < 1 {
		// Every die rolls at least 1, so a non-positive target is always met.
		return 1
	}
	return 1 - math.Pow(1-float64(1+sides-target)/float64(sides), float64(count))
}

// Hash writes this object's contents into the hasher.
//
//nolint:errcheck // Ignore failure to check error return on binary.Write
func (dice *Dice) Hash(h hash.Hash) {
	if dice == nil {
		return
	}
	count, sides, modifier, multiplier := dice.clamped()
	_ = binary.Write(h, binary.LittleEndian, int64(count))
	_ = binary.Write(h, binary.LittleEndian, int64(sides))
	_ = binary.Write(h, binary.LittleEndian, int64(modifier))
	_ = binary.Write(h, binary.LittleEndian, int64(multiplier))
}
