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
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
	"strconv"
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

func TestApplyExtraDiceFromModifiersRespectsMaxCount(t *testing.T) {
	c := check.New(t)

	newCapped := func(maxCount int) *dice.Roller {
		cfg := dice.DefaultConfig()
		cfg.MaxCount = maxCount
		cfg.ExtraDiceFromModifiers = true
		r, err := dice.NewRoller(cfg)
		c.NoError(err)
		return r
	}

	r := newCapped(2)
	for i, one := range []struct {
		Text     string
		Count    int
		Sides    int
		Modifier int
	}{
		{"1d5+10", 2, 5, 7}, // 0 - odd sides (exact): 1d5+10 would be 4d5+1, but MaxCount caps it to 2d5+7
		{"1d6+8", 2, 6, 4},  // 1 - even sides: 1d6+8 would be 3d6+1, but MaxCount caps it to 2d6+4
		{"2d6+8", 2, 6, 8},  // 2 - already at MaxCount: nothing converts, the full modifier is retained
		{"1d6+3", 1, 6, 3},  // 3 - modifier too small to form even one die: unchanged
		{"1d6+4", 2, 6, 0},  // 4 - exactly one die fits under the cap
	} {
		desc := fmt.Sprintf("Table index %d: %s", i, one.Text)
		d := r.ApplyExtraDiceFromModifiers(r.Parse(one.Text))
		c.Equal(one.Count, d.Count, desc)
		c.Equal(one.Sides, d.Sides, desc)
		c.Equal(one.Modifier, d.Modifier, desc)
		c.True(d.Count <= 2, desc)
	}

	// With MaxCount 1 a large modifier once ballooned Count to hundreds of thousands; no dice may be added past the
	// cap.
	r1 := newCapped(1)
	d := r1.ApplyExtraDiceFromModifiers(r1.Parse("1d2+999999"))
	c.Equal(1, d.Count)
	c.Equal(2, d.Sides)
	c.Equal(999999, d.Modifier)
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
		{"x3", "0"},          // 9 - degenerate: no dice and no modifier is 0
	} {
		desc := fmt.Sprintf("Table index %d: %s", i, one.Text)
		d := r.Parse(one.Text)
		s := r.Format(d)
		c.Equal(one.Expected, s, desc)
		c.True(r.IsEquivalent(d, r.Parse(s)), "%s: %q did not round-trip", desc, s)
	}
}

