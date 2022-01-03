// Copyright ©2017-2022 by Richard A. Wilkes. All rights reserved.
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

	"github.com/stretchr/testify/assert"

	"github.com/richardwilkes/rpgtools/calendar"
)

func TestLeapYearIs(t *testing.T) {
	ly := &calendar.LeapYear{
		Month:  2,
		Every:  4,
		Except: 100,
		Unless: 400,
	}
	assert.False(t, ly.Is(1))
	assert.False(t, ly.Is(2))
	assert.False(t, ly.Is(3))
	assert.True(t, ly.Is(4))
	assert.False(t, ly.Is(5))
	assert.False(t, ly.Is(6))
	assert.False(t, ly.Is(7))
	assert.True(t, ly.Is(8))
	assert.False(t, ly.Is(9))
	assert.True(t, ly.Is(96))
	assert.False(t, ly.Is(100))
	assert.False(t, ly.Is(200))
	assert.False(t, ly.Is(300))
	assert.True(t, ly.Is(400))

	assert.True(t, ly.Is(-1))
	assert.False(t, ly.Is(-2))
	assert.False(t, ly.Is(-3))
	assert.False(t, ly.Is(-4))
	assert.True(t, ly.Is(-5))
	assert.False(t, ly.Is(-6))
	assert.False(t, ly.Is(-7))
	assert.False(t, ly.Is(-8))
	assert.True(t, ly.Is(-9))
	assert.False(t, ly.Is(-10))
	assert.True(t, ly.Is(-97))
	assert.False(t, ly.Is(-101))
	assert.False(t, ly.Is(-201))
	assert.False(t, ly.Is(-301))
	assert.True(t, ly.Is(-401))
}

func TestLeapYearSince(t *testing.T) {
	ly := &calendar.LeapYear{
		Month:  2,
		Every:  4,
		Except: 100,
		Unless: 400,
	}
	assert.Equal(t, 0, ly.Since(1))
	assert.Equal(t, 0, ly.Since(2))
	assert.Equal(t, 0, ly.Since(3))
	assert.Equal(t, 0, ly.Since(4))
	assert.Equal(t, 1, ly.Since(5))
	assert.Equal(t, 1, ly.Since(6))
	assert.Equal(t, 1, ly.Since(7))
	assert.Equal(t, 1, ly.Since(8))
	assert.Equal(t, 2, ly.Since(9))
	assert.Equal(t, 2, ly.Since(10))
	assert.Equal(t, 24, ly.Since(99))
	assert.Equal(t, 24, ly.Since(100))
	assert.Equal(t, 24, ly.Since(101))
	assert.Equal(t, 48, ly.Since(199))
	assert.Equal(t, 48, ly.Since(200))
	assert.Equal(t, 48, ly.Since(201))
	assert.Equal(t, 72, ly.Since(299))
	assert.Equal(t, 72, ly.Since(300))
	assert.Equal(t, 72, ly.Since(301))
	assert.Equal(t, 96, ly.Since(399))
	assert.Equal(t, 96, ly.Since(400))
	assert.Equal(t, 97, ly.Since(401))

	assert.Equal(t, 0, ly.Since(-1))
	assert.Equal(t, 1, ly.Since(-2))
	assert.Equal(t, 1, ly.Since(-3))
	assert.Equal(t, 1, ly.Since(-4))
	assert.Equal(t, 1, ly.Since(-5))
	assert.Equal(t, 2, ly.Since(-6))
	assert.Equal(t, 2, ly.Since(-7))
	assert.Equal(t, 2, ly.Since(-8))
	assert.Equal(t, 2, ly.Since(-9))
	assert.Equal(t, 3, ly.Since(-10))
	assert.Equal(t, 24, ly.Since(-96))
	assert.Equal(t, 24, ly.Since(-97))
	assert.Equal(t, 25, ly.Since(-98))
	assert.Equal(t, 25, ly.Since(-100))
	assert.Equal(t, 25, ly.Since(-101))
	assert.Equal(t, 25, ly.Since(-102))
	assert.Equal(t, 49, ly.Since(-200))
	assert.Equal(t, 49, ly.Since(-201))
	assert.Equal(t, 49, ly.Since(-202))
	assert.Equal(t, 73, ly.Since(-300))
	assert.Equal(t, 73, ly.Since(-301))
	assert.Equal(t, 73, ly.Since(-302))
	assert.Equal(t, 97, ly.Since(-400))
	assert.Equal(t, 97, ly.Since(-401))
	assert.Equal(t, 98, ly.Since(-402))
}
