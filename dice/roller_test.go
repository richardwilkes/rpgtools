// Copyright (c) 2017-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package dice_test

import (
	"fmt"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xrand"
)

func newRoller(c check.Checker, rnd xrand.Randomizer, gurpsFormat, extraDiceFromModifiers bool) *dice.Roller {
	c.Helper()
	cfg := dice.DefaultConfig()
	if rnd == nil {
		rnd = xrand.New()
	}
	cfg.Randomizer = rnd
	cfg.GURPSFormat = gurpsFormat
	cfg.ExtraDiceFromModifiers = extraDiceFromModifiers
	r, err := dice.NewRoller(cfg)
	c.NoError(err)
	return r
}

//nolint:goconst // The tests are more readable without constants for duplicated string
func TestCreation(t *testing.T) {
	c := check.New(t)
	for i, one := range []struct {
		Text                   string
		Expected               string
		Count                  int
		Sides                  int
		Modifier               int
		Multiplier             int
		GURPS                  bool
		ExtraDiceFromModifiers bool
	}{
		{" 1d6+2x3 ", "d6+2x3", 1, 6, 2, 3, false, false}, // 0
		{"1d6", "d6", 1, 6, 0, 1, false, false},           // 1
		{"1d6", "1d", 1, 6, 0, 1, true, false},            // 2
		{"d", "0", 0, 0, 0, 1, false, false},              // 3
		{"d8", "d8", 1, 8, 0, 1, false, false},            // 4
		{"2d", "2d6", 2, 6, 0, 1, false, false},           // 5
		{"2d4x2", "2d4x2", 2, 4, 0, 2, false, false},      // 6
		{"3d5+1", "3d5+1", 3, 5, 1, 1, false, false},      // 7
		{"abcd", "0", 0, 0, 0, 1, false, false},           // 8
		{"1d6+2x3", "d6+2x3", 1, 6, 2, 3, false, false},   // 9
		{"3d8-13", "3d8-13", 3, 8, -13, 1, false, false},  // 10
		{"3d8+13", "3d8+13", 3, 8, 13, 1, false, false},   // 11
		{"3d8+13", "3d8+13", 3, 8, 13, 1, true, false},    // 12
		{"3d8+13", "5d8+4", 3, 8, 13, 1, true, true},      // 13
		{"3d8+13", "5d8+4", 3, 8, 13, 1, false, true},     // 14
		{"3d6+13", "6d6+2", 3, 6, 13, 1, false, true},     // 15
		{"3d6+13", "6d+2", 3, 6, 13, 1, true, true},       // 16
		{"6d+2", "6d6+2", 6, 6, 2, 1, false, false},       // 17
		{"1d6", "d6", 1, 6, 0, 1, false, true},            // 18
		{"1d6+3", "d6+3", 1, 6, 3, 1, false, true},        // 19
		{"1d6+4", "2d6", 1, 6, 4, 1, false, true},         // 20
		{"1d6+5", "2d6+1", 1, 6, 5, 1, false, true},       // 21
		{"1d6+8", "3d6+1", 1, 6, 8, 1, false, true},       // 22
		{"-1", "-1", 0, 0, -1, 1, false, false},           // 23
		{"+2", "2", 0, 0, 2, 1, false, false},             // 24
		{"x3", "0", 0, 0, 0, 1, false, false},             // 25
		{"4", "4", 0, 0, 4, 1, false, false},              // 26
	} {
		desc := fmt.Sprintf("Table index %d: %s", i, one.Text)
		r := newRoller(c, nil, one.GURPS, one.ExtraDiceFromModifiers)
		d := r.Parse(one.Text)
		c.Equal(one.Expected, r.Format(d), desc)
		c.Equal(one.Count, d.Count, desc)
		c.Equal(one.Sides, d.Sides, desc)
		c.Equal(one.Modifier, d.Modifier, desc)
		c.Equal(one.Multiplier, d.Multiplier, desc)
	}
}

