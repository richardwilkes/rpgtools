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
	"encoding/binary"
	"hash"
	"strconv"
	"strings"
	"unicode"

	"github.com/richardwilkes/toolbox/v2/errs"
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

// MarshalText implements the encoding.TextMarshaler interface. The Dice is formatted using the default Config's
// GURPSFormat setting.
func (dice Dice) MarshalText() (text []byte, err error) {
	// Read the one field needed rather than cloning the whole default Config.
	defaultConfigLock.RLock()
	gurpsFormat := defaultConfig.GURPSFormat
	defaultConfigLock.RUnlock()
	return dice.normalize().formatBytes(gurpsFormat), nil
}

func (dice Dice) format(gurpsFormat bool) string {
	return string(dice.formatBytes(gurpsFormat))
}

// formatBytes returns the textual form of the Dice, which must already be normalized.
func (dice Dice) formatBytes(gurpsFormat bool) []byte {
	buf := make([]byte, 0, 32)
	if dice.Count > 0 {
		if gurpsFormat || dice.Count > 1 {
			buf = strconv.AppendInt(buf, int64(dice.Count), 10)
		}
		buf = append(buf, 'd')
		if !gurpsFormat || dice.Sides != 6 {
			buf = strconv.AppendInt(buf, int64(dice.Sides), 10)
		}
	}
	if dice.Modifier != 0 {
		if dice.Modifier > 0 && len(buf) != 0 {
			buf = append(buf, '+')
		}
		buf = strconv.AppendInt(buf, int64(dice.Modifier), 10)
	}
	if len(buf) == 0 {
		buf = append(buf, '0')
	}
	if dice.Multiplier != 1 && (dice.Count > 0 || dice.Modifier != 0) {
		buf = append(buf, 'x')
		buf = strconv.AppendInt(buf, int64(dice.Multiplier), 10)
	}
	return buf
}

// UnmarshalText implements the encoding.TextUnmarshaler interface. Unlike Roller.Parse, it is strict: apart from
// surrounding whitespace, the whole of the text must be a dice specification. Empty text is the zero Dice. The Dice is
// left unchanged when an error is returned.
func (dice *Dice) UnmarshalText(text []byte) error {
	d, rest := parseDice(string(text))
	if rest != "" {
		return errs.Newf("invalid dice specification: %q", string(text))
	}
	*dice = d
	return nil
}

