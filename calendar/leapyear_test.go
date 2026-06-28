// Copyright (c) 2017-2026 by Richard A. Wilkes. All rights reserved.
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

// bruteSince independently counts the leap years strictly between year 1 and the given year using only Is(), serving as
// an oracle for Since(). There is no year 0, so the negative side walks -1, -2, ... year+1.
func bruteSince(ly *calendar.LeapYear, year int) int {
	count := 0
	switch {
	case year > 1:
		for y := 2; y < year; y++ {
			if ly.Is(y) {
				count++
			}
		}
	case year < -1:
		for y := -1; y > year; y-- {
			if ly.Is(y) {
				count++
			}
		}
	}
	return count
}

func TestLeapYearSinceMatchesIs(t *testing.T) {
	c := check.New(t)
	// Since() must agree with a brute-force count over Is() for every leap-rule shape, across both positive and
	// negative years. The Except-set/Unless-unset shape is the regression: Since() previously overcounted every
	// negative year by one because it unconditionally treated year -1 (magnitude 0) as a leap year.
	for _, ly := range []*calendar.LeapYear{
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
			c.Equal(bruteSince(ly, year), ly.Since(year), "Since(%d) for %+v", year, *ly)
		}
	}
}

func TestExceptOnlyCalendarNegativeDates(t *testing.T) {
	c := check.New(t)
	// A valid calendar whose leap rule sets Except but not Unless. Its negative-year date math previously broke because
	// LeapYear.Since() disagreed with Is(); e.g. NewDateByDays(-61).Month() panicked with "Unable to determine month".
	// Accessors must now round-trip across a wide span of days, including negative (BC) years.
	cal := &calendar.Calendar{
		DayZeroWeekDay: 0,
		WeekDays:       []string{"A", "B", "C"},
		Months:         []calendar.Month{{Name: "M1", Days: 30}, {Name: "M2", Days: 30}},
		Seasons:        []calendar.Season{{Name: "S", StartMonth: 1, StartDay: 1, EndMonth: 2, EndDay: 30}},
		LeapYear:       &calendar.LeapYear{Month: 1, Every: 4, Except: 8},
	}
	c.NoError(cal.Valid())

	// The exact case from the bug report no longer panics.
	c.NotPanics(func() { _ = cal.NewDateByDays(-61).Month() })

	for d := -1000; d <= 1000; d++ {
		date := cal.NewDateByDays(d)
		c.NotPanics(func() {
			year := date.Year()
			month := date.Month()
			dayInMonth := date.DayInMonth()
			daysInMonth := date.DaysInMonth()
			c.True(year != 0, "year must never be 0 (days=%d)", d)
			c.True(dayInMonth >= 1 && dayInMonth <= daysInMonth,
				"days=%d: dayInMonth %d out of range 1..%d", d, dayInMonth, daysInMonth)
			c.Equal(d, cal.MustNewDate(month, dayInMonth, year).Days,
				"days=%d did not round-trip through %d/%d/%d", d, month, dayInMonth, year)
		}, "days=%d", d)
	}
}
