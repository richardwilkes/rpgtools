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

// TestConfigValidRejectsEachRule pins every rule Valid enforces, one violation at a time, starting from the Gregorian
// Config so that only the mutated field is at fault. Each error must name the rule that fired so a violation cannot be
// caught by an earlier, unrelated check by accident.
func TestConfigValidRejectsEachRule(t *testing.T) {
	c := check.New(t)
	const (
		badSeasonStartDay = "seasons must start in a valid day within the month"
		badSeasonEndDay   = "seasons must end in a valid day within the month"
	)
	var nilCfg *calendar.Config
	err := nilCfg.Valid()
	c.HasError(err)
	if err != nil {
		c.Contains(err.Error(), "nil")
	}
	for _, tc := range []struct {
		name   string
		mutate func(cfg *calendar.Config)
		want   string
	}{
		{"no week days", func(cfg *calendar.Config) { cfg.WeekDays = nil }, "at least one week day"},
		{"empty week day name", func(cfg *calendar.Config) { cfg.WeekDays[3] = "" }, "week day names must not be empty"},
		{
			"week day name with leading whitespace", func(cfg *calendar.Config) { cfg.WeekDays[0] = " Sunday" },
			"week day names may not begin or end with whitespace",
		},
		{
			"week day name with trailing whitespace", func(cfg *calendar.Config) { cfg.WeekDays[6] = "Saturday\t" },
			"week day names may not begin or end with whitespace",
		},
		{"DayZeroWeekDay below 0", func(cfg *calendar.Config) { cfg.DayZeroWeekDay = -1 }, "DayZeroWeekDay"},
		{"DayZeroWeekDay past the last week day", func(cfg *calendar.Config) { cfg.DayZeroWeekDay = 7 }, "DayZeroWeekDay"},
		{"no months", func(cfg *calendar.Config) { cfg.Months = nil }, "at least one month"},
		{"empty month name", func(cfg *calendar.Config) { cfg.Months[5].Name = "" }, "month names must not be empty"},
		{
			"month name with surrounding whitespace", func(cfg *calendar.Config) { cfg.Months[0].Name = " January " },
			"month names may not begin or end with whitespace",
		},
		{"month with no days", func(cfg *calendar.Config) { cfg.Months[1].Days = 0 }, "at least 1 day"},
		{"month with negative days", func(cfg *calendar.Config) { cfg.Months[1].Days = -30 }, "at least 1 day"},
		{"leap month below 1", func(cfg *calendar.Config) { cfg.LeapYear.Month = 0 }, "LeapYear.Month"},
		{"leap month past the last month", func(cfg *calendar.Config) { cfg.LeapYear.Month = 13 }, "LeapYear.Month"},
		{
			"Unless without Except", func(cfg *calendar.Config) { cfg.LeapYear.Except = 0 },
			"LeapYear.Unless may not be set if LeapYear.Except is 0",
		},
		{
			"Unless equal to Except", func(cfg *calendar.Config) { cfg.LeapYear.Unless = 100 },
			"LeapYear.Unless must be greater than LeapYear.Except",
		},
		{
			"Unless below Except", func(cfg *calendar.Config) { cfg.LeapYear.Unless = 8 },
			"LeapYear.Unless must be greater than LeapYear.Except",
		},
		{"empty season name", func(cfg *calendar.Config) { cfg.Seasons[2].Name = "" }, "season names must not be empty"},
		{
			"season name with surrounding whitespace", func(cfg *calendar.Config) { cfg.Seasons[0].Name = "Winter " },
			"season names may not begin or end with whitespace",
		},
		{
			"season start month below 1", func(cfg *calendar.Config) { cfg.Seasons[1].StartMonth = 0 },
			"seasons must start in a valid month",
		},
		{
			"season start month past the last month", func(cfg *calendar.Config) { cfg.Seasons[1].StartMonth = 13 },
			"seasons must start in a valid month",
		},
		{
			"season start day below 1", func(cfg *calendar.Config) { cfg.Seasons[1].StartDay = 0 },
			badSeasonStartDay,
		},
		{
			"season start day past the end of its month", func(cfg *calendar.Config) { cfg.Seasons[1].StartDay = 32 },
			badSeasonStartDay,
		},
		{"season start day past the leap day", func(cfg *calendar.Config) {
			cfg.Seasons[1].StartMonth, cfg.Seasons[1].StartDay = 2, 30
		}, badSeasonStartDay},
		{
			"season end month below 1", func(cfg *calendar.Config) { cfg.Seasons[3].EndMonth = 0 },
			"seasons must end in a valid month",
		},
		{
			"season end month past the last month", func(cfg *calendar.Config) { cfg.Seasons[3].EndMonth = 13 },
			"seasons must end in a valid month",
		},
		{
			"season end day below 1", func(cfg *calendar.Config) { cfg.Seasons[3].EndDay = 0 },
			badSeasonEndDay,
		},
		{
			"season end day past the end of its month", func(cfg *calendar.Config) { cfg.Seasons[3].EndDay = 32 },
			badSeasonEndDay,
		},
		{
			"season end day past the leap day", func(cfg *calendar.Config) { cfg.Seasons[0].EndDay = 30 },
			badSeasonEndDay,
		},
		{
			"era with surrounding whitespace", func(cfg *calendar.Config) { cfg.Era = " AD" },
			"era may not begin or end with whitespace",
		},
		{
			"previous era with surrounding whitespace", func(cfg *calendar.Config) { cfg.PreviousEra = "BC " },
			"previous era may not begin or end with whitespace",
		},
		{"era without previous era", func(cfg *calendar.Config) { cfg.PreviousEra = "" }, "both be set or neither"},
		{"previous era without era", func(cfg *calendar.Config) { cfg.Era = "" }, "both be set or neither"},
		{
			"eras differing only in case", func(cfg *calendar.Config) { cfg.Era, cfg.PreviousEra = "ad", "AD" },
			"differ only in letter case",
		},
		{"names past their total length budget", func(cfg *calendar.Config) {
			cfg.Months[0].Name = strings.Repeat("a", 65536)
		}, "may not total more than"},
	} {
		cfg := calendar.Gregorian().Config()
		tc.mutate(cfg)
		err = cfg.Valid()
		c.HasError(err, tc.name)
		if err != nil {
			c.Contains(err.Error(), tc.want, tc.name)
		}
		_, err = calendar.New(cfg)
		c.HasError(err, tc.name)
	}

	// The boundaries each rule permits are accepted: the last week day as day zero, season boundaries on the leap day
	// (the day is a valid day-of-month across all years, even though it exists only in leap years), and a single shared
	// era label.
	cfg := calendar.Gregorian().Config()
	cfg.DayZeroWeekDay = 6
	cfg.Seasons[0].EndDay = 29
	cfg.Seasons[1].StartMonth, cfg.Seasons[1].StartDay = 2, 29
	cfg.Era, cfg.PreviousEra = "AR", "AR"
	c.NoError(cfg.Valid())
}

