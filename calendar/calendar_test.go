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
	"bytes"
	"strings"
	"testing"

	"github.com/richardwilkes/rpgtools/calendar"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestDateSeason(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian() // Winter wraps the year boundary (11/1-2/28); the other three seasons do not.
	assertSeason := func(month, day, year int, want string) {
		t.Helper()
		s, ok := cal.MustNewDate(month, day, year).Season()
		c.True(ok)
		c.Equal(want, s.Name)
	}

	// Interior of each season, including the tail of the year that belongs to the wrapping winter.
	assertSeason(1, 15, 2017, "Winter")
	assertSeason(4, 15, 2017, "Spring")
	assertSeason(7, 15, 2017, "Summer")
	assertSeason(10, 15, 2017, "Fall")
	assertSeason(12, 15, 2017, "Winter")

	// Boundaries around the wrapping winter and its neighbors.
	assertSeason(11, 1, 2017, "Winter") // winter start
	assertSeason(10, 31, 2017, "Fall")  // the day before winter starts
	assertSeason(2, 28, 2017, "Winter") // winter end in a non-leap year
	assertSeason(3, 1, 2017, "Spring")  // the day after winter ends

	// Leap-day edge: winter ends 2/28, but the end-of-month boundary extends through February, so 2/29 in a leap year
	// still falls in winter rather than in no season at all.
	assertSeason(2, 29, 2016, "Winter")
	assertSeason(3, 1, 2016, "Spring")
}

func TestDateSeasonGapsAndOverlap(t *testing.T) {
	c := check.New(t)
	weekDays := []string{"A", "B", "C", "D", "E", "F", "G"}
	months := []calendar.Month{{Name: "One", Days: 30}, {Name: "Two", Days: 30}, {Name: "Three", Days: 30}}

	// A calendar whose lone season covers only the middle month leaves the first and last months in no season.
	gapCal, err := calendar.New(&calendar.Config{
		WeekDays: weekDays,
		Months:   months,
		Seasons:  []calendar.Season{{Name: "Mid", StartMonth: 2, StartDay: 1, EndMonth: 2, EndDay: 30}},
	})
	c.NoError(err)
	s, ok := gapCal.MustNewDate(2, 15, 1).Season()
	c.True(ok)
	c.Equal("Mid", s.Name)
	_, ok = gapCal.MustNewDate(1, 15, 1).Season()
	c.False(ok)
	_, ok = gapCal.MustNewDate(3, 15, 1).Season()
	c.False(ok)

	// A calendar whose two seasons overlap in the middle month: the first one in declaration order wins there.
	overlapCal, err := calendar.New(&calendar.Config{
		WeekDays: weekDays,
		Months:   months,
		Seasons: []calendar.Season{
			{Name: "First", StartMonth: 1, StartDay: 1, EndMonth: 2, EndDay: 30},
			{Name: "Second", StartMonth: 2, StartDay: 1, EndMonth: 3, EndDay: 30},
		},
	})
	c.NoError(err)
	s, ok = overlapCal.MustNewDate(2, 15, 1).Season()
	c.True(ok)
	c.Equal("First", s.Name)
	s, ok = overlapCal.MustNewDate(3, 15, 1).Season()
	c.True(ok)
	c.Equal("Second", s.Name)
}