func TestApplyExtraDiceFromModifiersAfter(t *testing.T) {
	c := check.New(t)
	r := newRoller(c, nil, false, false)
	for i, one := range []struct {
		Text     string
		Expected string
		Count    int
		Modifier int
	}{
		{"d6", "d6", 1, 0},      // 0
		{"d6+3", "d6+3", 1, 3},  // 1
		{"d6+4", "2d6", 2, 0},   // 2
		{"d6+5", "2d6+1", 2, 1}, // 3
		{"d6+8", "3d6+1", 3, 1}, // 4
	} {
		desc := fmt.Sprintf("Table index %d: %s", i, one.Text)
		d := r.ApplyExtraDiceFromModifiers(r.Parse(one.Text))
		c.Equal(one.Expected, r.Format(d), desc)
		c.Equal(one.Count, d.Count, desc)
		c.Equal(one.Modifier, d.Modifier, desc)
	}
}

//nolint:goconst // The tests are more readable without constants for duplicated string
func TestStringRoundTrip(t *testing.T) {
	c := check.New(t)
	r := newRoller(c, nil, false, false)
	for i, one := range []struct {
		Text     string
		Expected string
	}{
		{"3d0+2", "2"},       // 0 - degenerate "no dice" specs collapse to just the modifier
		{"d0+2", "2"},        // 1
		{"3d0-2", "-2"},      // 2
		{"d6+2", "d6+2"},     // 3
		{"3d6+13", "3d6+13"}, // 4
		{"2d4x2", "2d4x2"},   // 5
		{"3d8-13", "3d8-13"}, // 6
		{"4", "4"},           // 7
		{"-1", "-1"},         // 8
		{"x3", "0"},          // 9 - degnerate, no dice and no modifiers results in just 0
	} {
		desc := fmt.Sprintf("Table index %d: %s", i, one.Text)
		d := r.Parse(one.Text)
		s := r.Format(d)
		c.Equal(one.Expected, s, desc)
		c.True(r.IsEquivalent(d, r.Parse(s)), "%s: %q did not round-trip", desc, s)
	}
}

func TestNoDiceCanonicalizesToModifier(t *testing.T) {
	c := check.New(t)
	r := newRoller(c, nil, false, false)
	modifierTwo := dice.Dice{Count: 0, Sides: 0, Modifier: 2, Multiplier: 1}
	for _, d := range []dice.Dice{
		{Count: 0, Sides: 6, Modifier: 2, Multiplier: 1}, // zero count, non-zero sides
		{Count: 3, Sides: 0, Modifier: 2, Multiplier: 1}, // non-zero count, zero sides
		r.Parse("0d6+2"),
		r.Parse("3d0+2"),
	} {
		c.Equal("2", r.Format(d), "%+v", d)
		c.True(r.IsEquivalent(d, r.Parse("2")), "%+v did not round-trip to 2", d)
		c.True(r.IsEquivalent(d, modifierTwo), "%+v not equivalent to bare modifier", d)
		n := r.Normalize(d)
		c.Equal(0, n.Count, "%+v", d)
		c.Equal(0, n.Sides, "%+v", d)
		c.Equal(2, n.Modifier, "%+v", d)
	}

	// With no modifier either, a no-dice spec is the canonical empty spec "0".
	empty := dice.Dice{Count: 0, Sides: 20, Modifier: 0, Multiplier: 1}
	c.Equal("0", r.Format(empty))
	c.True(r.IsEquivalent(empty, r.Parse("0")))

	// A real spec (both Count and Sides non-zero) is left untouched.
	withDice := r.Parse("3d6+2")
	c.Equal("3d6+2", r.Format(withDice))
	c.Equal(3, withDice.Count)
	c.Equal(6, withDice.Sides)
}

func TestRollSingleSided(t *testing.T) {
	c := check.New(t)
	r := newRoller(c, nil, false, false)
	for i, one := range []struct {
		Text     string
		Expected int
	}{
		{"2d1", 2},      // 0
		{"2d1+3", 5},    // 1
		{"1d1+5", 6},    // 2
		{"3d1-1", 2},    // 3
		{"2d1+3x2", 10}, // 4
	} {
		desc := fmt.Sprintf("Table index %d: %s", i, one.Text)
		d := r.Parse(one.Text)
		c.Equal(one.Expected, r.Roll(d), desc)
		c.Equal(one.Expected, r.Minimum(d), desc)
		c.Equal(one.Expected, r.Maximum(d), desc)
	}
}