func TestMarshalTextRoundTrip(t *testing.T) {
	c := check.New(t)
	// MarshalText normalizes before formatting, so a degenerate Dice marshals to its normalized form, reparses to
	// Reparsed, and re-marshals to the same text.
	for i, one := range []struct {
		Text     string
		In       dice.Dice
		Reparsed dice.Dice
	}{
		{"3", dice.Dice{Count: 0, Sides: 6, Modifier: 3, Multiplier: 1}, dice.Dice{Modifier: 3, Multiplier: 1}},                           // 0 - orphan sides dropped
		{"3", dice.Dice{Count: 0, Sides: 6, Modifier: 3, Multiplier: 0}, dice.Dice{Modifier: 3, Multiplier: 1}},                           // 1 - orphan sides with a zero multiplier
		{"d6", dice.Dice{Count: 1, Sides: 6, Modifier: 0, Multiplier: 0}, dice.Dice{Count: 1, Sides: 6, Multiplier: 1}},                   // 2 - zero multiplier on a real die
		{"2", dice.Dice{Count: 3, Sides: 0, Modifier: 2, Multiplier: 1}, dice.Dice{Modifier: 2, Multiplier: 1}},                           // 3 - orphan count dropped
		{"2d6+3", dice.Dice{Count: 2, Sides: 6, Modifier: 3, Multiplier: -5}, dice.Dice{Count: 2, Sides: 6, Modifier: 3, Multiplier: 1}},  // 4 - negative multiplier
		{"0", dice.Dice{Count: 0, Sides: 0, Modifier: 0, Multiplier: 0}, dice.Dice{Multiplier: 1}},                                        // 5 - fully empty
		{"2d6+3x4", dice.Dice{Count: 2, Sides: 6, Modifier: 3, Multiplier: 4}, dice.Dice{Count: 2, Sides: 6, Modifier: 3, Multiplier: 4}}, // 6 - well-formed, unchanged
	} {
		desc := fmt.Sprintf("Table index %d: %+v", i, one.In)
		text, err := one.In.MarshalText()
		c.NoError(err, desc)
		c.Equal(one.Text, string(text), desc)

		var back dice.Dice
		c.NoError(back.UnmarshalText(text), desc)
		c.Equal(one.Reparsed, back, desc)

		reText, err := back.MarshalText()
		c.NoError(err, desc)
		c.Equal(one.Text, string(reText), desc)
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

// topFaceRandomizer always reports the highest face, so a roll equals Maximum and proves the loop ran for every die.
type topFaceRandomizer struct{}

func (topFaceRandomizer) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	return n - 1
}

// TestExtraDiceFromModifiersChangesResults pins that ExtraDiceFromModifiers changes Roll, Minimum, Average and Maximum,
// not just Format.
func TestExtraDiceFromModifiersChangesResults(t *testing.T) {
	c := check.New(t)
	plain := newRoller(c, topFaceRandomizer{}, false, false)
	extra := newRoller(c, topFaceRandomizer{}, false, true)
	d := plain.Parse("1d6+8")

	c.Equal("d6+8", plain.Format(d))
	c.Equal(9, plain.Minimum(d))
	c.Equal(11, plain.Average(d))
	c.Equal(14, plain.Maximum(d))
	c.Equal(14, plain.Roll(d))

	c.Equal("3d6+1", extra.Format(d))
	c.Equal(4, extra.Minimum(d))
	c.Equal(11, extra.Average(d))
	c.Equal(19, extra.Maximum(d))
	c.Equal(19, extra.Roll(d))

	// With a real randomizer, the roll ranges over [4,19] rather than [9,14].
	extra = newRoller(c, nil, false, true)
	for range 100 {
		got := extra.Roll(d)
		c.True(got >= 4 && got <= 19, "roll %d outside [4,19]", got)
	}
}

func TestRollTerminatesOnHugeCount(t *testing.T) {
	c := check.New(t)
	r := newRoller(c, nil, false, false)
	rd := newRoller(c, topFaceRandomizer{}, false, false)
	// Count must be clamped to MaxCount before the roll loop. Cover both a parsed spec and a field set directly to
	// math.MaxInt, which bypasses the parser's cap. Each roll runs in a goroutine so a regression times out rather than
	// hanging the suite.
	for i, d := range []dice.Dice{
		r.Parse("99999999999999999999d6"),
		{Count: math.MaxInt, Sides: 6, Multiplier: 1},
	} {
		desc := fmt.Sprintf("case %d", i)
		minimum := r.Minimum(d)
		maximum := r.Maximum(d)
		done := make(chan [2]int, 1)
		go func() {
			// Only the range of the real roll can be checked; the top-face roll must equal Maximum, proving every
			// clamped die was rolled.
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

	// Differences that normalize away are still equivalent.
	a := dice.Dice{Count: 1, Sides: 6, Modifier: 2, Multiplier: 1}
	b := dice.Dice{Count: 1, Sides: 6, Modifier: 2, Multiplier: 0}
	c.True(r.IsEquivalent(a, b))

	c.False(r.IsEquivalent(a, dice.Dice{Count: 2, Sides: 6, Modifier: 2, Multiplier: 1}))
}

func TestPoolProbability(t *testing.T) {
	c := check.New(t)
	r := newRoller(c, nil, false, false)
	d := dice.Dice{Count: 3, Sides: 6}

	// A non-positive target is met by every roll, so the probability is exactly 1.
	for _, target := range []int{0, -1, -100} {
		c.Equal(1.0, r.PoolProbability(d, target), "target %d", target)
	}

	// A target of 1 is met by every face.
	c.Equal(1.0, r.PoolProbability(d, 1))

	// A target beyond the number of sides is impossible.
	c.Equal(0.0, r.PoolProbability(d, 7))

	// No dice, or a zero-sided die, yields 0 rather than a division by zero.
	c.Equal(0.0, r.PoolProbability(dice.Dice{Count: 0, Sides: 6}, 3))
	c.Equal(0.0, r.PoolProbability(dice.Dice{Count: 3, Sides: 0}, 3))
	c.Equal(0.0, r.PoolProbability(dice.Dice{Count: 3, Sides: 0}, 0))

	// A representative interior value: 3d6 rolling at least one 6 is 1-(5/6)^3 = 91/216.
	c.True(math.Abs(r.PoolProbability(d, 6)-91.0/216.0) < 1e-12, "3d6 >=6 probability = %v, want ~%v",
		r.PoolProbability(d, 6), 91.0/216.0)

	// The probability stays within [0,1] and strictly decreases as the target rises.
	prev := 2.0
	for target := 1; target <= 6; target++ {
		p := r.PoolProbability(d, target)
		c.True(p >= 0 && p <= 1, "target %d produced out-of-range probability %v", target, p)
		c.True(p < prev, "probability did not decrease at target %d: %v >= %v", target, p, prev)
		prev = p
	}
}

func TestPoolProbabilityHonorsExtraDiceFromModifiers(t *testing.T) {
	c := check.New(t)
	plain := newRoller(c, nil, false, false)
	extra := newRoller(c, nil, false, true)
	d := plain.Parse("1d6+8")
	c.Equal(dice.Dice{Count: 1, Sides: 6, Modifier: 8, Multiplier: 1}, d)

	// PoolProbability must see the same three-die pool Roll, Format and the rest do: 1-(5/6)^3 = 91/216.
	c.Equal("3d6+1", extra.Format(d))
	c.Equal(4, extra.Minimum(d))
	got := extra.PoolProbability(d, 6)
	c.True(math.Abs(got-91.0/216.0) < 1e-12, "3d6+1 pool >=6 probability = %v, want ~%v", got, 91.0/216.0)

	// The conversion still applies only when the flag is set.
	got = plain.PoolProbability(d, 6)
	c.True(math.Abs(got-1.0/6.0) < 1e-12, "1d6+8 pool >=6 probability = %v, want ~%v", got, 1.0/6.0)
}

func TestExtractValueOverflow(t *testing.T) {
	c := check.New(t)
	r := newRoller(c, nil, false, false)
	cfg := r.Config()
	const huge = "99999999999999999999" // 20 nines: far larger than the field cap

	// Each field caps at the Config's Max* value and parsing continues past the oversized number.
	d := r.Parse(huge + "d6")
	c.Equal(cfg.MaxCount, d.Count)
	c.Equal(6, d.Sides)

	d = r.Parse("3d" + huge)
	c.Equal(3, d.Count)
	c.Equal(cfg.MaxSides, d.Sides)

	d = r.Parse("d6+" + huge)
	c.Equal(cfg.MaxModifier, d.Modifier)

	// A negative modifier caps at -MaxModifier.
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
	// extractValue must reach the cap without overflowing along the way.
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

func TestUnmarshalTextSaturatesBareNumberSum(t *testing.T) {
	c := check.New(t)
	const fieldCap = math.MaxInt - 1 // UnmarshalText parses against maxFieldValue, one below math.MaxInt
	// The halves of a bare number ("5+3") are capped individually, so their sum must saturate rather than wrap.
	var d dice.Dice
	c.NoError(d.UnmarshalText([]byte("5000000000000000000+5000000000000000000")))
	c.Equal(fieldCap, d.Modifier)
	c.Equal(0, d.Count)

	c.NoError(d.UnmarshalText([]byte("99999999999999999999+99999999999999999999")))
	c.Equal(fieldCap, d.Modifier)

	// Exactly at the cap on either side of the sign is unchanged; one past it saturates.
	capText := strconv.Itoa(fieldCap)
	c.NoError(d.UnmarshalText([]byte(capText + "+0")))
	c.Equal(fieldCap, d.Modifier)
	c.NoError(d.UnmarshalText([]byte("0+" + capText)))
	c.Equal(fieldCap, d.Modifier)
	c.NoError(d.UnmarshalText([]byte(capText + "+1")))
	c.Equal(fieldCap, d.Modifier)
	c.NoError(d.UnmarshalText([]byte("1+" + capText)))
	c.Equal(fieldCap, d.Modifier)

	// A difference cannot overflow, and the largest magnitudes cancel exactly.
	c.NoError(d.UnmarshalText([]byte("5000000000000000000-5000000000000000000")))
	c.Equal(0, d.Modifier)
	c.NoError(d.UnmarshalText([]byte(capText + "-" + capText)))
	c.Equal(0, d.Modifier)
	c.NoError(d.UnmarshalText([]byte("1-" + capText)))
	c.Equal(1-fieldCap, d.Modifier)

	// Ordinary sums are unaffected.
	c.NoError(d.UnmarshalText([]byte("5+3")))
	c.Equal(8, d.Modifier)
	c.NoError(d.UnmarshalText([]byte("5-8")))
	c.Equal(-3, d.Modifier)
}

// TestParseClampsBareNumberSumToMaxModifier pins that the folded bare-number sum is clamped to MaxModifier, not
// MaxCount.
func TestParseClampsBareNumberSumToMaxModifier(t *testing.T) {
	c := check.New(t)
	cfg := dice.DefaultConfig()
	cfg.MaxCount = 100
	cfg.MaxModifier = 5
	r, err := dice.NewRoller(cfg)
	c.NoError(err)
	c.Equal(5, r.Parse("50+3").Modifier)
	c.Equal(5, r.Parse("50-1").Modifier)
	c.Equal(-5, r.Parse("0-50").Modifier)
	c.Equal(4, r.Parse("3+1").Modifier)
	c.Equal(0, r.Parse("50+3").Count)
	// The clamp applies to the sum, not each half: 50-60 is -10, which clamps to -5.
	c.Equal(-5, r.Parse("50-60").Modifier)
}

// TestParseBareNumberNotCappedAtMaxCount pins that a bare number, being a modifier, is not capped at MaxCount, so
// Format followed by Parse round-trips when MaxCount is below MaxModifier.
func TestParseBareNumberNotCappedAtMaxCount(t *testing.T) {
	c := check.New(t)
	cfg := dice.DefaultConfig()
	cfg.MaxCount = 10
	cfg.MaxModifier = 999_999
	r, err := dice.NewRoller(cfg)
	c.NoError(err)
	c.Equal(dice.Dice{Modifier: 15, Multiplier: 1}, r.Parse("15"))
	c.Equal(dice.Dice{Modifier: 100, Multiplier: 1}, r.Parse("100"))
	c.Equal(dice.Dice{Modifier: 18, Multiplier: 1}, r.Parse("15+3"))
	c.Equal(dice.Dice{Modifier: 12, Multiplier: 1}, r.Parse("15-3"))
	c.Equal(dice.Dice{Modifier: 15, Multiplier: 2}, r.Parse("15x2"))
	c.Equal(dice.Dice{Modifier: 999_999, Multiplier: 1}, r.Parse("5000000"))
	for _, d := range []dice.Dice{
		{Modifier: 15, Multiplier: 1},
		{Modifier: 999_999, Multiplier: 1},
		{Modifier: -50, Multiplier: 3},
	} {
		c.Equal(d, r.Parse(r.Format(d)), "%+v did not round-trip", d)
	}
	// A number that really is a count is still capped at MaxCount.
	c.Equal(dice.Dice{Count: 10, Sides: 6, Multiplier: 1}, r.Parse("15d6"))
}

// TestParseSignedDiceSpecAgreesWithExtractor pins that Parse treats a sign-prefixed or sign-joined dice spec as
// ExtractDicePosition does: neither the sign nor the number before it is part of the notation.
//
//nolint:goconst // The tests are more readable without constants for duplicated string
func TestParseSignedDiceSpecAgreesWithExtractor(t *testing.T) {
	c := check.New(t)
	r := newRoller(c, nil, false, false)
	for i, one := range []struct {
		Text     string
		Expected string
	}{
		{"+3d6", "3d6"},           // 0
		{"-3d6", "3d6"},           // 1
		{"12+3d6", "3d6"},         // 2
		{"12-3d6", "3d6"},         // 3
		{"12+3d6+2x3", "3d6+2x3"}, // 4
		{"0+3d6", "3d6"},          // 5
		{"+d6", "d6"},             // 6 - a leading sign with no operand before a spec
		{" -4d6-", "4d6"},         // 7
		{"3d6+2d6", "3d6+2"},      // 8 - control: a spec that already has a 'd' ends at the second one
		{"d6+d6", "d6"},           // 9 - control: a die marker directly after a sign is dangling
		{"5+d6", "5"},             // 10 - control: a bare number with a dangling sign keeps the number
		{"+2", "2"},               // 11 - control: a signed bare number is still a modifier
		{"-1", "-1"},              // 12
		{"5+3", "8"},              // 13
	} {
		desc := fmt.Sprintf("Table index %d: %s", i, one.Text)
		d := r.Parse(one.Text)
		c.Equal(one.Expected, r.Format(d), desc)
		if start, end := dice.ExtractDicePosition(one.Text); start != -1 {
			c.Equal(r.Parse(one.Text[start:end]), d, "%s: Parse disagrees with the extracted span %q", desc,
				one.Text[start:end])
		}
	}
}

// TestAverageMultipliesBeforeRounding pins that Average applies the multiplier before rounding down: 1d6x10 is 35, not
// 30.
func TestAverageMultipliesBeforeRounding(t *testing.T) {
	c := check.New(t)
	r := newRoller(c, nil, false, false)
	for i, one := range []struct {
		Dice     dice.Dice
		Expected int
	}{
		{dice.Dice{Count: 1, Sides: 6, Multiplier: 10}, 35},                // 0 - 3.5 x 10
		{dice.Dice{Count: 3, Sides: 6, Multiplier: 2}, 21},                 // 1 - 10.5 x 2
		{dice.Dice{Count: 1, Sides: 6, Multiplier: 3}, 10},                 // 2 - 10.5 rounds down
		{dice.Dice{Count: 1, Sides: 6, Modifier: -10, Multiplier: 3}, -20}, // 3 - -19.5 rounds down
		{dice.Dice{Count: 2, Sides: 6, Multiplier: 5}, 35},                 // 4 - a whole average is exact
		{dice.Dice{Count: 1, Sides: 5, Multiplier: 3}, 9},                  // 5 - odd sides have a whole average
		{dice.Dice{Count: 1, Sides: 6, Modifier: 8, Multiplier: 1}, 11},    // 6 - no multiplier: unchanged behavior
		{dice.Dice{Count: 1, Sides: 6, Multiplier: 1}, 3},                  // 7
		{dice.Dice{Count: 1, Sides: 1, Modifier: 2, Multiplier: 7}, 21},    // 8 - single-sided dice
		{dice.Dice{Modifier: 5, Multiplier: 3}, 15},                        // 9 - no dice at all
	} {
		c.Equal(one.Expected, r.Average(one.Dice), "Table index %d: %+v", i, one.Dice)
	}
}

// TestUnmarshalTextRejectsMalformedInput pins that UnmarshalText reports wholly unparseable data and trailing garbage
// as errors.
//
//nolint:goconst // The tests are more readable without constants for duplicated string
func TestUnmarshalTextRejectsMalformedInput(t *testing.T) {
	c := check.New(t)
	for _, text := range []string{
		"total nonsense",
		"abcd",
		"3d6 extra",
		"3d6+2 x3",
		"3d6+2x3y",
		"3d6+2d6", // Parse stops at the second 'd'
		"5+d6",    // the dangling sign is consumed, the d6 is not
		"12 3d6",
	} {
		d := dice.Dice{Count: 2, Sides: 4, Modifier: 1, Multiplier: 1}
		c.HasError(d.UnmarshalText([]byte(text)), "%q", text)
		c.Equal(dice.Dice{Count: 2, Sides: 4, Modifier: 1, Multiplier: 1}, d, "%q", text)
	}

	// Well-formed text, including whatever Parse consumes in full and empty text, is accepted.
	for i, one := range []struct {
		Text     string
		Expected dice.Dice
	}{
		{"3d6+2x3", dice.Dice{Count: 3, Sides: 6, Modifier: 2, Multiplier: 3}},   // 0
		{" 1d6+2x3 ", dice.Dice{Count: 1, Sides: 6, Modifier: 2, Multiplier: 3}}, // 1
		{"3d6+", dice.Dice{Count: 3, Sides: 6, Multiplier: 1}},                   // 2
		{"3d", dice.Dice{Count: 3, Sides: 6, Multiplier: 1}},                     // 3
		{"5x2", dice.Dice{Modifier: 5, Multiplier: 2}},                           // 4
		{"12+3d6", dice.Dice{Count: 3, Sides: 6, Multiplier: 1}},                 // 5
		{"-3d6", dice.Dice{Count: 3, Sides: 6, Multiplier: 1}},                   // 6
		{"-1", dice.Dice{Modifier: -1, Multiplier: 1}},                           // 7
		{"0", dice.Dice{Multiplier: 1}},                                          // 8
		{"", dice.Dice{Multiplier: 1}},                                           // 9
		{"  ", dice.Dice{Multiplier: 1}},                                         // 10
	} {
		var d dice.Dice
		c.NoError(d.UnmarshalText([]byte(one.Text)), "Table index %d: %q", i, one.Text)
		c.Equal(one.Expected, d, "Table index %d: %q", i, one.Text)
	}
}

// TestMarshalTextUsesDefaultGURPSFormat pins that MarshalText honors the default GURPSFormat without cloning the
// default Config: the only allocation is the returned []byte.
func TestMarshalTextUsesDefaultGURPSFormat(t *testing.T) {
	c := check.New(t)
	original := dice.DefaultConfig()
	defer func() { c.NoError(dice.SetDefaultConfig(original)) }()
	d := dice.Dice{Count: 1, Sides: 6, Multiplier: 1}
	for _, gurps := range []bool{false, true} {
		cfg := dice.DefaultConfig()
		cfg.GURPSFormat = gurps
		c.NoError(dice.SetDefaultConfig(cfg))
		want := "d6"
		if gurps {
			want = "1d"
		}
		var text []byte
		var err error
		allocs := testing.AllocsPerRun(100, func() { text, err = d.MarshalText() })
		c.NoError(err, "GURPSFormat %v", gurps)
		c.Equal(want, string(text), "GURPSFormat %v", gurps)
		c.Equal(1.0, allocs, "GURPSFormat %v", gurps)
	}
}

// TestZeroRollerFetchesDefaultConfigOnce pins that a zero-value or nil Roller takes exactly one copy of the default
// Config per exported call: one allocation more than a configured Roller.
func TestZeroRollerFetchesDefaultConfigOnce(t *testing.T) {
	c := check.New(t)
	configured := newRoller(c, nil, false, false)
	var zero dice.Roller
	var nilRoller *dice.Roller
	d := dice.Dice{Count: 3, Sides: 6, Modifier: 2, Multiplier: 1}
	for _, one := range []struct {
		call func(r *dice.Roller)
		name string
	}{
		{func(r *dice.Roller) { r.Roll(d) }, "Roll"},
		{func(r *dice.Roller) { r.Format(d) }, "Format"},
		{func(r *dice.Roller) { r.Parse("3d6+2") }, "Parse"},
		{func(r *dice.Roller) { r.Normalize(d) }, "Normalize"},
		{func(r *dice.Roller) { r.ApplyExtraDiceFromModifiers(d) }, "ApplyExtraDiceFromModifiers"},
		{func(r *dice.Roller) { r.IsEquivalent(d, d) }, "IsEquivalent"},
		{func(r *dice.Roller) { r.Minimum(d) }, "Minimum"},
		{func(r *dice.Roller) { r.Average(d) }, "Average"},
		{func(r *dice.Roller) { r.Maximum(d) }, "Maximum"},
		{func(r *dice.Roller) { r.PoolProbability(d, 4) }, "PoolProbability"},
	} {
		want := testing.AllocsPerRun(100, func() { one.call(configured) }) + 1
		c.Equal(want, testing.AllocsPerRun(100, func() { one.call(&zero) }), one.name)
		c.Equal(want, testing.AllocsPerRun(100, func() { one.call(nilRoller) }), one.name+" on a nil Roller")
	}
	// Config returns the copy DefaultConfig made rather than cloning it again.
	c.Equal(1.0, testing.AllocsPerRun(100, func() { zero.Config() }))
	c.Equal(1.0, testing.AllocsPerRun(100, func() { nilRoller.Config() }))
}

func TestDiceHash(t *testing.T) {
	c := check.New(t)
	digest := func(d *dice.Dice) []byte {
		h := sha256.New()
		d.Hash(h)
		return h.Sum(nil)
	}
	base := dice.Dice{Count: 3, Sides: 6, Modifier: 2, Multiplier: 1}
	want := digest(&base)
	c.Equal(want, digest(&dice.Dice{Count: 3, Sides: 6, Modifier: 2, Multiplier: 1}))
	// Every field contributes.
	c.NotEqual(want, digest(&dice.Dice{Count: 3, Sides: 6, Modifier: 1, Multiplier: 1}))
	c.NotEqual(want, digest(&dice.Dice{Count: 2, Sides: 6, Modifier: 2, Multiplier: 1}))
	c.NotEqual(want, digest(&dice.Dice{Count: 3, Sides: 8, Modifier: 2, Multiplier: 1}))
	c.NotEqual(want, digest(&dice.Dice{Count: 3, Sides: 6, Modifier: 2, Multiplier: 2}))
	// The fields are hashed in a fixed order.
	c.NotEqual(digest(&dice.Dice{Count: 3, Sides: 6}), digest(&dice.Dice{Count: 6, Sides: 3}))
	// A nil receiver writes nothing.
	var nilDice *dice.Dice
	h := sha256.New()
	c.NotPanics(func() { nilDice.Hash(h) })
	c.Equal(sha256.New().Sum(nil), h.Sum(nil))
}

// bigEvenAdjust computes the even-sided conversion with arbitrary-precision math as a reference: k =
// floor(2*modifier/(2*average+1)) dice are extracted, consuming ceil(k*(2*average+1)/2) of the modifier.
func bigEvenAdjust(count, sides, modifier int) (wantCount, wantModifier int) {
	average := (sides + 1) / 2
	perPair := big.NewInt(int64(2*average + 1))
	m := big.NewInt(int64(modifier))
	k := new(big.Int).Quo(new(big.Int).Lsh(m, 1), perPair)
	cost := new(big.Int).Mul(k, perPair)
	if k.Bit(0) == 1 { // the half-die rounds up
		cost.Add(cost, big.NewInt(1))
	}
	cost.Rsh(cost, 1)
	r := new(big.Int).Sub(m, cost)
	return count + int(k.Int64()), int(r.Int64())
}

func TestExtraDiceEvenSidedMatchesReference(t *testing.T) {
	c := check.New(t)
	r := newRoller(c, nil, false, false)
	cfg := r.Config()
	// The conversion must match the reference exactly; the large modifiers exercise the path where an O(modifier) loop
	// would hang.
	for _, sides := range []int{2, 4, 6, 8, 10, 12, 20, 100} {
		for mod := 0; mod <= 600; mod++ {
			d := dice.Dice{Count: 1, Sides: sides, Modifier: mod, Multiplier: 1}
			d = r.ApplyExtraDiceFromModifiers(d)
			wantCount, wantMod := bigEvenAdjust(1, sides, mod)
			c.Equal(wantCount, d.Count, "sides=%d mod=%d count", sides, mod)
			c.Equal(wantMod, d.Modifier, "sides=%d mod=%d modifier", sides, mod)
		}
	}
	largestEvenSides := cfg.MaxSides &^ 1 // do not assume MaxSides is even
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
		// A 'd' with no digit after it is not notation and must not let the trailing bare number be reported.
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
		// A 'd' inside a word is not a discarded die marker, so a trailing bare number stays reportable.
		{"read 5", 5, 6}, // 21 - 'd' at the end of a word
		{"old 5", 4, 5},  // 22 - 'd' at the end of a word
		{"add 5", 4, 5},  // 23 - 'd' at the end of a word
		{"hold 5", 5, 6}, // 24 - 'd' at the end of a word
		{"drum 5", 5, 6}, // 25 - 'd' at the start of a word (followed by a prose letter)
		{"the 5", 4, 5},  // 26 - control: a word without any 'd' already worked
		// A dangling trailing operator is excluded from the span.
		{"d6+", 0, 2},    // 27
		{"d6-", 0, 2},    // 28
		{"3d6x", 0, 3},   // 29
		{"d6+x", 0, 2},   // 30 - multiple dangling operators
		{"2d6+2x", 0, 5}, // 31 - trailing 'x' trimmed, modifier retained
		{"d6+ ", 0, 2},   // 32 - operator followed by a trailing space
		{"3d", 0, 2},     // 33 - control: a trailing 'd' (meaning d6) is a valid operand, not trimmed
		// A lone 'd' is not a spec even at the end of the text.
		{"d", -1, -1},      // 34
		{"roll d", -1, -1}, // 35
		// Trailing spaces after a bare number are trimmed; a bare number followed by prose (cases 7-9) is not reported.
		{"5 ", 0, 1},      // 36
		{"roll 5 ", 5, 6}, // 37
		{"13 ", 0, 2},     // 38
		// A bare number is reported only when it is the final token; a later token supersedes an earlier bare number.
		{"5 5", 2, 3}, // 39
		// A space between an operator and its operand ends the spec, as it does in Parse.
		{"3d6+ 2", 0, 3}, // 40
		{"d6- 5", 0, 2},  // 41
		// A sign directly followed by a multiplier is dangling, so the spec ends before it and the sign is trimmed.
		{"d6+x2", 0, 2},  // 42
		{"d6-x2", 0, 2},  // 43
		{"3d6+x5", 0, 3}, // 44
		{"5+x2", 0, 1},   // 45
		{"d6+X2", 0, 2},  // 46 - uppercase multiplier
		// A sign with no candidate before it is prose, so the scan continues; the span never starts with a sign.
		{"Deal +2d6 fire damage", 6, 9}, // 47
		{"cost + 3d6 damage", 7, 10},    // 48
		{"+3d6", 1, 4},                  // 49
		{" -4d6-", 2, 5},                // 50 - the trailing dangling '-' is still trimmed
		{"+-3d6", 2, 5},                 // 51 - repeated leading signs
		{"+3d6+2", 1, 6},                // 52 - a real modifier after the spec is kept
		{"-5x2 3d6", 5, 8},              // 53 - a signed bare number with a multiplier is skipped, not the whole scan
		// A signed bare number is never reported.
		{"+13", -1, -1},     // 54
		{"-13", -1, -1},     // 55
		{"roll -5", -1, -1}, // 56
		{"-5x2", -1, -1},    // 57
		// A die marker after a sign's operand makes the operand a count, so the span is the dice spec, not the
		// arithmetic prefix.
		{"12+3d6", 3, 6},      // 58
		{"5+3d6", 2, 5},       // 59
		{"12+3d6+2x3", 3, 10}, // 60
		{"d+5d6", 2, 5},       // 61 - after a discarded 'd', like case 15
		{"3d6+2d6", 0, 5},     // 62 - control: a spec that already has a 'd' ends at the second one, as Parse does
		{"d6+d6", 0, 2},       // 63 - control: a die marker directly after a sign is dangling, so the sign is trimmed
		{"+d6", 1, 3},         // 64 - a leading sign before a 'd' spec
		// Discarding a 'd' re-examines the character, so a second 'd' can start a new candidate.
		{"dd6", 1, 3},  // 65
		{"Dd6", 1, 3},  // 66
		{"dd", -1, -1}, // 67
		// The re-examined 'd' inherits the word status of its run.
		{"adds 5", 5, 6}, // 68
		{"dd 5", -1, -1}, // 69
		// A bare number with a multiplier is a spec in its own right, reported even when text follows it.
		{"5x2", 0, 3},             // 70
		{"5X2", 0, 3},             // 71
		{"5x2+3", 0, 3},           // 72 - the spec ends before the modifier, exactly as Parse stops there
		{"5x", 0, 1},              // 73 - dangling multiplier trimmed
		{"roll 5x2", 5, 8},        // 74
		{"roll 5x2 for me", 5, 8}, // 75
		{"x2", -1, -1},            // 76 - an orphan multiplier's operand is not a bare number (Parse("x2") is 0)
		{"5 x2", -1, -1},          // 77 - nor is it after a bare number that is not the final token
		// Any trailing whitespace counts, matching the TrimSpace Parse applies.
		{"roll 5\n", 5, 6},    // 78
		{"roll 5\t", 5, 6},    // 79
		{"5\r\n", 0, 1},       // 80
		{"5 \n ", 0, 1},       // 81
		{"5\u00a0", 0, 1},     // 82 - a non-breaking space
		{"5\n5", 2, 3},        // 83 - a later token still supersedes an earlier bare number
		{"5\tthings", -1, -1}, // 84 - and a bare number followed by prose is still not reported
		{"3d6\n", 0, 3},       // 85 - control: a real spec was never affected
	} {
		desc := fmt.Sprintf("Table index %d: %s", i, one.Text)
		start, end := dice.ExtractDicePosition(one.Text)
		c.Equal(one.Start, start, desc)
		c.Equal(one.End, end, desc)
	}
}

// TestExtractedSpanIsCanonical pins that Parse consumes the extracted span in full, so the span means exactly what
// Parse makes of it. Usually the span is also its own canonical spelling; the shorthand ("3d") and bare arithmetic
// ("5+3") rows list the canonical form and must show that their tail was consumed rather than dropped.
func TestExtractedSpanIsCanonical(t *testing.T) {
	c := check.New(t)
	r := newRoller(c, nil, false, false)
	for _, one := range []struct {
		text      string
		canonical string // Format(Parse(span)); empty means the span is its own canonical form
	}{
		{"3d6+ 2", ""},
		{"d6- 5", ""},
		{"5 ", ""},
		{"roll 5 ", ""},
		{"13 ", ""},
		{"5 5", ""},
		{"d6+", ""},
		{"3d6x", ""},
		{"2d6+2x", ""},
		{"d6+x", ""},
		{"roll 3d6 for me", ""},
		{"d6x2", ""},
		{"5", ""},
		{"roll 5", ""},
		{"d6", ""},
		{"2d6+2", ""},
		// A sign directly before a multiplier is dangling.
		{"d6+x2", ""},
		{"d6-x2", ""},
		{"3d6+x5", ""},
		{"5+x2", ""},
		{"d6+X2", ""},
		// A leading sign, or a bare number joined by a sign to a dice spec, is left out of the span.
		{"Deal +2d6 fire damage", ""},
		{"+3d6", ""},
		{" -4d6-", ""},
		{"12+3d6", ""},
		{"5+3d6", ""},
		{"12+3d6+2x3", ""},
		{"+d6", ""},
		{"dd6", ""},
		// Shorthand: a trailing 'd' means d6.
		{"3d", "3d6"},
		{"roll 3d", "3d6"},
		// Bare arithmetic formats as the sum.
		{"5+3", "8"},
		{"roll 5+3 things", "8"},
		{"10-4", "6"},
		// A bare number with a multiplier, and a bare number followed by whitespace other than a space.
		{"5x2", ""},
		{"5x2+3", ""},
		{"roll 5x2 for me", ""},
		{"roll 5\n", ""},
		{"5\t", ""},
		{"5 \n ", ""},
	} {
		start, end := dice.ExtractDicePosition(one.text)
		c.True(start >= 0 && start < end, one.text)
		span := one.text[start:end]
		parsed := r.Parse(span)
		if one.canonical == "" {
			c.Equal(span, r.Format(parsed), one.text)
			continue
		}
		c.Equal(one.canonical, r.Format(parsed), one.text)
		// The tail was consumed: parsing without it gives different dice.
		c.NotEqual(parsed, r.Parse(span[:len(span)-1]), one.text)
	}
}