func TestTextCalendarMonthGolden(t *testing.T) {
	c := check.New(t)
	greg := calendar.Gregorian()
	// Golden output captured from the pre-optimization implementation; building each day by incrementing Days instead
	// of reconstructing a fresh Date per day must produce byte-for-byte identical text.
	for i, one := range []struct {
		date     calendar.Date
		expected string
	}{
		{
			greg.MustNewDate(1, 1, 2017),
			"1: January\n S  M  T  W  T  F  S\n 1  2  3  4  5  6  7\n 8  9 10 11 12 13 14\n" +
				"15 16 17 18 19 20 21\n22 23 24 25 26 27 28\n29 30 31 \n",
		},
		{ // Leap February has 29 days.
			greg.MustNewDate(2, 1, 2016),
			"2: February\n S  M  T  W  T  F  S\n    1  2  3  4  5  6\n 7  8  9 10 11 12 13\n" +
				"14 15 16 17 18 19 20\n21 22 23 24 25 26 27\n28 29 \n",
		},
		{ // Negative year exercises the year-convergence path.
			greg.MustNewDate(9, 1, -44),
			"9: September\n S  M  T  W  T  F  S\n 1  2  3  4  5  6  7\n 8  9 10 11 12 13 14\n" +
				"15 16 17 18 19 20 21\n22 23 24 25 26 27 28\n29 30 \n",
		},
		{ // A different calendar with its own week and month structure.
			calendar.PathfinderAbsalomReckoning().MustNewDate(1, 1, 4707),
			"1: Abadius\n M  T  W  O  F  S  S\n       1  2  3  4  5\n 6  7  8  9 10 11 12\n" +
				"13 14 15 16 17 18 19\n20 21 22 23 24 25 26\n27 28 29 30 31 \n",
		},
	} {
		var buf bytes.Buffer
		one.date.TextCalendarMonth(&buf)
		c.Equal(one.expected, buf.String(), "table index %d", i)
	}
}

func TestTextCalendarMonthSpacing(t *testing.T) {
	c := check.New(t)
	// A two-digit-wide month (12 days) whose first day lands three columns in (day-zero week day 1) exercises both
	// padding paths: the week-day legend pads each abbreviation to the column width, and the first week is indented by
	// weekDay*(width+1). Pinning the exact bytes guards the strings.Repeat-based padding against off-by-one drift.
	cal, err := calendar.New(&calendar.Config{
		DayZeroWeekDay: 1,
		WeekDays:       []string{"Aardvark", "Bee", "Cat"},
		Months:         []calendar.Month{{Name: "M", Days: 12}},
		Seasons:        []calendar.Season{{Name: "Whole", StartMonth: 1, StartDay: 1, EndMonth: 1, EndDay: 12}},
	})
	c.NoError(err)
	var buf bytes.Buffer
	cal.MustNewDate(1, 1, 1).TextCalendarMonth(&buf)
	c.Equal("1: M\n A  B  C\n    1  2\n 3  4  5\n 6  7  8\n 9 10 11\n12 \n", buf.String())
}

func TestTextHoistedWidthMatchesPerMonth(t *testing.T) {
	c := check.New(t)
	// Calendar.Text now computes the day-of-month column width once and passes it to every month rather than letting
	// each month rescan the calendar (which made the loop O(months²)). The hoisted width must equal the width each
	// month's public TextCalendarMonth computes on its own, so every month block Text emits must appear verbatim in the
	// full-year output. A wrong hoisted width would change the padding and break this containment.
	for _, cal := range []*calendar.Calendar{calendar.Gregorian(), calendar.PathfinderAbsalomReckoning()} {
		const year = 2017
		var full bytes.Buffer
		cal.Text(year, &full)
		out := full.String()
		months := cal.Config().Months
		for month := 1; month <= len(months); month++ {
			var monthBuf bytes.Buffer
			cal.MustNewDate(month, 1, year).TextCalendarMonth(&monthBuf)
			c.True(strings.Contains(out, monthBuf.String()),
				"month %d block must appear verbatim in the full-year text", month)
		}
	}
}

func TestDateAccessorsRoundTrip(t *testing.T) {
	c := check.New(t)
	// The cheaper Year/Month/DayInMonth/DaysInMonth must still agree with each other and reconstruct the original day
	// count across a wide range of days, including negative years and leap years.
	for _, cal := range []*calendar.Calendar{calendar.Gregorian(), calendar.PathfinderAbsalomReckoning()} {
		for d := -1000; d <= 1000; d++ {
			date := cal.NewDateByDays(d)
			year := date.Year()
			month := date.Month()
			dayInMonth := date.DayInMonth()
			daysInMonth := date.DaysInMonth()
			c.True(year != 0, "year must never be 0 (days=%d)", d)
			c.True(dayInMonth >= 1 && dayInMonth <= daysInMonth,
				"days=%d: dayInMonth %d out of range 1..%d", d, dayInMonth, daysInMonth)
			c.Equal(d, cal.MustNewDate(month, dayInMonth, year).Days(),
				"days=%d did not round-trip through %d/%d/%d", d, month, dayInMonth, year)
		}
	}
}