// topFaceRandomizer always reports the highest face (n-1), making a roll deterministic: every die contributes its
// maximum, so the total equals Maximum(). This both removes randomness from the assertion and proves the loop ran for
// each clamped die.
type topFaceRandomizer struct{}

func (topFaceRandomizer) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	return n - 1
}

func TestRollTerminatesOnHugeCount(t *testing.T) {
	c := check.New(t)
	r := newRoller(c, nil, false, false)
	rd := newRoller(c, topFaceRandomizer{}, false, false)
	// Regression: a roll iterates Count times, so Count must be clamped to dice.MaxValue before the loop runs;
	// otherwise an enormous count is an effective hang. Cover both the parsed spec (extractValue caps the number) and a
	// field set directly to math.MaxInt, which bypasses the parser's cap and relies solely on the clamp inside the
	// roll. Run each roll in a goroutine so a regression times out rather than hanging the whole suite.
	for i, d := range []dice.Dice{
		r.Parse("99999999999999999999d6"),
		{Count: math.MaxInt, Sides: 6, Multiplier: 1},
	} {
		desc := fmt.Sprintf("case %d", i)
		minimum := r.Minimum(d)
		maximum := r.Maximum(d)
		done := make(chan [2]int, 1)
		go func() {
			// Roll() (the public entry point at line 230) uses crypto rand, so only its range can be checked.
			// RollWithRandomizer pinned to the top face must equal Maximum(), proving every clamped die was rolled.
			done <- [2]int{r.Roll(d), rd.Roll(d)}
		}()
		select {
		case got := <-done:
			c.True(got[0] >= minimum && got[0] <= maximum, "%s: roll %d outside [%d,%d]", desc, got[0], minimum, maximum)
			c.Equal(maximum, got[1], desc)
		case <-time.After(5 * time.Second):
			t.Fatalf("%s: roll did not terminate: the unbounded-count hang has regressed", desc)
		}
	}
}

func TestIsEquivalent(t *testing.T) {
	c := check.New(t)
	r := newRoller(c, nil, false, false)

	// Differences that normalize away (a sub-1 multiplier becomes 1) are still equivalent.
	a := dice.Dice{Count: 1, Sides: 6, Modifier: 2, Multiplier: 1}
	b := dice.Dice{Count: 1, Sides: 6, Modifier: 2, Multiplier: 0}
	c.True(r.IsEquivalent(a, b))

	// Genuinely different dice are not equivalent.
	c.False(r.IsEquivalent(a, dice.Dice{Count: 2, Sides: 6, Modifier: 2, Multiplier: 1}))
}

func TestPoolProbability(t *testing.T) {
	c := check.New(t)
	r := newRoller(c, nil, false, false)
	d := dice.Dice{Count: 3, Sides: 6}

	// Regression: a non-positive target is met by every roll, so the probability must be exactly 1,
	// never greater than 1 as it was previously (e.g. 1.0046 for target 0).
	for _, target := range []int{0, -1, -100} {
		c.Equal(1.0, r.PoolProbability(d, target), "target %d", target)
	}

	// A target of 1 is met by every face.
	c.Equal(1.0, r.PoolProbability(d, 1))

	// A target beyond the number of sides is impossible.
	c.Equal(0.0, r.PoolProbability(d, 7))

	// No dice, or a zero-sided die (which cannot roll), yields 0 rather than a division by zero.
	c.Equal(0.0, r.PoolProbability(dice.Dice{Count: 0, Sides: 6}, 3))
	c.Equal(0.0, r.PoolProbability(dice.Dice{Count: 3, Sides: 0}, 3))
	c.Equal(0.0, r.PoolProbability(dice.Dice{Count: 3, Sides: 0}, 0))

	// A representative interior value: 3d6 rolling at least one 6 is 1-(5/6)^3 = 91/216.
	c.True(math.Abs(r.PoolProbability(d, 6)-91.0/216.0) < 1e-12, "3d6 >=6 probability = %v, want ~%v",
		r.PoolProbability(d, 6), 91.0/216.0)

	// Across the valid range the probability stays within [0,1] and strictly decreases as the target
	// rises.
	prev := 2.0
	for target := 1; target <= 6; target++ {
		p := r.PoolProbability(d, target)
		c.True(p >= 0 && p <= 1, "target %d produced out-of-range probability %v", target, p)
		c.True(p < prev, "probability did not decrease at target %d: %v >= %v", target, p, prev)
		prev = p
	}
}

