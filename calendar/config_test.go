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
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/richardwilkes/rpgtools/calendar"
	"github.com/richardwilkes/toolbox/v2/check"
	"gopkg.in/yaml.v3"
)

func TestValidPrefabs(t *testing.T) {
	c := check.New(t)
	c.NoError(calendar.Gregorian().Config().Valid())
	c.NoError(calendar.PathfinderAbsalomReckoning().Config().Valid())
	c.NoError(calendar.PathfinderImperialCalendar().Config().Valid())
}

func TestMinDaysPerYearMatchesMonthSum(t *testing.T) {
	c := check.New(t)
	// MinDaysPerYear is summed once when the Calendar is built and cached (it is a pure function of the immutable
	// Config). For every construction path -- the built-ins and New -- the cached value must equal the live sum of
	// every month's Days, so the cache can never drift from the Config it was built from.
	assertSum := func(cal *calendar.Calendar) {
		want := 0
		for _, m := range cal.Config().Months {
			want += m.Days
		}
		c.Equal(want, cal.MinDaysPerYear())
	}
	c.Equal(365, calendar.Gregorian().MinDaysPerYear())
	assertSum(calendar.Gregorian())
	assertSum(calendar.PathfinderAbsalomReckoning())
	assertSum(calendar.PathfinderImperialCalendar())

	custom, err := calendar.New(&calendar.Config{
		WeekDays:       []string{"A", "B"},
		DayZeroWeekDay: 0,
		Months: []calendar.Month{
			{Name: "M1", Days: 10},
			{Name: "M2", Days: 20},
			{Name: "M3", Days: 33},
		},
	})
	c.NoError(err)
	c.Equal(63, custom.MinDaysPerYear())
	assertSum(custom)
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

	// The total is capped at math.MaxInt32. A single month exactly at the cap is accepted and MinDaysPerYear reports
	// it faithfully. Year math.MaxInt32 on such a calendar lies beyond DaysLimit, so NewDate rejects it rather than
	// returning a saturated date, while a date at the limit itself stays finite and resolves rather than wrapping to a
	// negative day count or panicking in resolve().
	atCap := base(calendar.Month{Name: "A", Days: math.MaxInt32})
	c.NoError(atCap.Valid())
	cal, err = calendar.New(atCap)
	c.NoError(err)
	c.Equal(math.MaxInt32, cal.MinDaysPerYear())
	_, err = cal.NewDate(1, 1, math.MaxInt32)
	c.HasError(err)
	d := cal.NewDateByDays(calendar.DaysLimit)
	c.True(d.Days() > 0, "extreme date must not wrap to a negative day count")
	c.Equal(1, d.Month())

	// One day past the cap (spread across two months) is rejected, proving the bound is inclusive and not off by one.
	overByOne := base(calendar.Month{Name: "A", Days: math.MaxInt32}, calendar.Month{Name: "B", Days: 1})
	c.HasError(overByOne.Valid())

	// A month total well within the cap remains valid regardless of how many months contribute to it.
	ok := base(calendar.Month{Name: "A", Days: 30}, calendar.Month{Name: "B", Days: 31}, calendar.Month{Name: "C", Days: 30})
	c.NoError(ok.Valid())
}

func TestConfigOmitsEmptySeasons(t *testing.T) {
	c := check.New(t)
	// Seasons is optional, like LeapYear and the eras, and must be omitted from both encodings when empty rather than
	// appearing as an empty list in YAML while being absent from JSON.
	cfg := &calendar.Config{
		WeekDays: []string{"A", "B"},
		Months:   []calendar.Month{{Name: "M", Days: 10}},
	}
	c.NoError(cfg.Valid())
	out, err := yaml.Marshal(cfg)
	c.NoError(err)
	c.False(strings.Contains(string(out), "seasons"), "YAML: %s", out)
	out, err = json.Marshal(cfg)
	c.NoError(err)
	c.False(strings.Contains(string(out), "seasons"), "JSON: %s", out)

	cfg.Seasons = []calendar.Season{{Name: "Year", StartMonth: 1, StartDay: 1, EndMonth: 1, EndDay: 10}}
	out, err = yaml.Marshal(cfg)
	c.NoError(err)
	c.True(strings.Contains(string(out), "seasons:"), "YAML: %s", out)
	var back calendar.Config
	c.NoError(yaml.Unmarshal(out, &back))
	c.Equal(cfg.Seasons, back.Seasons)
	out, err = json.Marshal(cfg)
	c.NoError(err)
	c.True(strings.Contains(string(out), `"seasons"`), "JSON: %s", out)
	back = calendar.Config{}
	c.NoError(json.Unmarshal(out, &back))
	c.Equal(cfg.Seasons, back.Seasons)
}