func TestLeapYearIs(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	c.False(cal.IsLeapYear(1))
	c.False(cal.IsLeapYear(2))
	c.False(cal.IsLeapYear(3))
	c.True(cal.IsLeapYear(4))
	c.False(cal.IsLeapYear(5))
	c.False(cal.IsLeapYear(6))
	c.False(cal.IsLeapYear(7))
	c.True(cal.IsLeapYear(8))
	c.False(cal.IsLeapYear(9))
	c.True(cal.IsLeapYear(96))
	c.False(cal.IsLeapYear(100))
	c.False(cal.IsLeapYear(200))
	c.False(cal.IsLeapYear(300))
	c.True(cal.IsLeapYear(400))

	c.True(cal.IsLeapYear(-1))
	c.False(cal.IsLeapYear(-2))
	c.False(cal.IsLeapYear(-3))
	c.False(cal.IsLeapYear(-4))
	c.True(cal.IsLeapYear(-5))
	c.False(cal.IsLeapYear(-6))
	c.False(cal.IsLeapYear(-7))
	c.False(cal.IsLeapYear(-8))
	c.True(cal.IsLeapYear(-9))
	c.False(cal.IsLeapYear(-10))
	c.True(cal.IsLeapYear(-97))
	c.False(cal.IsLeapYear(-101))
	c.False(cal.IsLeapYear(-201))
	c.False(cal.IsLeapYear(-301))
	c.True(cal.IsLeapYear(-401))
}

func TestLeapYearValidMultiples(t *testing.T) {
	c := check.New(t)
	cfg := calendar.Config{
		Months: []calendar.Month{
			{Name: "M1", Days: 30},
			{Name: "M2", Days: 30},
			{Name: "M3", Days: 30},
			{Name: "M4", Days: 30},
		},
		WeekDays: []string{"A"},
	}
	for _, tc := range []struct {
		name string
		ly   calendar.LeapYear
		ok   bool
	}{
		{"every only", calendar.LeapYear{Month: 1, Every: 4}, true},
		{"except is a multiple of every", calendar.LeapYear{Month: 1, Every: 4, Except: 8}, true},
		{"except not a multiple of every", calendar.LeapYear{Month: 1, Every: 4, Except: 6}, false},
		{"gregorian-style multiples", calendar.LeapYear{Month: 1, Every: 4, Except: 100, Unless: 400}, true},
		{"unless is a multiple of except", calendar.LeapYear{Month: 1, Every: 3, Except: 9, Unless: 27}, true},
		{"unless not a multiple of except", calendar.LeapYear{Month: 1, Every: 4, Except: 8, Unless: 20}, false},
		{"every below 2", calendar.LeapYear{Month: 1, Every: 1}, false},
		{"except not greater than every", calendar.LeapYear{Month: 1, Every: 4, Except: 4}, false},
	} {
		cfg.LeapYear = &tc.ly
		if tc.ok {
			c.NoError(cfg.Valid(), tc.name)
		} else {
			c.HasError(cfg.Valid(), tc.name)
		}
	}
}

func TestExceptOnlyCalendarNegativeDates(t *testing.T) {
	c := check.New(t)
	// A valid calendar whose leap rule sets Except but not Unless. Its negative-year date math previously broke because
	// LeapYear.Since() disagreed with Is(); e.g. NewDateByDays(-61).Month() panicked with "Unable to determine month".
	// Accessors must now round-trip across a wide span of days, including negative (BC) years.
	cal, err := calendar.New(&calendar.Config{
		DayZeroWeekDay: 0,
		WeekDays:       []string{"A", "B", "C"},
		Months:         []calendar.Month{{Name: "M1", Days: 30}, {Name: "M2", Days: 30}},
		Seasons:        []calendar.Season{{Name: "S", StartMonth: 1, StartDay: 1, EndMonth: 2, EndDay: 30}},
		LeapYear:       &calendar.LeapYear{Month: 1, Every: 4, Except: 8},
	})
	c.NoError(err)

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
			c.Equal(d, cal.MustNewDate(month, dayInMonth, year).Days(),
				"days=%d did not round-trip through %d/%d/%d", d, month, dayInMonth, year)
		}, "days=%d", d)
	}
}
