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
	"testing"

	"github.com/richardwilkes/rpgtools/calendar"
	"github.com/richardwilkes/rpgtools/calendar/pathfinder"
	"github.com/richardwilkes/toolbox/v2/check"
)

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
			pathfinder.AbsalomReckoning().MustNewDate(1, 1, 4707),
			"1: Abadius\n M  T  W  O  F  S  S\n       1  2  3  4  5\n 6  7  8  9 10 11 12\n" +
				"13 14 15 16 17 18 19\n20 21 22 23 24 25 26\n27 28 29 30 31 \n",
		},
	} {
		var buf bytes.Buffer
		one.date.TextCalendarMonth(&buf)
		c.Equal(one.expected, buf.String(), "table index %d", i)
	}
}

func TestDateAccessorsRoundTrip(t *testing.T) {
	c := check.New(t)
	// The cheaper Year/Month/DayInMonth/DaysInMonth must still agree with each other and reconstruct the original day
	// count across a wide range of days, including negative years and leap years.
	for _, cal := range []*calendar.Calendar{calendar.Gregorian(), pathfinder.AbsalomReckoning()} {
		for d := -1000; d <= 1000; d++ {
			date := cal.NewDateByDays(d)
			year := date.Year()
			month := date.Month()
			dayInMonth := date.DayInMonth()
			daysInMonth := date.DaysInMonth()
			c.True(year != 0, "year must never be 0 (days=%d)", d)
			c.True(dayInMonth >= 1 && dayInMonth <= daysInMonth,
				"days=%d: dayInMonth %d out of range 1..%d", d, dayInMonth, daysInMonth)
			c.Equal(d, cal.MustNewDate(month, dayInMonth, year).Days,
				"days=%d did not round-trip through %d/%d/%d", d, month, dayInMonth, year)
		}
	}
}
