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
)

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
		{"x3", "0x3", 0, 0, 0, 3, false, false},           // 25
		{"4", "4", 0, 0, 4, 1, false, false},              // 26
	} {
		desc := fmt.Sprintf("Table index %d: %s", i, one.Text)
		d := dice.New(one.Text)
		dice.GURPSFormat = one.GURPS
		c.Equal(one.Expected, d.StringExtra(one.ExtraDiceFromModifiers), desc)
		c.Equal(one.Count, d.Count, desc)
		c.Equal(one.Sides, d.Sides, desc)
		c.Equal(one.Modifier, d.Modifier, desc)
		c.Equal(one.Multiplier, d.Multiplier, desc)
	}
	dice.GURPSFormat = false
}

func TestApplyExtraDiceFromModifiersAfter(t *testing.T) {
	c := check.New(t)
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
		d := dice.New(one.Text)
		d.ApplyExtraDiceFromModifiers()
		c.Equal(one.Expected, d.String(), desc)
		c.Equal(one.Count, d.Count, desc)
		c.Equal(one.Modifier, d.Modifier, desc)
	}
}

//nolint:goconst // The tests are more readable without constants for duplicated string
func TestStringRoundTrip(t *testing.T) {
	c := check.New(t)
	// String() must emit text that New() parses back into an equivalent Dice. A "+" separator precedes a positive
	// modifier whenever a die term was written, so a real spec like "d6+2" is never emitted as the ambiguous "d62". A
	// degenerate spec that rolls no dice (a zero count or zero sides) has no die term and collapses to just its
	// modifier, so "3d0+2", "d0+2" and "0d6+2" all render as "2".
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
		{"x3", "0x3"},        // 9
	} {
		desc := fmt.Sprintf("Table index %d: %s", i, one.Text)
		d := dice.New(one.Text)
		s := d.String()
		c.Equal(one.Expected, s, desc)
		c.True(d.IsEquivalent(dice.New(s)), "%s: %q did not round-trip", desc, s)
	}
}

func TestNoDiceCanonicalizesToModifier(t *testing.T) {
	c := check.New(t)
	// A Dice that rolls no dice has no meaningful Count or Sides, whichever happens to be zero. Normalize collapses
	// both to 0 so every such spec has one canonical form: it renders as just its modifier and compares equal to any
	// other no-dice spec with the same modifier. The parser never produces a zero Count with non-zero Sides, but
	// exported fields let any caller construct one, and "Nd0" specs reach here through New.
	modifierTwo := &dice.Dice{Count: 0, Sides: 0, Modifier: 2, Multiplier: 1}
	for _, d := range []*dice.Dice{
		{Count: 0, Sides: 6, Modifier: 2, Multiplier: 1}, // zero count, non-zero sides
		{Count: 3, Sides: 0, Modifier: 2, Multiplier: 1}, // non-zero count, zero sides
		dice.New("0d6+2"),
		dice.New("3d0+2"),
	} {
		c.Equal("2", d.String(), "%+v", *d)
		c.True(d.IsEquivalent(dice.New("2")), "%+v did not round-trip to 2", *d)
		c.True(d.IsEquivalent(modifierTwo), "%+v not equivalent to bare modifier", *d)
		n := *d
		n.Normalize()
		c.Equal(0, n.Count, "%+v", *d)
		c.Equal(0, n.Sides, "%+v", *d)
		c.Equal(2, n.Modifier, "%+v", *d)
	}

	// With no modifier either, a no-dice spec is the canonical empty spec "0".
	empty := &dice.Dice{Count: 0, Sides: 20, Modifier: 0, Multiplier: 1}
	c.Equal("0", empty.String())
	c.True(empty.IsEquivalent(dice.New("0")))

	// A real spec (both Count and Sides non-zero) is left untouched.
	withDice := dice.New("3d6+2")
	c.Equal("3d6+2", withDice.String())
	c.Equal(3, withDice.Count)
	c.Equal(6, withDice.Sides)
}

