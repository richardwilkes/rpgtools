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
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestLeapYearIs(t *testing.T) {
	c := check.New(t)
	ly := &calendar.LeapYear{
		Month:  2,
		Every:  4,
		Except: 100,
		Unless: 400,
	}
	c.False(ly.Is(1))
	c.False(ly.Is(2))
	c.False(ly.Is(3))
	c.True(ly.Is(4))
	c.False(ly.Is(5))
	c.False(ly.Is(6))
	c.False(ly.Is(7))
	c.True(ly.Is(8))
	c.False(ly.Is(9))
	c.True(ly.Is(96))
	c.False(ly.Is(100))
	c.False(ly.Is(200))
	c.False(ly.Is(300))
	c.True(ly.Is(400))

	c.True(ly.Is(-1))
	c.False(ly.Is(-2))
	c.False(ly.Is(-3))
	c.False(ly.Is(-4))
	c.True(ly.Is(-5))
	c.False(ly.Is(-6))
	c.False(ly.Is(-7))
	c.False(ly.Is(-8))
	c.True(ly.Is(-9))
	c.False(ly.Is(-10))
	c.True(ly.Is(-97))
	c.False(ly.Is(-101))
	c.False(ly.Is(-201))
	c.False(ly.Is(-301))
	c.True(ly.Is(-401))
}

func TestLeapYearSince(t *testing.T) {
	c := check.New(t)
	ly := &calendar.LeapYear{
		Month:  2,
		Every:  4,
		Except: 100,
		Unless: 400,
	}
	c.Equal(0, ly.Since(1))
	c.Equal(0, ly.Since(2))
	c.Equal(0, ly.Since(3))
	c.Equal(0, ly.Since(4))
	c.Equal(1, ly.Since(5))
	c.Equal(1, ly.Since(6))
	c.Equal(1, ly.Since(7))
	c.Equal(1, ly.Since(8))
	c.Equal(2, ly.Since(9))
	c.Equal(2, ly.Since(10))
	c.Equal(24, ly.Since(99))
	c.Equal(24, ly.Since(100))
	c.Equal(24, ly.Since(101))
	c.Equal(48, ly.Since(199))
	c.Equal(48, ly.Since(200))
	c.Equal(48, ly.Since(201))
	c.Equal(72, ly.Since(299))
	c.Equal(72, ly.Since(300))
	c.Equal(72, ly.Since(301))
	c.Equal(96, ly.Since(399))
	c.Equal(96, ly.Since(400))
	c.Equal(97, ly.Since(401))

	c.Equal(0, ly.Since(-1))
	c.Equal(1, ly.Since(-2))
	c.Equal(1, ly.Since(-3))
	c.Equal(1, ly.Since(-4))
	c.Equal(1, ly.Since(-5))
	c.Equal(2, ly.Since(-6))
	c.Equal(2, ly.Since(-7))
	c.Equal(2, ly.Since(-8))
	c.Equal(2, ly.Since(-9))
	c.Equal(3, ly.Since(-10))
	c.Equal(24, ly.Since(-96))
	c.Equal(24, ly.Since(-97))
	c.Equal(25, ly.Since(-98))
	c.Equal(25, ly.Since(-100))
	c.Equal(25, ly.Since(-101))
	c.Equal(25, ly.Since(-102))
	c.Equal(49, ly.Since(-200))
	c.Equal(49, ly.Since(-201))
	c.Equal(49, ly.Since(-202))
	c.Equal(73, ly.Since(-300))
	c.Equal(73, ly.Since(-301))
	c.Equal(73, ly.Since(-302))
	c.Equal(97, ly.Since(-400))
	c.Equal(97, ly.Since(-401))
	c.Equal(98, ly.Since(-402))
}
