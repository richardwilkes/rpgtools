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
	"encoding/json"
	"fmt"
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

	// The longest year is capped at 2^30 days. A single month exactly at the cap is accepted and MinDaysPerYear reports
	// it faithfully, and even year math.MaxInt32 on such a calendar is representable: its last day stays finite and
	// resolves rather than wrapping to a negative day count or panicking in resolve().
	atCap := base(calendar.Month{Name: "A", Days: 1 << 30})
	c.NoError(atCap.Valid())
	cal, err = calendar.New(atCap)
	c.NoError(err)
	c.Equal(1<<30, cal.MinDaysPerYear())
	d, err := cal.NewDate(1, 1<<30, math.MaxInt32)
	c.NoError(err)
	c.True(d.Days() > 0, "extreme date must not wrap to a negative day count")
	c.Equal(1, d.Month())
	c.Equal(math.MaxInt32, d.Year())
	c.Equal(d, cal.NewDateByDays(math.MaxInt64))

	// One day past the cap (spread across two months) is rejected, proving the bound is inclusive and not off by one.
	overByOne := base(calendar.Month{Name: "A", Days: 1 << 30}, calendar.Month{Name: "B", Days: 1})
	c.HasError(overByOne.Valid())

	// A leap rule adds a day to the longest year, so a leap calendar must leave room for it under the cap.
	leapAtCap := base(calendar.Month{Name: "A", Days: 1 << 30})
	leapAtCap.LeapYear = &calendar.LeapYear{Month: 1, Every: 2}
	c.HasError(leapAtCap.Valid())
	leapUnderCap := base(calendar.Month{Name: "A", Days: 1<<30 - 1})
	leapUnderCap.LeapYear = &calendar.LeapYear{Month: 1, Every: 2}
	c.NoError(leapUnderCap.Valid())
	cal, err = calendar.New(leapUnderCap)
	c.NoError(err)
	c.Equal(1<<30-1, cal.Days(1))
	c.Equal(1<<30, cal.Days(2))

	// A month total well within the cap remains valid regardless of how many months contribute to it.
	ok := base(calendar.Month{Name: "A", Days: 30}, calendar.Month{Name: "B", Days: 31}, calendar.Month{Name: "C", Days: 30})
	c.NoError(ok.Valid())
}

// TestConfigBoundsListLengths pins the caps on the number of week days, months and seasons: each list is accepted at
// its cap and rejected one past it, independently of the others, and a calendar at every cap at once is fully usable.
func TestConfigBoundsListLengths(t *testing.T) {
	c := check.New(t)
	names := func(prefix string, n int) []string {
		out := make([]string, n)
		for i := range out {
			out[i] = fmt.Sprintf("%s%d", prefix, i+1)
		}
		return out
	}
	build := func(weekDays, months, seasons int) *calendar.Config {
		cfg := &calendar.Config{WeekDays: names("W", weekDays)}
		for _, name := range names("M", months) {
			cfg.Months = append(cfg.Months, calendar.Month{Name: name, Days: 1})
		}
		for _, name := range names("S", seasons) {
			cfg.Seasons = append(cfg.Seasons, calendar.Season{Name: name, StartMonth: 1, StartDay: 1, EndMonth: 1, EndDay: 1})
		}
		return cfg
	}
	c.NoError(build(99, 99, 99).Valid())
	c.HasError(build(100, 1, 0).Valid())
	c.HasError(build(1, 100, 0).Valid())
	c.HasError(build(1, 1, 100).Valid())

	cal, err := calendar.New(build(99, 99, 99))
	c.NoError(err)
	d := cal.MustNewDate(99, 1, 1)
	c.Equal(99, d.Month())
	c.Equal("M99", d.MonthName())
	c.Equal("99/1", d.Format("%n/%d")) // the %n padding never needs more than two digits
	c.Equal(98, d.WeekDay())
	c.Equal("W99", d.WeekDayName())
	s, ok := cal.MustNewDate(1, 1, 1).Season()
	c.True(ok)
	c.Equal("S1", s.Name)
	var buf bytes.Buffer
	c.NoError(cal.Text(1, &buf))
	c.True(strings.Contains(buf.String(), "99: M99"), "%s", buf.String())
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