func TestExtractValueOverflow(t *testing.T) {
	c := check.New(t)
	r := newRoller(c, nil, false, false)
	cfg := r.Config()
	const huge = "99999999999999999999" // 20 nines: far larger than the field cap

	// Each numeric field caps at dice.MaxValue rather than wrapping or exceeding the permitted range, and parsing still
	// continues past the oversized number.
	d := r.Parse(huge + "d6")
	c.Equal(cfg.MaxCount, d.Count)
	c.Equal(6, d.Sides)

	d = r.Parse("3d" + huge)
	c.Equal(3, d.Count)
	c.Equal(cfg.MaxSides, d.Sides)

	d = r.Parse("d6+" + huge)
	c.Equal(cfg.MaxModifier, d.Modifier)

	// A negative modifier caps at -MaxModifier rather than wrapping past the permitted range.
	d = r.Parse("d6-" + huge)
	c.Equal(-cfg.MaxModifier, d.Modifier)

	d = r.Parse("2d6x" + huge)
	c.Equal(cfg.MaxMultiplier, d.Multiplier)

	// Normal values are unaffected.
	d = r.Parse("3d6+2x4")
	c.Equal(3, d.Count)
	c.Equal(6, d.Sides)
	c.Equal(2, d.Modifier)
	c.Equal(4, d.Multiplier)
}

func TestUnmarshalTextCapsOversizedNumbers(t *testing.T) {
	c := check.New(t)
	const huge = "99999999999999999999" // 20 nines: larger than math.MaxInt
	const fieldCap = math.MaxInt - 1    // UnmarshalText parses against maxFieldValue, one below math.MaxInt
	// Regression: UnmarshalText caps each field at maxFieldValue. extractValue must reach that cap without overflowing
	// an int along the way; previously value*10 wrapped past math.MaxInt and the cap check never fired, storing a
	// garbage (and sometimes negative) value instead. Each field must end up exactly at the cap, never negative and
	// never at the bare math.MaxInt that would let a later sides+1 intermediate wrap.
	var d dice.Dice
	c.NoError(d.UnmarshalText([]byte(huge + "d6")))
	c.Equal(fieldCap, d.Count)
	c.Equal(6, d.Sides)

	c.NoError(d.UnmarshalText([]byte("3d" + huge)))
	c.Equal(3, d.Count)
	c.Equal(fieldCap, d.Sides)

	c.NoError(d.UnmarshalText([]byte("d6+" + huge)))
	c.Equal(fieldCap, d.Modifier)

	c.NoError(d.UnmarshalText([]byte("d6-" + huge)))
	c.Equal(-fieldCap, d.Modifier)

	c.NoError(d.UnmarshalText([]byte("2d6x" + huge)))
	c.Equal(fieldCap, d.Multiplier)
}

// bigEvenAdjust independently computes the even-sided modifier-to-extra-dice conversion using arbitrary-precision math,
// providing a reference that is immune to fixed-width arithmetic mistakes regardless of how large the inputs grow. It
// mirrors the package rule: an even-sided die's true average is average+0.5, k = floor(2*modifier/(2*average+1)) dice
// are extracted, and those dice consume ceil(k*(2*average+1)/2) of the modifier, leaving the remainder.
func bigEvenAdjust(count, sides, modifier int) (wantCount, wantModifier int) {
	average := (sides + 1) / 2
	perPair := big.NewInt(int64(2*average + 1))
	m := big.NewInt(int64(modifier))
	k := new(big.Int).Quo(new(big.Int).Lsh(m, 1), perPair) // equivalent to floor(2*modifier / perPair)
	cost := new(big.Int).Mul(k, perPair)
	if k.Bit(0) == 1 { // k odd: the half-die rounds up
		cost.Add(cost, big.NewInt(1))
	}
	cost.Rsh(cost, 1) // /2
	r := new(big.Int).Sub(m, cost)
	return count + int(k.Int64()), int(r.Int64())
}