// TestEraCaseIsSignificantToNeither pins the fix for eras that differed only in letter case. ParseDate matches an era
// suffix without regard to case, so with eras "ad" and "AD" a date formatted in the current era ("1/5/2017 ad") parsed
// back as year -2017 with no error. Such a pair is now rejected up front; a pair that differs beyond case, and a pair
// that is identical, both remain usable and round-trip through their own text in every case they are written in.
func TestEraCaseIsSignificantToNeither(t *testing.T) {
	c := check.New(t)
	cfg := calendar.Gregorian().Config()
	cfg.Era, cfg.PreviousEra = "ad", "AD"
	c.HasError(cfg.Valid())
	cal, err := calendar.New(cfg)
	c.HasError(err)
	c.True(cal == nil)

	for _, eras := range [][2]string{{"AD", "BC"}, {"ad", "bc"}, {"AR", "AR"}} {
		cfg.Era, cfg.PreviousEra = eras[0], eras[1]
		cal, err = calendar.New(cfg)
		c.NoError(err, "eras %q/%q", eras[0], eras[1])
		for _, year := range []int{2017, -2017} {
			want := cal.MustNewDate(1, 5, year)
			formatted := want.Format("%N/%D/%y")
			for _, text := range []string{formatted, strings.ToUpper(formatted), strings.ToLower(formatted)} {
				var got calendar.Date
				got, err = cal.ParseDate(text)
				c.NoError(err, "eras %q/%q: %q", eras[0], eras[1], text)
				c.True(want.Equal(got), "eras %q/%q: %q parsed back as %s", eras[0], eras[1], text, got)
			}
		}
	}
}

