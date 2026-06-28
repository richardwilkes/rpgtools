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
	// String() must emit text that New() parses back into an equivalent Dice. The "+" separator before a positive
	// modifier is required whenever a die term was already written, including the degenerate zero-sided case (e.g.
	// "3d0+2" must not collapse to the ambiguous "3d02").
	for i, one := range []struct {
		Text     string
		Expected string
	}{
		{"3d0+2", "3d0+2"},   // 0 - regression: previously emitted "3d02" -> reparsed as "3d2"
		{"d0+2", "d0+2"},     // 1
		{"3d0-2", "3d0-2"},   // 2
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
	const huge = "99999999999999999999" // 20 nines: larger than math.MaxInt, which would otherwise wrap

	// Each numeric field saturates at math.MaxInt rather than wrapping to a negative or garbage value, and parsing
	// still continues past the oversized number.
	d := dice.New(huge + "d6")
	c.Equal(math.MaxInt, d.Count)
	c.Equal(6, d.Sides)

	d = dice.New("3d" + huge)
	c.Equal(3, d.Count)
	c.Equal(math.MaxInt, d.Sides)

	d = dice.New("d6+" + huge)
	c.Equal(math.MaxInt, d.Modifier)

	d = dice.New("2d6x" + huge)
	c.Equal(math.MaxInt, d.Multiplier)

	// Normal values are unaffected.
	d = dice.New("3d6+2x4")
	c.Equal(3, d.Count)
	c.Equal(6, d.Sides)
	c.Equal(2, d.Modifier)
	c.Equal(4, d.Multiplier)
}

// bigEvenAdjust independently computes the even-sided modifier-to-extra-dice conversion using arbitrary-precision math,
// so it stays correct even when modifier is math.MaxInt and the intermediate products would overflow int64. It mirrors
// the package rule: an even-sided die's true average is average+0.5, k = floor(2*modifier/(2*average+1)) dice are
// extracted, and those dice consume ceil(k*(2*average+1)/2) of the modifier, leaving the remainder.
func bigEvenAdjust(count, sides, modifier int) (wantCount, wantModifier int) {
	average := (sides + 1) / 2
	perPair := big.NewInt(int64(2*average + 1))
	m := big.NewInt(int64(modifier))
	k := new(big.Int).Quo(new(big.Int).Lsh(m, 1), perPair) // floor(2*modifier / perPair)
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
	// modifiers lock in the prior shipped behavior; the large/saturated ones exercise the path where the previous
	// O(modifier) loop would have hung.
	for _, sides := range []int{2, 4, 6, 8, 10, 12, 20, 100} {
		for mod := 0; mod <= 600; mod++ {
			d := dice.Dice{Count: 1, Sides: sides, Modifier: mod, Multiplier: 1}
			d.ApplyExtraDiceFromModifiers()
			wantCount, wantMod := bigEvenAdjust(1, sides, mod)
			c.Equal(wantCount, d.Count, "sides=%d mod=%d count", sides, mod)
			c.Equal(wantMod, d.Modifier, "sides=%d mod=%d modifier", sides, mod)
		}
	}
	for _, sides := range []int{2, 4, 6, 8, 100, math.MaxInt - 1} {
		for _, mod := range []int{99999, 1000000, math.MaxInt / 3, math.MaxInt - 1, math.MaxInt} {
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
	// Regression: extractValue saturates an oversized number to math.MaxInt, so dice.New("1d6+99999999999999999999")
	// yields Modifier=math.MaxInt with an even number of sides. The conversion to extra dice must be O(1); the old loop
	// removed only ~2*average per iteration, needing ~1.3e18 iterations (an effective hang). Run in a goroutine so a
	// regression fails the test rather than hanging the whole suite.
	type result struct{ count, modifier int }
	done := make(chan result, 1)
	go func() {
		d := *dice.New("1d6+99999999999999999999")
		d.ApplyExtraDiceFromModifiers()
		done <- result{d.Count, d.Modifier}
	}()
	select {
	case got := <-done:
		c.Equal(math.MaxInt, dice.New("1d6+99999999999999999999").Modifier, "modifier should saturate to MaxInt")
		// math.MaxInt is divisible by 7 (=2*average+1 for d6), so it converts to whole pairs with no remainder.
		c.Equal(2*(math.MaxInt/7)+1, got.count)
		c.Equal(0, got.modifier)
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
	} {
		desc := fmt.Sprintf("Table index %d: %s", i, one.Text)
		start, end := dice.ExtractDicePosition(one.Text)
		c.Equal(one.Start, start, desc)
		c.Equal(one.End, end, desc)
	}
}