func TestExtraDiceEvenSidedMatchesReference(t *testing.T) {
	c := check.New(t)
	r := newRoller(c, nil, false, false)
	cfg := r.Config()
	// Converting modifiers to extra dice for even-sided dice must match an independent reference exactly. The small
	// modifiers lock in the prior shipped behavior; the large ones (up to the field cap) exercise the path where the
	// previous O(modifier) loop would have hung.
	for _, sides := range []int{2, 4, 6, 8, 10, 12, 20, 100} {
		for mod := 0; mod <= 600; mod++ {
			d := dice.Dice{Count: 1, Sides: sides, Modifier: mod, Multiplier: 1}
			d = r.ApplyExtraDiceFromModifiers(d)
			wantCount, wantMod := bigEvenAdjust(1, sides, mod)
			c.Equal(wantCount, d.Count, "sides=%d mod=%d count", sides, mod)
			c.Equal(wantMod, d.Modifier, "sides=%d mod=%d modifier", sides, mod)
		}
	}
	// Exercise the largest in-range even-sided value; mask off the low bit rather than assuming dice.MaxValue is even.
	largestEvenSides := cfg.MaxSides &^ 1
	for _, sides := range []int{2, 4, 6, 8, 100, largestEvenSides} {
		for _, mod := range []int{99999, 500000, cfg.MaxSides / 3, cfg.MaxSides - 1, cfg.MaxSides} {
			d := dice.Dice{Count: 1, Sides: sides, Modifier: mod, Multiplier: 1}
			d = r.ApplyExtraDiceFromModifiers(d)
			wantCount, wantMod := bigEvenAdjust(1, sides, mod)
			c.Equal(wantCount, d.Count, "sides=%d mod=%d count", sides, mod)
			c.Equal(wantMod, d.Modifier, "sides=%d mod=%d modifier", sides, mod)
		}
	}
}

func TestExtraDiceEvenSidedTerminatesOnSaturatedModifier(t *testing.T) {
	c := check.New(t)
	r := newRoller(c, nil, false, false)
	cfg := r.Config()
	d := r.Parse("1d6+99999999999999999999")
	d = r.ApplyExtraDiceFromModifiers(d)
	c.Equal(cfg.MaxModifier, r.Parse("1d6+99999999999999999999").Modifier, "modifier should cap to MaxValue")
	c.Equal(1+2*(cfg.MaxModifier/7), d.Count)
	c.Equal(cfg.MaxModifier%7, d.Modifier)
}