// TestConfigBoundsNamesLength pins the budget on the total length of a Config's names, which exists to keep the
// ParseDate patterns (built from the month and era names as literals) inside the size a regular expression may have.
// A Config that spends the whole budget is accepted, builds, and parses its own text; one character more, wherever it
// is spent, is rejected. New once reached regexp.MustCompile with an unbounded name and panicked with "expression too
// large" -- reachable from any YAML- or JSON-loaded Config -- so the original 16 MB name must now be turned away by
// Valid, quickly and without a panic.
func TestConfigBoundsNamesLength(t *testing.T) {
	c := check.New(t)
	const maxNames = 65536 // the documented cap on the total length of a Config's names, in characters
	const budgeted = "the names of the week days, months, seasons and eras may not total more than"
	// The names below total exactly the cap: 4 + 3 + 2 + 1 characters spread over the other kinds so each is shown to
	// count, a five-character second month, and the rest on the first month's name, which the patterns embed in full.
	// The dates parsed lie in the short month: the patterns must work with the huge name among their alternatives,
	// but matching text that is itself tens of thousands of identical characters is quadratic in Go's regexp engine,
	// which is a property of the engine rather than of this budget.
	build := func(weekDay, month, season, era, previousEra string) *calendar.Config {
		return &calendar.Config{
			WeekDays:    []string{weekDay},
			Months:      []calendar.Month{{Name: month, Days: 30}, {Name: "Brief", Days: 30}},
			Seasons:     []calendar.Season{{Name: season, StartMonth: 1, StartDay: 1, EndMonth: 1, EndDay: 30}},
			Era:         era,
			PreviousEra: previousEra,
		}
	}
	month := strings.Repeat("m", maxNames-15)
	cfg := build("wwww", month, "sss", "ee", "p")
	c.NoError(cfg.Valid())
	cal, err := calendar.New(cfg)
	c.NoError(err)
	want := cal.MustNewDate(2, 5, 100)
	for _, layout := range []string{calendar.LongFormat, calendar.MediumFormat, "%M %D, %y", "%N/%D/%y"} {
		text := want.Format(layout)
		var got calendar.Date
		got, err = cal.ParseDate(text)
		c.NoError(err, "layout %q", layout)
		c.True(want.Equal(got), "layout %q", layout)
	}
	got, err := cal.ParseDate("2/5/100 p")
	c.NoError(err)
	c.True(cal.MustNewDate(2, 5, -100).Equal(got))

	// The budget is measured in characters, not bytes: the same count of two-byte characters is accepted.
	cfg = build("wwww", strings.Repeat("\u00e9", maxNames-15), "sss", "ee", "p")
	c.NoError(cfg.Valid())
	_, err = calendar.New(cfg)
	c.NoError(err)

	// One character over, spent on any kind of name, is rejected.
	for name, over := range map[string]*calendar.Config{
		"week day":     build("wwwww", month, "sss", "ee", "p"),
		"month":        build("wwww", month+"m", "sss", "ee", "p"),
		"season":       build("wwww", month, "ssss", "ee", "p"),
		"era":          build("wwww", month, "sss", "eee", "p"),
		"previous era": build("wwww", month, "sss", "ee", "pp"),
	} {
		err = over.Valid()
		c.HasError(err, name)
		if err != nil {
			c.Contains(err.Error(), budgeted, name)
		}
		_, err = calendar.New(over)
		c.HasError(err, name)
	}

	// The original panic: a single name past what a regular expression can hold.
	cfg = build("w", strings.Repeat("m", 1<<24+1), "s", "e", "p")
	c.NotPanics(func() { cal, err = calendar.New(cfg) })
	c.HasError(err)
	c.True(cal == nil)
	if err != nil {
		c.Contains(err.Error(), budgeted)
	}
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
