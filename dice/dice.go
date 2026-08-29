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
	// Only the one bool is needed, so read it under the lock rather than paying for a clone of the whole default Config
	// on every call, which adds up when a collection of Dice is encoded.
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
	buf := make([]byte, 0, 32) // Large enough that ordinary specs never need to grow it.
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
// surrounding whitespace, the whole of the text must be a dice specification, so malformed data ("total nonsense", or
// "3d6 extra") is reported as an error rather than silently decoded as some other value. Empty text is the zero Dice.
// The Dice is left unchanged when an error is returned.
func (dice *Dice) UnmarshalText(text []byte) error {
	d, rest := parseDice(string(text))
	if rest != "" {
		return errs.Newf("invalid dice specification: %q", string(text))
	}
	*dice = d
	return nil
}

// parseDice parses the dice specification that begins in (after trimming its surrounding whitespace) and also returns
// the unconsumed remainder, which is empty when the whole of the text was a specification. Parsing stops at the first
// character that cannot continue the specification; whatever has been parsed by then is the result.
//
// Every number is read against maxFieldValue, the largest value any Dice field may hold, and is only clamped into a
// particular field's range by the caller (Roller.Parse via Config.normalize) once its role is known. A leading number
// is a die count only if a die marker follows it, and is otherwise part of a bare modifier: capping it at MaxCount
// before that is known would silently shrink an ordinary modifier whenever MaxCount is below MaxModifier ("15" would
// parse as 10 with a MaxCount of 10). Reading both halves of a bare modifier ("50-60") in full likewise lets them be
// summed before the clamp, so the result is the clamped sum rather than a sum of separately clamped halves.
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
				// The sign's operand is really the count of a dice spec ("12+3d6", "+3d6") or, when nothing precedes
				// the sign, the spec begins right after it ("+d6"). As in ExtractDicePosition, neither the sign nor the
				// number before it is part of that spec, so parse the spec itself rather than dropping its dice. This
				// cannot recurse again: the re-parse begins at the die marker or the digits right before it, so it takes
				// the hadD path above and never reaches this branch.
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
		// A bare number is a modifier, so fold the lead into it. Both halves are in [0, maxFieldValue], so the
		// difference cannot overflow...
		dice.Modifier = lead - operand
	case operand > maxFieldValue-lead:
		// ...but the sum can, so saturate it at maxFieldValue rather than let it wrap to a sign-flipped value.
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
// If none can be found, -1, -1 will be returned. The span never contains an internal space, always begins with a digit
// or a die marker, and never ends with a dangling operator, so Roller.Parse consumes text[start:end] in full and yields
// exactly the specification the span represents. The span is not necessarily the canonical spelling of that
// specification, though: it may use shorthand ("3d" for 3d6) or be bare arithmetic ("5+3", which is the modifier 8), so
// formatting the parsed result may not reproduce the span.
func ExtractDicePosition(text string) (start, end int) {
	start = -1
	state := 0
	foundDigit := false   // The current candidate contains at least one digit (a count or a number of sides).
	hasD := false         // The current candidate contains a 'd'.
	droppedD := false     // A standalone (non-word) 'd' was discarded because no digit followed it.
	dInWord := false      // The 'd' starting the current candidate is adjacent to a letter, so it is part of a word.
	signHasDigit := false // A digit has followed the latest sign, so the sign has an operand and is not dangling.
	operandStart := -1    // The index of the first digit following the latest sign.
	numberEnd := -1       // The index just past a bare number whose trailing whitespace is being skipped (state 5).
	maximum := len(text)
	var prev rune
	for i, ch := range text {
		// A transition may discard the current candidate and ask for the character to be examined again in the state
		// it moves to, so that the character can begin a new candidate rather than being consumed by the discard.
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
					// A 'd' directly after a discarded 'd' (which is what a preceding die marker must be, since we are
					// back in this state) belongs to the same run of letters, so it is in a word if that one was, as
					// the second 'd' of "add 5" is.
					dInWord = isProseLetter(prev) || (isDieMarker(prev) && dInWord)
					state = 1
				case isSign(ch):
					signHasDigit = false
					state = 2
				case isMultiplier(ch):
					// A bare number with a multiplier ("5x2") is a spec in its own right, just as it is to Parse (which
					// makes 10 of it), so keep the number rather than discarding it and reporting only the multiplier's
					// operand as a misleading fragment of the spec. With no number before it ("x2") the multiplier is
					// not notation and its operand must not become a bare-number candidate either; state 3 consumes the
					// operand and, finding no candidate, rescans from whatever follows.
					state = 3
				case unicode.IsSpace(ch) && start != -1:
					// Whitespace after a bare number may just be trailing; defer judging it until we learn whether any
					// non-space content follows (handled by state 5). Any whitespace counts, matching the TrimSpace that
					// Parse applies, so a line ending or a tab does not hide a trailing number. Remember where the number
					// ends so the whitespace itself is never part of the span.
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
					// Discard the 'd': no digit followed it, so it is not a die marker. Only remember the discard when
					// the 'd' was standalone; a 'd' that is part of a word (adjacent to a prose letter before or after
					// it, as in "read 5" or "drum 5") must not suppress an unrelated bare number later in the text.
					// The current character is examined again so it can start a new candidate; consuming it here would
					// lose the second 'd' of "dd6".
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
			case 2: // Found a sign; take its digit operand, then a multiplier if present, as New does.
				switch {
				case isDigit(ch):
					if !signHasDigit {
						signHasDigit = true
						operandStart = i
					}
				case isMultiplier(ch) && signHasDigit:
					state = 3
				case isDieMarker(ch) && signHasDigit && !hasD:
					// The sign's operand turns out to be the count of a dice spec, as in "12+3d6" or "+3d6". Neither
					// the sign nor whatever preceded it (at most a bare number) belongs to that spec, so the candidate
					// restarts at the operand's first digit. The digits directly before this 'd' mean it cannot be part
					// of a word.
					start = operandStart
					foundDigit = true
					hasD = true
					dInWord = false
					state = 1
				case start == -1:
					// The sign had no candidate before it, so nothing has been collected: the sign and any operand were
					// merely prose (e.g. "+13 years"). Rescan from this character rather than ending the scan, which
					// would hide a spec later in the text.
					state = 0
					again = true
				default:
					// A sign with no digit operand is dangling: New reads an empty operand and drops the sign, so it
					// cannot carry a following multiplier into the spec. End the spec here so the trailing-operator
					// trim drops the dangling sign, keeping the span canonical (e.g. "d6+x2" yields "d6").
					state = 4
				}
			case 3: // Found an 'x'; allow digits. A space ends the spec, just as it does in New.
				if !isDigit(ch) {
					if start == -1 {
						// A signed bare number with a multiplier (e.g. "-5x2"), which is never reported; rescan from
						// this character.
						state = 0
						again = true
					} else {
						state = 4
					}
				}
			case 5:
				// A bare number was found and we are skipping the whitespace that follows it. It stays a valid result
				// only if the text ends here; any other character means the number was not the final token, so discard
				// it and rescan from this character.
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
		// The text ended while skipping the whitespace after a bare number, so the number is the spec.
		maximum = numberEnd
	}
	// A real specification must contain a digit, so a lone 'd' (as in "d" or "roll d", even at the end of the text) is
	// rejected. Additionally, once a standalone 'd' has been discarded as non-dice notation, only a candidate that
	// itself contains a 'd' is a real dice spec; without this the discarded 'd' would let an unrelated trailing bare
	// number (the "5" in "d 5" or "d-5") be reported, which is inconsistent with bare numbers like "13 years" returning
	// none. A 'd' embedded in a word (as in "read 5") is not treated as discarded, so it leaves a trailing bare number
	// reportable just like prose without any 'd'.
	if start != -1 && foundDigit && (hasD || !droppedD) {
		// Trim a trailing operator ('+', '-' or 'x'/'X') left without an operand, so the span covers only the dice spec
		// itself (e.g. "d6+" yields "d6" and "3d6x" yields "3d6"). Within the span such an operator is always dangling,
		// since an operand digit would otherwise follow it. Whitespace never reaches here: every state that can consume
		// it either ends the span before it or (state 5) excludes it via numberEnd.
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