func TestExtractFirstPosition(t *testing.T) {
	c := check.New(t)
	for i, one := range []struct {
		Text  string
		Start int
		End   int
	}{
		{"d6", 0, 2},                         // 0
		{"roll 3d6 for me", 5, 8},            // 1
		{"d not for me, roll 2d6+2", 19, 24}, // 2
		{"roll d6x2", 5, 9},                  // 3
		{"roll 3dx2", 5, 9},                  // 4
		{"Just text", -1, -1},                // 5
		{"and two years later...", -1, -1},   // 6
		{"and 13 years later...", -1, -1},    // 7
		{"and +13 years later...", -1, -1},   // 8
		{"and -13 years later...", -1, -1},   // 9
		// A 'd' with no digit after it is not dice notation; the trailing bare number that follows must not be
		// reported as the spec, matching how bare numbers (cases 7-9) are ignored.
		{"d 5", -1, -1},  // 10
		{"d-5", -1, -1},  // 11
		{"d+5", -1, -1},  // 12
		{"d x2", -1, -1}, // 13
		{"dx2", -1, -1},  // 14
		// A genuine dice spec appearing after a discarded 'd' is still found.
		{"d 5d6", 2, 5}, // 15
		{"d 2d6", 2, 5}, // 16
		{"ddd6", 2, 4},  // 17
		// Lone trailing bare numbers (no discarded 'd') remain valid specs.
		{"5", 0, 1},      // 18
		{"13", 0, 2},     // 19
		{"roll 5", 5, 6}, // 20
		// A 'd' that is part of an ordinary word is not a discarded die marker, so it must not suppress a trailing
		// bare number the way a standalone 'd' (cases 10-14) does. The number stays reportable, exactly as it is
		// after a 'd'-free word like "roll" (case 20) or "the".
		{"read 5", 5, 6}, // 21 - 'd' at the end of a word
		{"old 5", 4, 5},  // 22 - 'd' at the end of a word
		{"add 5", 4, 5},  // 23 - 'd' at the end of a word
		{"hold 5", 5, 6}, // 24 - 'd' at the end of a word
		{"drum 5", 5, 6}, // 25 - 'd' at the start of a word (followed by a prose letter)
		{"the 5", 4, 5},  // 26 - control: a word without any 'd' already worked
		// A spec that ends with an operator but no operand: the dangling '+'/'-'/'x' (and any trailing spaces) must be
		// excluded from the returned span so a consumer slicing text[start:end] gets just the dice spec.
		{"d6+", 0, 2},    // 27
		{"d6-", 0, 2},    // 28
		{"3d6x", 0, 3},   // 29
		{"d6+x", 0, 2},   // 30 - multiple dangling operators
		{"2d6+2x", 0, 5}, // 31 - trailing 'x' trimmed, modifier retained
		{"d6+ ", 0, 2},   // 32 - operator followed by a trailing space
		{"3d", 0, 2},     // 33 - control: a trailing 'd' (meaning d6) is a valid operand, not trimmed
		// A lone 'd' with no digit is not a dice spec, even when it runs to the end of the text (where there is no
		// following character to trigger the discard mid-scan).
		{"d", -1, -1},      // 34
		{"roll d", -1, -1}, // 35
		// A trailing bare number followed only by spaces is still the spec; the spaces are trimmed. Contrast a bare
		// number followed by prose ("13 years", cases 7-9), which is not reported.
		{"5 ", 0, 1},      // 36
		{"roll 5 ", 5, 6}, // 37
		{"13 ", 0, 2},     // 38
		// A bare number is reported only when it is the final token; a later token supersedes an earlier bare number.
		{"5 5", 2, 3}, // 39
		// A space between an operator and its operand ends the spec, matching New (which stops at the inner space and
		// drops the operand), so the span excludes the unparsed tail rather than over-reporting it.
		{"3d6+ 2", 0, 3}, // 40
		{"d6- 5", 0, 2},  // 41
		// A sign immediately followed by a multiplier has no operand of its own, so it is dangling: New drops the sign
		// and keeps the multiplier ("d6+x2" parses to "d6x2"). A contiguous span cannot reproduce that, so the spec
		// ends at the dangling sign and the trim drops it, yielding just the dice (never an interior dangling sign).
		{"d6+x2", 0, 2},  // 42
		{"d6-x2", 0, 2},  // 43
		{"3d6+x5", 0, 3}, // 44
		{"5+x2", 0, 1},   // 45
		{"d6+X2", 0, 2},  // 46 - uppercase multiplier
	} {
		desc := fmt.Sprintf("Table index %d: %s", i, one.Text)
		start, end := dice.ExtractDicePosition(one.Text)
		c.Equal(one.Start, start, desc)
		c.Equal(one.End, end, desc)
	}
}

// TestExtractedSpanIsCanonical pins the reconciliation guarantee between ExtractDicePosition and New: the span the
// extractor returns is already exactly what New parses it back into, with no trailing operator or interior space that
// New would silently drop. Before this was reconciled, e.g. "3d6+ 2" yielded the span "3d6+ 2" while New("3d6+ 2")
// parsed only "3d6".
func TestExtractedSpanIsCanonical(t *testing.T) {
	c := check.New(t)
	r := newRoller(c, nil, false, false)
	for _, text := range []string{
		"3d6+ 2", "d6- 5", "5 ", "roll 5 ", "13 ", "5 5", "d6+", "3d6x", "2d6+2x", "d6+x", "roll 3d6 for me", "d6x2",
		"5", "roll 5", "d6", "2d6+2",
		// A sign directly before a multiplier is dangling; the span must still be exactly what New parses it back into.
		"d6+x2", "d6-x2", "3d6+x5", "5+x2", "d6+X2",
	} {
		start, end := dice.ExtractDicePosition(text)
		c.True(start >= 0 && start < end, text)
		span := text[start:end]
		c.Equal(span, r.Format(r.Parse(span)), text)
	}
}
