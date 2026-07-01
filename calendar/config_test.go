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
	"math"
	"testing"

	"github.com/richardwilkes/rpgtools/calendar"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestValidPrefabs(t *testing.T) {
	c := check.New(t)
	c.NoError(calendar.Gregorian().Config().Valid())
	c.NoError(calendar.PathfinderAbsalomReckoning().Config().Valid())
	c.NoError(calendar.PathfinderImperialCalendar().Config().Valid())
}

func TestConfigBoundsTotalDaysPerYear(t *testing.T) {
	c := check.New(t)

	base := func(months ...calendar.Month) *calendar.Config {
		return &calendar.Config{
			WeekDays:       []string{"A", "B"},
			DayZeroWeekDay: 0,
			Months:         months,
		}
	}

	// Regression: Config.Valid() previously placed no upper bound on the per-month Days or their sum, so a config whose
	// months summed past math.MaxInt wrapped MinDaysPerYear() negative and silently corrupted every date computation --
	// e.g. NewDate(1,1,5) returned Days=-8 with Year()==3 on a Valid() config. Two math.MaxInt-day months must now be
	// rejected outright, so no such Calendar can be built.
	huge := base(calendar.Month{Name: "A", Days: math.MaxInt}, calendar.Month{Name: "B", Days: math.MaxInt})
	c.HasError(huge.Valid())
	cal, err := calendar.New(huge)
	c.HasError(err)
	c.True(cal == nil)

	// The total is capped at math.MaxInt32. A single month exactly at the cap is accepted, MinDaysPerYear reports it
	// faithfully, and building the most extreme valid date stays finite (saturating to the day limit) rather than
	// wrapping to a negative day count or panicking in resolve().
	atCap := base(calendar.Month{Name: "A", Days: math.MaxInt32})
	c.NoError(atCap.Valid())
	cal, err = calendar.New(atCap)
	c.NoError(err)
	c.Equal(math.MaxInt32, cal.MinDaysPerYear())
	d, err := cal.NewDate(1, 1, math.MaxInt32)
	c.NoError(err)
	c.True(d.Days() > 0, "extreme date must not wrap to a negative day count")
	c.Equal(1, d.Month())

	// One day past the cap (spread across two months) is rejected, proving the bound is inclusive and not off by one.
	overByOne := base(calendar.Month{Name: "A", Days: math.MaxInt32}, calendar.Month{Name: "B", Days: 1})
	c.HasError(overByOne.Valid())

	// A month total well within the cap remains valid regardless of how many months contribute to it.
	ok := base(calendar.Month{Name: "A", Days: 30}, calendar.Month{Name: "B", Days: 31}, calendar.Month{Name: "C", Days: 30})
	c.NoError(ok.Valid())
}
