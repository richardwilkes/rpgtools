// Copyright ©2017-2023 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calendar_test

import (
	"testing"

	"github.com/richardwilkes/rpgtools/calendar"
	"github.com/richardwilkes/toolbox/check"
)

func TestLeapYearIs(t *testing.T) {
	ly := &calendar.LeapYear{
		Month:  2,
		Every:  4,
		Except: 100,
		Unless: 400,
	}
	check.False(t, ly.Is(1))
	check.False(t, ly.Is(2))
	check.False(t, ly.Is(3))
	check.True(t, ly.Is(4))
	check.False(t, ly.Is(5))
	check.False(t, ly.Is(6))
	check.False(t, ly.Is(7))
	check.True(t, ly.Is(8))
	check.False(t, ly.Is(9))
	check.True(t, ly.Is(96))
	check.False(t, ly.Is(100))
	check.False(t, ly.Is(200))
	check.False(t, ly.Is(300))
	check.True(t, ly.Is(400))

	check.True(t, ly.Is(-1))
	check.False(t, ly.Is(-2))
	check.False(t, ly.Is(-3))
	check.False(t, ly.Is(-4))
	check.True(t, ly.Is(-5))
	check.False(t, ly.Is(-6))
	check.False(t, ly.Is(-7))
	check.False(t, ly.Is(-8))
	check.True(t, ly.Is(-9))
	check.False(t, ly.Is(-10))
	check.True(t, ly.Is(-97))
	check.False(t, ly.Is(-101))
	check.False(t, ly.Is(-201))
	check.False(t, ly.Is(-301))
	check.True(t, ly.Is(-401))
}

func TestLeapYearSince(t *testing.T) {
	ly := &calendar.LeapYear{
		Month:  2,
		Every:  4,
		Except: 100,
		Unless: 400,
	}
	check.Equal(t, 0, ly.Since(1))
	check.Equal(t, 0, ly.Since(2))
	check.Equal(t, 0, ly.Since(3))
	check.Equal(t, 0, ly.Since(4))
	check.Equal(t, 1, ly.Since(5))
	check.Equal(t, 1, ly.Since(6))
	check.Equal(t, 1, ly.Since(7))
	check.Equal(t, 1, ly.Since(8))
	check.Equal(t, 2, ly.Since(9))
	check.Equal(t, 2, ly.Since(10))
	check.Equal(t, 24, ly.Since(99))
	check.Equal(t, 24, ly.Since(100))
	check.Equal(t, 24, ly.Since(101))
	check.Equal(t, 48, ly.Since(199))
	check.Equal(t, 48, ly.Since(200))
	check.Equal(t, 48, ly.Since(201))
	check.Equal(t, 72, ly.Since(299))
	check.Equal(t, 72, ly.Since(300))
	check.Equal(t, 72, ly.Since(301))
	check.Equal(t, 96, ly.Since(399))
	check.Equal(t, 96, ly.Since(400))
	check.Equal(t, 97, ly.Since(401))

	check.Equal(t, 0, ly.Since(-1))
	check.Equal(t, 1, ly.Since(-2))
	check.Equal(t, 1, ly.Since(-3))
	check.Equal(t, 1, ly.Since(-4))
	check.Equal(t, 1, ly.Since(-5))
	check.Equal(t, 2, ly.Since(-6))
	check.Equal(t, 2, ly.Since(-7))
	check.Equal(t, 2, ly.Since(-8))
	check.Equal(t, 2, ly.Since(-9))
	check.Equal(t, 3, ly.Since(-10))
	check.Equal(t, 24, ly.Since(-96))
	check.Equal(t, 24, ly.Since(-97))
	check.Equal(t, 25, ly.Since(-98))
	check.Equal(t, 25, ly.Since(-100))
	check.Equal(t, 25, ly.Since(-101))
	check.Equal(t, 25, ly.Since(-102))
	check.Equal(t, 49, ly.Since(-200))
	check.Equal(t, 49, ly.Since(-201))
	check.Equal(t, 49, ly.Since(-202))
	check.Equal(t, 73, ly.Since(-300))
	check.Equal(t, 73, ly.Since(-301))
	check.Equal(t, 73, ly.Since(-302))
	check.Equal(t, 97, ly.Since(-400))
	check.Equal(t, 97, ly.Since(-401))
	check.Equal(t, 98, ly.Since(-402))
}
