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
	"strconv"
	"strings"
	"unicode"
)

// Dice holds the basic dice information.
type Dice struct {
	Count      int
	Sides      int
	Modifier   int
	Multiplier int
}

func (dice Dice) normalize() Dice {
	if dice.Count < 1 || dice.Sides < 1 {
		dice.Count = 0
		dice.Sides = 0
	}
	if dice.Multiplier < 1 || (dice.Count == 0 && dice.Modifier == 0) {
		dice.Multiplier = 1
	}
	return dice
}

// MarshalText implements the encoding.TextMarshaler interface.
func (dice Dice) MarshalText() (text []byte, err error) {
	return []byte(dice.normalize().format(DefaultConfig().GURPSFormat)), nil
}

func (dice Dice) format(gurpsFormat bool) string {
	var buffer bytes.Buffer
	if dice.Count > 0 {
		if gurpsFormat || dice.Count > 1 {
			buffer.WriteString(strconv.Itoa(dice.Count))
		}
		buffer.WriteString("d")
		if !gurpsFormat || dice.Sides != 6 {
			buffer.WriteString(strconv.Itoa(dice.Sides))
		}
	}
	if dice.Modifier != 0 {
		if dice.Modifier > 0 {
			if buffer.Len() != 0 {
				buffer.WriteString("+")
			}
		}
		buffer.WriteString(strconv.Itoa(dice.Modifier))
	}
	if buffer.Len() == 0 {
		buffer.WriteString("0")
	}
	if dice.Multiplier != 1 && (dice.Count > 0 || dice.Modifier != 0) {
		buffer.WriteString("x")
		buffer.WriteString(strconv.Itoa(dice.Multiplier))
	}
	return buffer.String()
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (dice *Dice) UnmarshalText(text []byte) error {
	*dice = parseDice(string(text), maxFieldValue, maxFieldValue, maxFieldValue, maxFieldValue)
	return nil
}

func parseDice(in string, maxCount, maxSides, maxModifier, maxMultiplier int) Dice {
	in = strings.TrimSpace(in)
	var dice Dice
	var i int
	dice.Count, i = extractValue(in, 0, maxCount)
	hadCount := i != 0
	var ch byte
	ch, i = nextChar(in, i)
	hadSides := false
	hadD := false
	if isDieMarker(rune(ch)) {
		hadD = true
		j := i
		dice.Sides, i = extractValue(in, i, maxSides)
		hadSides = i != j
		ch, i = nextChar(in, i)
	}
	if hadSides && !hadCount {
		dice.Count = 1
	} else if hadD && !hadSides && hadCount {
		dice.Sides = 6
	}
	if isSign(rune(ch)) {
		neg := ch == '-'
		dice.Modifier, i = extractValue(in, i, maxModifier)
		if neg {
			dice.Modifier = -dice.Modifier
		}
		ch, i = nextChar(in, i)
	}
	if !hadD {
		dice.Modifier += dice.Count
		dice.Count = 0
	}
	if isMultiplier(rune(ch)) {
		dice.Multiplier, _ = extractValue(in, i, maxMultiplier)
	}
	return dice.normalize()
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

// ExtractDicePosition returns the start (inclusive) and end (exclusive) index of a Dice specification within the text.
// If none can be found, -1, -1 will be returned. The span never contains an internal space and always begins with a
// digit or a die marker, so parsing text[start:end] yields exactly the specification the span represents.
func ExtractDicePosition(text string) (start, end int) {
	start = -1
	state := 0
	foundDigit := false   // The current candidate contains at least one digit (a count or a number of sides).
	hasD := false         // The current candidate contains a 'd'.
	droppedD := false     // A standalone (non-word) 'd' was discarded because no digit followed it.
	dInWord := false      // The 'd' starting the current candidate is adjacent to a letter, so it is part of a word.
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
