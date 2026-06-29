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

func TestSeasonLeapDayBoundary(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian() // February (month 2) is the leap month; its base length is 28.

	// A season may legitimately start or end on the 29th of the leap month, which is a real calendar day in leap years
	// even though February's base length is 28.
	c.NoError((&calendar.Season{Name: "Winter", StartMonth: 1, StartDay: 1, EndMonth: 2, EndDay: 29}).Valid(cal))
	c.NoError((&calendar.Season{Name: "Thaw", StartMonth: 2, StartDay: 29, EndMonth: 3, EndDay: 1}).Valid(cal))

	// One day past the leap day is still rejected...
	c.HasError((&calendar.Season{Name: "Winter", StartMonth: 1, StartDay: 1, EndMonth: 2, EndDay: 30}).Valid(cal))
	c.HasError((&calendar.Season{Name: "Thaw", StartMonth: 2, StartDay: 30, EndMonth: 3, EndDay: 1}).Valid(cal))
	// ...and a non-leap month gains no extra day, so March (31 days) still rejects the 32nd.
	c.HasError((&calendar.Season{Name: "Spring", StartMonth: 3, StartDay: 1, EndMonth: 3, EndDay: 32}).Valid(cal))
}

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
	gapCal := &calendar.Calendar{
		WeekDays: weekDays,
		Months:   months,
		Seasons:  []calendar.Season{{Name: "Mid", StartMonth: 2, StartDay: 1, EndMonth: 2, EndDay: 30}},
	}
	s, ok := gapCal.MustNewDate(2, 15, 1).Season()
	c.True(ok)
	c.Equal("Mid", s.Name)
	_, ok = gapCal.MustNewDate(1, 15, 1).Season()
	c.False(ok)
	_, ok = gapCal.MustNewDate(3, 15, 1).Season()
	c.False(ok)

	// A calendar whose two seasons overlap in the middle month: the first one in declaration order wins there.
	overlapCal := &calendar.Calendar{
		WeekDays: weekDays,
		Months:   months,
		Seasons: []calendar.Season{
			{Name: "First", StartMonth: 1, StartDay: 1, EndMonth: 2, EndDay: 30},
			{Name: "Second", StartMonth: 2, StartDay: 1, EndMonth: 3, EndDay: 30},
		},
	}
	s, ok = overlapCal.MustNewDate(2, 15, 1).Season()
	c.True(ok)
	c.Equal("First", s.Name)
	s, ok = overlapCal.MustNewDate(3, 15, 1).Season()
	c.True(ok)
	c.Equal("Second", s.Name)
}
