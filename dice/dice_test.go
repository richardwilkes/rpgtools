// Copyright ©2017-2020 by Richard A. Wilkes. All rights reserved.
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
	"testing"

	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/xmath/rand"
	"github.com/stretchr/testify/assert"
)

func TestCreation(t *testing.T) {
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
		{"d", "d6", 1, 6, 0, 1, false, false},             // 3
		{"d8", "d8", 1, 8, 0, 1, false, false},            // 4
		{"2d", "2d6", 2, 6, 0, 1, false, false},           // 5
		{"2d4x2", "2d4x2", 2, 4, 0, 2, false, false},      // 6
		{"3d5+1", "3d5+1", 3, 5, 1, 1, false, false},      // 7
		{"abcd", "d6", 1, 6, 0, 1, false, false},          // 8
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
	} {
		desc := fmt.Sprintf("Table index %d: %s", i, one.Text)
		d := dice.New(&dice.Config{
			Randomizer:             rand.NewCryptoRand(),
			GURPSFormat:            one.GURPS,
			ExtraDiceFromModifiers: one.ExtraDiceFromModifiers,
		}, one.Text)
		assert.Equal(t, one.Expected, d.String(), desc)
		assert.Equal(t, one.Count, d.Count, desc)
		assert.Equal(t, one.Sides, d.Sides, desc)
		assert.Equal(t, one.Modifier, d.Modifier, desc)
		assert.Equal(t, one.Multiplier, d.Multiplier, desc)
	}
}

func TestApplyExtraDiceFromModifiersAfter(t *testing.T) {
	cfg := &dice.Config{Randomizer: rand.NewCryptoRand()}
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
		d := dice.New(cfg, one.Text)
		d.ApplyExtraDiceFromModifiers()
		assert.Equal(t, one.Expected, d.String(), desc)
		assert.Equal(t, one.Count, d.Count, desc)
		assert.Equal(t, one.Modifier, d.Modifier, desc)
	}
}

func TestExtractFirstPosition(t *testing.T) {
	for i, one := range []struct {
		Text string
		Pos  []int
	}{
		{"roll 3d6 for me", []int{5, 8}},            // 0
		{"d not for me, roll 2d6+2", []int{19, 24}}, // 1
		{"roll d6x2", []int{5, 9}},                  // 2
		{"roll 3dx2", []int{5, 9}},                  // 3
		{"Just text", []int(nil)},                   // 4
	} {
		assert.Equal(t, one.Pos, dice.ExtractFirstPosition(one.Text), fmt.Sprintf("Table index %d: %s", i, one.Text))
	}
}

func TestExtractAllPositions(t *testing.T) {
	for i, one := range []struct {
		Text string
		Pos  [][]int
	}{
		{"roll 1d5 then 5d8", [][]int{{5, 8}, {14, 17}}},                // 0
		{"roll 2d4, 3d6+1x2, 4d8", [][]int{{5, 8}, {10, 17}, {19, 22}}}, // 1
		{"Just text", [][]int(nil)},                                     // 2
	} {
		assert.Equal(t, one.Pos, dice.ExtractAllPositions(one.Text, -1), fmt.Sprintf("Table index %d: %s", i, one.Text))
	}
}
