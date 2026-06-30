// Copyright (c) 2017-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calendar

import (
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

func TestLeapYearSince(t *testing.T) {
	c := check.New(t)
	cal := Gregorian()
	c.Equal(0, cal.leapYearsSince(1))
	c.Equal(0, cal.leapYearsSince(2))
	c.Equal(0, cal.leapYearsSince(3))
	c.Equal(0, cal.leapYearsSince(4))
	c.Equal(1, cal.leapYearsSince(5))
	c.Equal(1, cal.leapYearsSince(6))
	c.Equal(1, cal.leapYearsSince(7))
	c.Equal(1, cal.leapYearsSince(8))
	c.Equal(2, cal.leapYearsSince(9))
	c.Equal(2, cal.leapYearsSince(10))
	c.Equal(24, cal.leapYearsSince(99))
	c.Equal(24, cal.leapYearsSince(100))
	c.Equal(24, cal.leapYearsSince(101))
	c.Equal(48, cal.leapYearsSince(199))
	c.Equal(48, cal.leapYearsSince(200))
	c.Equal(48, cal.leapYearsSince(201))
	c.Equal(72, cal.leapYearsSince(299))
	c.Equal(72, cal.leapYearsSince(300))
	c.Equal(72, cal.leapYearsSince(301))
	c.Equal(96, cal.leapYearsSince(399))
	c.Equal(96, cal.leapYearsSince(400))
	c.Equal(97, cal.leapYearsSince(401))

	c.Equal(0, cal.leapYearsSince(-1))
	c.Equal(1, cal.leapYearsSince(-2))
	c.Equal(1, cal.leapYearsSince(-3))
	c.Equal(1, cal.leapYearsSince(-4))
	c.Equal(1, cal.leapYearsSince(-5))
	c.Equal(2, cal.leapYearsSince(-6))
	c.Equal(2, cal.leapYearsSince(-7))
	c.Equal(2, cal.leapYearsSince(-8))
	c.Equal(2, cal.leapYearsSince(-9))
	c.Equal(3, cal.leapYearsSince(-10))
	c.Equal(24, cal.leapYearsSince(-96))
	c.Equal(24, cal.leapYearsSince(-97))
	c.Equal(25, cal.leapYearsSince(-98))
	c.Equal(25, cal.leapYearsSince(-100))
	c.Equal(25, cal.leapYearsSince(-101))
	c.Equal(25, cal.leapYearsSince(-102))
	c.Equal(49, cal.leapYearsSince(-200))
	c.Equal(49, cal.leapYearsSince(-201))
	c.Equal(49, cal.leapYearsSince(-202))
	c.Equal(73, cal.leapYearsSince(-300))
	c.Equal(73, cal.leapYearsSince(-301))
	c.Equal(73, cal.leapYearsSince(-302))
	c.Equal(97, cal.leapYearsSince(-400))
	c.Equal(97, cal.leapYearsSince(-401))
	c.Equal(98, cal.leapYearsSince(-402))
}

func TestLeapYearSinceMatchesIs(t *testing.T) {
	c := check.New(t)
	cfg := Gregorian().Config()
	// Since() must agree with a brute-force count over Is() for every leap-rule shape, across both positive and
	// negative years. The Except-set/Unless-unset shape is the regression: Since() previously overcounted every
	// negative year by one because it unconditionally treated year -1 (magnitude 0) as a leap year.
	for _, ly := range []*LeapYear{
		{Month: 2, Every: 4, Except: 100, Unless: 400}, // Gregorian-style: all three tiers
		{Month: 1, Every: 4, Except: 8},                // Except set, Unless unset (the bug)
		{Month: 1, Every: 2},                           // Every only
		{Month: 1, Every: 3, Except: 9, Unless: 27},    // all tiers, different moduli
		{Month: 1, Every: 5, Except: 25},               // another Except-only shape
	} {
		for year := -500; year <= 500; year++ {
			if year == 0 {
				continue
			}
			cfg.LeapYear = ly
			cal, err := New(cfg)
			c.NoError(err)
			c.Equal(bruteSince(cal, year), cal.leapYearsSince(year), "Since(%d) for %+v", year, *ly)
		}
	}
}

// bruteSince independently counts the leap years strictly between year 1 and the given year using only Is(), serving as
// an oracle for leapYearsSince(). There is no year 0, so the negative side walks -1, -2, ... year+1.
func bruteSince(cal *Calendar, year int) int {
	count := 0
	switch {
	case year > 1:
		for y := 2; y < year; y++ {
			if cal.IsLeapYear(y) {
				count++
			}
		}
	case year < -1:
		for y := -1; y > year; y-- {
			if cal.IsLeapYear(y) {
				count++
			}
		}
	}
	return count
}