// parseDice parses the dice specification that begins in (after trimming surrounding whitespace) and returns the
// unconsumed remainder. Parsing stops at the first character that cannot continue the specification.
//
// Every number is read against maxFieldValue and clamped to a field's range only by the caller, once its role is known:
// a leading number is a count only if a die marker follows it, and both halves of a bare modifier ("50-60") must be
// summed before the clamp.
func parseDice(in string) (dice Dice, rest string) {
	in = strings.TrimSpace(in)
	lead, i := extractValue(in, 0, maxFieldValue)
	hadCount := i != 0
	consumed := i
	var ch byte
	ch, i = nextChar(in, i)
	hadSides := false
	hadD := isDieMarker(rune(ch))
	if hadD {
		dice.Count = lead
		j := i
		dice.Sides, i = extractValue(in, i, maxFieldValue)
		hadSides = i != j
		consumed = i
		ch, i = nextChar(in, i)
	}
	if hadSides && !hadCount {
		dice.Count = 1
	} else if hadD && !hadSides && hadCount {
		dice.Sides = 6
	}
	neg := false
	operand := 0
	if isSign(rune(ch)) {
		neg = ch == '-'
		operandStart := i
		operand, i = extractValue(in, i, maxFieldValue)
		if !hadD {
			if next, _ := nextChar(in, i); isDieMarker(rune(next)) && (i != operandStart || !hadCount) {
				// The sign's operand is the count of a dice spec ("12+3d6", "+3d6", "+d6"). As in ExtractDicePosition,
				// neither the sign nor what precedes it is part of the spec, so parse from there. The re-parse takes
				// the hadD path, so it cannot recurse again.
				return parseDice(in[operandStart:])
			}
		}
		consumed = i
		ch, i = nextChar(in, i)
	}
	switch {
	case hadD:
		dice.Modifier = operand
		if neg {
			dice.Modifier = -operand
		}
	case neg:
		// A bare number is a modifier, so fold the lead into it. The difference cannot overflow...
		dice.Modifier = lead - operand
	case operand > maxFieldValue-lead:
		// ...but the sum can, so saturate it.
		dice.Modifier = maxFieldValue
	default:
		dice.Modifier = lead + operand
	}
	if isMultiplier(rune(ch)) {
		dice.Multiplier, consumed = extractValue(in, i, maxFieldValue)
	}
	return dice.normalize(), in[consumed:]
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
// If none can be found, -1, -1 will be returned. Roller.Parse consumes text[start:end] in full and yields exactly the
// specification the span represents, though the span may be shorthand ("3d" for 3d6) or bare arithmetic ("5+3", the
// modifier 8), so formatting the parsed result may not reproduce it.
func ExtractDicePosition(text string) (start, end int) {
	start = -1
	state := 0
	foundDigit := false   // The current candidate contains a digit.
	hasD := false         // The current candidate contains a 'd'.
	droppedD := false     // A standalone (non-word) 'd' was discarded because no digit followed it.
	dInWord := false      // The 'd' starting the current candidate is adjacent to a letter, so it is part of a word.
	signHasDigit := false // A digit has followed the latest sign, so it is not dangling.
	operandStart := -1    // The index of the first digit following the latest sign.
	numberEnd := -1       // The index just past a bare number whose trailing whitespace is being skipped (state 5).
	maximum := len(text)
	var prev rune
	for i, ch := range text {
		// A transition that discards the current candidate re-examines the character so it can begin a new one.
		for again := true; again; {
			again = false
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
					// A 'd' right after a discarded 'd' shares its word status (the second 'd' of "add 5").
					dInWord = isProseLetter(prev) || (isDieMarker(prev) && dInWord)
					state = 1
				case isSign(ch):
					signHasDigit = false
					state = 2
				case isMultiplier(ch):
					// A bare number with a multiplier ("5x2") is a spec, as it is to Parse. With no number before it
					// ("x2"), state 3 consumes the operand and rescans from whatever follows.
					state = 3
				case unicode.IsSpace(ch) && start != -1:
					// Whitespace after a bare number may be trailing; state 5 decides once it sees what follows. Any
					// whitespace counts, matching the TrimSpace Parse applies.
					numberEnd = i
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
					// No digit followed the 'd', so discard it. Only a standalone 'd' is remembered as dropped; one in
					// a word ("read 5", "drum 5") must not suppress a later bare number. Re-examine the character so
					// the second 'd' of "dd6" can start a new candidate.
					if !dInWord && !isProseLetter(ch) {
						droppedD = true
					}
					start = -1
					hasD = false
					state = 0
					again = true
				case isSign(ch):
					signHasDigit = false
					state = 2
				case isMultiplier(ch):
					state = 3
				default:
					state = 4
				}
			case 2: // Found a sign; take its digit operand, then a multiplier if present.
				switch {
				case isDigit(ch):
					if !signHasDigit {
						signHasDigit = true
						operandStart = i
					}
				case isMultiplier(ch) && signHasDigit:
					state = 3
				case isDieMarker(ch) && signHasDigit && !hasD:
					// The sign's operand is the count of a dice spec ("12+3d6", "+3d6"), so the candidate restarts at
					// the operand's first digit; neither the sign nor what preceded it belongs to the spec.
					start = operandStart
					foundDigit = true
					hasD = true
					dInWord = false
					state = 1
				case start == -1:
					// The sign had no candidate before it ("+13 years"), so rescan from this character rather than
					// ending the scan.
					state = 0
					again = true
				default:
					// A sign with no operand is dangling ("d6+x2"). End the spec here so the trim below drops it.
					state = 4
				}
			case 3: // Found an 'x'; allow digits.
				if !isDigit(ch) {
					if start == -1 {
						// A signed bare number with a multiplier ("-5x2") is never reported; rescan from here.
						state = 0
						again = true
					} else {
						state = 4
					}
				}
			case 5:
				// Skipping whitespace after a bare number, which is the spec only if the text ends here.
				if !unicode.IsSpace(ch) {
					start = -1
					foundDigit = false
					state = 0
					again = true
				}
			}
		}
		if state == 4 {
			maximum = i
			break
		}
		prev = ch
	}
	if state == 5 {
		maximum = numberEnd
	}
	// A spec must contain a digit, so a lone 'd' is rejected. Once a standalone 'd' has been discarded, only a
	// candidate with its own 'd' counts, so "d 5" reports nothing, consistent with "13 years"; a 'd' inside a word
	// ("read 5") does not suppress the number.
	if start != -1 && foundDigit && (hasD || !droppedD) {
		// Trim a trailing operator left without an operand ("d6+" yields "d6"). Whitespace never reaches here.
		for maximum > start {
			if c := rune(text[maximum-1]); !isSign(c) && !isMultiplier(c) {
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

func isDigit(ch rune) bool { return ch >= '0' && ch <= '9' }

func isDieMarker(ch rune) bool { return ch == 'd' || ch == 'D' }

func isMultiplier(ch rune) bool { return ch == 'x' || ch == 'X' }

func isSign(ch rune) bool { return ch == '+' || ch == '-' }

// isProseLetter reports whether r is a letter with no meaning in dice notation, so a 'd' adjacent to it is part of a
// word.
func isProseLetter(r rune) bool {
	return unicode.IsLetter(r) && !isDieMarker(r) && !isMultiplier(r)
}