func TestRollSingleSided(t *testing.T) {
	c := check.New(t)
	// One-sided dice are deterministic, so a roll must match the min/max and must
	// include any modifier rather than discarding it.
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
		d := dice.New(one.Text)
		c.Equal(one.Expected, d.Roll(false), desc)
		c.Equal(one.Expected, d.Minimum(false), desc)
		c.Equal(one.Expected, d.Maximum(false), desc)
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
	// Regression: a roll iterates Count times, so Count must be clamped to dice.MaxValue before the loop runs;
	// otherwise an enormous count is an effective hang. Cover both the parsed spec (extractValue caps the number) and a
	// field set directly to math.MaxInt, which bypasses the parser's cap and relies solely on the clamp inside the
	// roll. Run each roll in a goroutine so a regression times out rather than hanging the whole suite.
	for i, d := range []*dice.Dice{
		dice.New("99999999999999999999d6"),
		{Count: math.MaxInt, Sides: 6, Multiplier: 1},
	} {
		desc := fmt.Sprintf("case %d", i)
		minimum := d.Minimum(false)
		maximum := d.Maximum(false)
		done := make(chan [2]int, 1)
		go func() {
			// Roll() (the public entry point at line 230) uses crypto rand, so only its range can be checked.
			// RollWithRandomizer pinned to the top face must equal Maximum(), proving every clamped die was rolled.
			done <- [2]int{d.Roll(false), d.RollWithRandomizer(topFaceRandomizer{}, false)}
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

	// Differences that normalize away (a sub-1 multiplier becomes 1) are still equivalent.
	a := &dice.Dice{Count: 1, Sides: 6, Modifier: 2, Multiplier: 1}
	b := &dice.Dice{Count: 1, Sides: 6, Modifier: 2, Multiplier: 0}
	c.True(a.IsEquivalent(b))

	// Genuinely different dice are not equivalent.
	c.False(a.IsEquivalent(&dice.Dice{Count: 2, Sides: 6, Modifier: 2, Multiplier: 1}))

	// Regression: comparing must not mutate either operand, even when their fields need normalizing.
	left := &dice.Dice{Count: -3, Sides: 6, Modifier: 0, Multiplier: 1}
	right := &dice.Dice{Count: 5, Sides: 0, Modifier: 0, Multiplier: 0}
	leftCopy := *left
	rightCopy := *right
	left.IsEquivalent(right)
	c.Equal(leftCopy, *left, "receiver was mutated by IsEquivalent")
	c.Equal(rightCopy, *right, "argument was mutated by IsEquivalent")
}

func TestPoolProbability(t *testing.T) {
	c := check.New(t)
	d := &dice.Dice{Count: 3, Sides: 6}

	// Regression: a non-positive target is met by every roll, so the probability must be exactly 1,
	// never greater than 1 as it was previously (e.g. 1.0046 for target 0).
	for _, target := range []int{0, -1, -100} {
		c.Equal(1.0, d.PoolProbability(target), "target %d", target)
	}

	// A target of 1 is met by every face.
	c.Equal(1.0, d.PoolProbability(1))

	// A target beyond the number of sides is impossible.
	c.Equal(0.0, d.PoolProbability(7))

	// No dice, or a zero-sided die (which cannot roll), yields 0 rather than a division by zero.
	c.Equal(0.0, (&dice.Dice{Count: 0, Sides: 6}).PoolProbability(3))
	c.Equal(0.0, (&dice.Dice{Count: 3, Sides: 0}).PoolProbability(3))
	c.Equal(0.0, (&dice.Dice{Count: 3, Sides: 0}).PoolProbability(0))

	// A representative interior value: 3d6 rolling at least one 6 is 1-(5/6)^3 = 91/216.
	c.True(math.Abs(d.PoolProbability(6)-91.0/216.0) < 1e-12, "3d6 >=6 probability = %v, want ~%v",
		d.PoolProbability(6), 91.0/216.0)

	// Across the valid range the probability stays within [0,1] and strictly decreases as the target
	// rises.
	prev := 2.0
	for target := 1; target <= 6; target++ {
		p := d.PoolProbability(target)
		c.True(p >= 0 && p <= 1, "target %d produced out-of-range probability %v", target, p)
		c.True(p < prev, "probability did not decrease at target %d: %v >= %v", target, p, prev)
		prev = p
	}
}

func TestExtractValueOverflow(t *testing.T) {
	c := check.New(t)
	const huge = "99999999999999999999" // 20 nines: far larger than the field cap

	// Each numeric field caps at dice.MaxValue rather than wrapping or exceeding the permitted range, and parsing still
	// continues past the oversized number.
	d := dice.New(huge + "d6")
	c.Equal(dice.MaxValue, d.Count)
	c.Equal(6, d.Sides)

	d = dice.New("3d" + huge)
	c.Equal(3, d.Count)
	c.Equal(dice.MaxValue, d.Sides)

	d = dice.New("d6+" + huge)
	c.Equal(dice.MaxValue, d.Modifier)

	// A negative modifier caps at dice.MinValue rather than wrapping past the permitted range.
	d = dice.New("d6-" + huge)
	c.Equal(dice.MinValue, d.Modifier)

	d = dice.New("2d6x" + huge)
	c.Equal(dice.MaxValue, d.Multiplier)

	// Normal values are unaffected.
	d = dice.New("3d6+2x4")
	c.Equal(3, d.Count)
	c.Equal(6, d.Sides)
	c.Equal(2, d.Modifier)
	c.Equal(4, d.Multiplier)
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
	// Converting modifiers to extra dice for even-sided dice must match an independent reference exactly. The small
	// modifiers lock in the prior shipped behavior; the large ones (up to the field cap) exercise the path where the
	// previous O(modifier) loop would have hung.
	for _, sides := range []int{2, 4, 6, 8, 10, 12, 20, 100} {
		for mod := 0; mod <= 600; mod++ {
			d := dice.Dice{Count: 1, Sides: sides, Modifier: mod, Multiplier: 1}
			d.ApplyExtraDiceFromModifiers()
			wantCount, wantMod := bigEvenAdjust(1, sides, mod)
			c.Equal(wantCount, d.Count, "sides=%d mod=%d count", sides, mod)
			c.Equal(wantMod, d.Modifier, "sides=%d mod=%d modifier", sides, mod)
		}
	}
	// Exercise the largest in-range even-sided value; mask off the low bit rather than assuming dice.MaxValue is even.
	largestEvenSides := dice.MaxValue &^ 1
	for _, sides := range []int{2, 4, 6, 8, 100, largestEvenSides} {
		for _, mod := range []int{99999, 500000, dice.MaxValue / 3, dice.MaxValue - 1, dice.MaxValue} {
			d := dice.Dice{Count: 1, Sides: sides, Modifier: mod, Multiplier: 1}
			d.ApplyExtraDiceFromModifiers()
			wantCount, wantMod := bigEvenAdjust(1, sides, mod)
			c.Equal(wantCount, d.Count, "sides=%d mod=%d count", sides, mod)
			c.Equal(wantMod, d.Modifier, "sides=%d mod=%d modifier", sides, mod)
		}
	}
}

func TestExtraDiceEvenSidedTerminatesOnSaturatedModifier(t *testing.T) {
	c := check.New(t)
	// Regression: extractValue caps an oversized number to dice.MaxValue, so dice.New("1d6+99999999999999999999")
	// yields Modifier=dice.MaxValue with an even number of sides. The conversion to extra dice must be O(1); the old
	// loop removed only ~2*average per iteration, which is millions of iterations even at the cap. Run in a goroutine
	// so a regression fails the test rather than hanging the whole suite.
	type result struct{ count, modifier int }
	done := make(chan result, 1)
	go func() {
		d := *dice.New("1d6+99999999999999999999")
		d.ApplyExtraDiceFromModifiers()
		done <- result{d.Count, d.Modifier}
	}()
	select {
	case got := <-done:
		c.Equal(dice.MaxValue, dice.New("1d6+99999999999999999999").Modifier, "modifier should cap to MaxValue")
		// dice.MaxValue for a d6 (perPair=7) converts to whole pairs plus the remainder dice.MaxValue%7.
		c.Equal(1+2*(dice.MaxValue/7), got.count)
		c.Equal(dice.MaxValue%7, got.modifier)
	case <-time.After(5 * time.Second):
		t.Fatal("ApplyExtraDiceFromModifiers did not terminate: the O(modifier) hang has regressed")
	}
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
	} {
		desc := fmt.Sprintf("Table index %d: %s", i, one.Text)
		start, end := dice.ExtractDicePosition(one.Text)
		c.Equal(one.Start, start, desc)
		c.Equal(one.End, end, desc)
	}
}
