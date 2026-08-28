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
	"math"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/richardwilkes/rpgtools/calendar"
	"github.com/richardwilkes/toolbox/v2/check"
	"gopkg.in/yaml.v3"
)

func TestNewDate(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	d, err := cal.NewDate(1, 1, 1)
	c.NoError(err)
	c.Equal(cal.NewDateByDays(0), d)
	d, err = cal.NewDate(12, 31, 1)
	c.NoError(err)
	c.Equal(cal.NewDateByDays(364), d)
	d, err = cal.NewDate(1, 1, 2)
	c.NoError(err)
	c.Equal(cal.NewDateByDays(365), d)

	d, err = cal.NewDate(1, 1, -1)
	c.NoError(err)
	c.Equal(cal.NewDateByDays(-366), d)
	d, err = cal.NewDate(12, 31, -1)
	c.NoError(err)
	c.Equal(cal.NewDateByDays(-1), d)
	d, err = cal.NewDate(1, 1, -2)
	c.NoError(err)
	c.Equal(cal.NewDateByDays(-731), d)
	d, err = cal.NewDate(12, 31, -2)
	c.NoError(err)
	c.Equal(cal.NewDateByDays(-367), d)
	d, err = cal.NewDate(12, 31, -3)
	c.NoError(err)
	c.Equal(cal.NewDateByDays(-732), d)

	_, err = cal.NewDate(1, 1, 0)
	c.HasError(err)
	_, err = cal.NewDate(13, 22, 2017)
	c.HasError(err)
	_, err = cal.NewDate(9, 888, 2017)
	c.HasError(err)
}

// TestNewDateAcceptsEveryInt32Year verifies that every year IsValidYear accepts is representable in full on the
// calendars with the longest years Config.Valid allows, and pins where their extreme days fall: maxDaysPerYear is
// chosen so that even these stay within 2^61 of 1/1/1. NewDate once rejected years that ran past a fixed day limit on
// such calendars, so a valid int32 year was not always a valid date.
func TestNewDateAcceptsEveryInt32Year(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		leap        bool
		first, last int64
	}{
		{false, -1 << 61, 1<<61 - 1<<30 - 1},
		{true, -1<<61 + 1<<30, 1<<61 - 1<<31 - 1},
	} {
		cal := maximalCalendar(c, tc.leap)
		first, err := cal.NewDate(1, 1, math.MinInt32)
		c.NoError(err, "leap=%t", tc.leap)
		c.Equal(tc.first, first.Days(), "leap=%t", tc.leap)
		c.Equal(math.MinInt32, first.Year(), "leap=%t", tc.leap)
		c.Equal(first, cal.NewDateByDays(math.MinInt64), "leap=%t", tc.leap)
		last, err := cal.NewDate(1, cal.Days(math.MaxInt32), math.MaxInt32)
		c.NoError(err, "leap=%t", tc.leap)
		c.Equal(tc.last, last.Days(), "leap=%t", tc.leap)
		c.Equal(math.MaxInt32, last.Year(), "leap=%t", tc.leap)
		c.Equal(cal.Days(math.MaxInt32), last.DayInMonth(), "leap=%t", tc.leap)
		c.Equal(last, cal.NewDateByDays(math.MaxInt64), "leap=%t", tc.leap)
		// The day past either end does not exist: stepping there saturates.
		c.Equal(first, first.Add(-1), "leap=%t", tc.leap)
		c.Equal(last, last.Add(1), "leap=%t", tc.leap)
	}
	// Ordinary calendars are nowhere near these extremes; the int32 limits remain accepted.
	for _, ordinary := range []*calendar.Calendar{calendar.Gregorian(), calendar.PathfinderAbsalomReckoning()} {
		_, err := ordinary.NewDate(1, 1, math.MaxInt32)
		c.NoError(err)
		_, err = ordinary.NewDate(1, 1, math.MinInt32)
		c.NoError(err)
	}
}

func TestYear(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	c.Equal(1, cal.NewDateByDays(0).Year(), "First day of year 1")
	c.Equal(1, cal.NewDateByDays(364).Year(), "Last day of year 1")
	c.Equal(2, cal.NewDateByDays(365).Year(), "First day of year 2")

	c.Equal(-1, cal.NewDateByDays(-366).Year(), "First day of year -1")
	c.Equal(-1, cal.NewDateByDays(-1).Year(), "Last day of year -1")
	c.Equal(-2, cal.NewDateByDays(-731).Year(), "First day of year -2")
	c.Equal(-2, cal.NewDateByDays(-367).Year(), "Last day of year -2")
	c.Equal(-3, cal.NewDateByDays(-732).Year(), "Last day of year -3")

	for year := 1; year < 5000; year++ {
		c.Equal(year, cal.MustNewDate(1, 1, year).Year())
		c.Equal(year, cal.MustNewDate(12, 31, year).Year())
	}

	for year := -1; year > -5000; year-- {
		c.Equal(year, cal.MustNewDate(1, 1, year).Year())
		c.Equal(year, cal.MustNewDate(12, 31, year).Year())
	}
}

// TestYearLargeDays guards against a regression to the old O(date.Days) convergence loop in Year(), which made a Date
// carrying a very large (but legal) Days value effectively hang. The binary search returns in O(log) steps, so these
// calls complete near-instantly; under the old loop the 1<<62 cases alone would have run for trillions of iterations.
// Each result is cross-checked against the calendar's own first-day-of-year boundaries: the reported year must start on
// or before the date and the following year must start strictly after it.
func TestYearLargeDays(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	for _, days := range []int64{math.MaxInt32, math.MaxInt32 - 2, math.MinInt32, math.MinInt32 + 2} {
		year := cal.NewDateByDays(days).Year()
		c.True(year != 0, "year is never 0 (days=%d)", days)
		c.True(cal.MustNewDate(1, 1, year).Days() <= days, "1/1/%d must start on or before days=%d", year, days)
		next := year + 1
		if next == 0 { // skip the nonexistent year 0 when stepping forward from year -1
			next = 1
		}
		c.True(days < cal.MustNewDate(1, 1, next).Days(), "1/1/%d must start after days=%d", next, days)
	}
}

// TestYearSaturatesAtInt32Bounds guards the overflow-safe search bounds in Year() and the confinement of every Date to
// the years IsValidYear accepts. A Date.Days value within a year of the int64 limits once drove the bound arithmetic
// (and the yearToDaysWith probe at that bound) to overflow, so Year() returned a wildly wrong, even sign-flipped,
// result. Date.Add (and so NewDateByDays) now saturates at the last day of year math.MaxInt32 and the first day of
// year math.MinInt32, far below the int64 limits even on a calendar whose year is a single day, and the search bounds
// are kept as tight as correctness allows, so the search never probes a year whose day count overflows and never
// settles on a year outside the int32 range. These cases use a minimal two-day, no-leap calendar (minDays == 2): year y
// then spans days (y-1)*2 .. y*2-1 for y > 0 and y*2 .. (y+1)*2-1 for y < 0, which pins the expected day count at each
// extreme.
func TestYearSaturatesAtInt32Bounds(t *testing.T) {
	c := check.New(t)
	cal, err := calendar.New(&calendar.Config{
		WeekDays: []string{"A", "B"},
		Months:   []calendar.Month{{Name: "First", Days: 1}, {Name: "Second", Days: 1}},
	})
	c.NoError(err)
	last := cal.NewDateByDays(math.MaxInt64)
	c.Equal(int64(math.MaxInt32)*2-1, last.Days())
	c.Equal(math.MaxInt32, last.Year())
	c.Equal(last, cal.MustNewDate(2, 1, math.MaxInt32))
	first := cal.NewDateByDays(math.MinInt64)
	c.Equal(int64(math.MinInt32)*2, first.Days())
	c.Equal(math.MinInt32, first.Year())
	c.Equal(first, cal.MustNewDate(1, 1, math.MinInt32))
	// Year must stay non-decreasing in Days right up to the limits, where the off-by-one used to appear.
	for _, base := range []int64{last.Days() - 5, first.Days()} {
		for off := int64(0); off <= 5; off++ {
			earlier := cal.NewDateByDays(base + off).Year()
			later := cal.NewDateByDays(base + off + 1).Year()
			c.True(earlier <= later, "Year must be non-decreasing in Days near %d (off=%d): %d > %d",
				base, off, earlier, later)
		}
	}
}

// TestYearNearInt32Boundary guards against the leap-counting year search overshooting for valid years near the int32
// limit. Date.Year's binary search upper bound (date.days/minDays+1) runs past math.MaxInt32 for a date whose year sits
// near the top of the valid range, so the search probes years just beyond the limit. The internal leap math must stay
// correct for those probe years or the search loses monotonicity and settles past the limit; the regression resolved
// Gregorian 1/1/MaxInt32 to year 2148910399 / month 11 instead of MaxInt32 / month 1. Exercise both int32 extremes on
// calendars that carry a leap rule (a no-leap calendar never hits the leap math, so it is unaffected).
func TestYearNearInt32Boundary(t *testing.T) {
	c := check.New(t)
	for _, cal := range []*calendar.Calendar{calendar.Gregorian(), calendar.PathfinderAbsalomReckoning()} {
		for _, year := range []int{
			math.MaxInt32, math.MaxInt32 - 1, math.MaxInt32 - 2, math.MaxInt32 - 100,
			math.MinInt32, math.MinInt32 + 1, math.MinInt32 + 2, math.MinInt32 + 100,
		} {
			first := cal.MustNewDate(1, 1, year)
			c.Equal(year, first.Year(), "Year() of 1/1/%d", year)
			c.Equal(1, first.Month(), "Month() of 1/1/%d", year)
			c.Equal(1, first.DayInMonth(), "DayInMonth() of 1/1/%d", year)
			// A later day in the same year must resolve back to the same year and a day count 14 days on.
			mid := cal.MustNewDate(1, 15, year)
			c.Equal(year, mid.Year(), "Year() of 1/15/%d", year)
			c.Equal(first.Days()+14, mid.Days(), "Days() of 1/15/%d", year)
		}
	}
}

func TestDayInYear(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	c.Equal(1, cal.MustNewDate(1, 1, 1).DayInYear())
	c.Equal(365, cal.MustNewDate(12, 31, 1).DayInYear())
	c.Equal(1, cal.MustNewDate(1, 1, 2).DayInYear())
	c.Equal(1, cal.MustNewDate(1, 1, 4).DayInYear())
	c.Equal(366, cal.MustNewDate(12, 31, 4).DayInYear())

	c.Equal(1, cal.MustNewDate(1, 1, -1).DayInYear())
	c.Equal(366, cal.MustNewDate(12, 31, -1).DayInYear())
	c.Equal(1, cal.MustNewDate(1, 1, -2).DayInYear())
	c.Equal(365, cal.MustNewDate(12, 31, -2).DayInYear())
	c.Equal(1, cal.MustNewDate(1, 1, -5).DayInYear())
	c.Equal(366, cal.MustNewDate(12, 31, -5).DayInYear())
}

func TestMonth(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	c.Equal(1, cal.MustNewDate(1, 1, 1).Month())
	c.Equal(1, cal.MustNewDate(1, 31, 1).Month())
	c.Equal(2, cal.MustNewDate(2, 1, 1).Month())
	c.Equal(2, cal.MustNewDate(2, 28, 1).Month())
	c.Equal(3, cal.MustNewDate(3, 1, 1).Month())
	c.Equal(12, cal.MustNewDate(12, 31, 1).Month())
	c.Equal(1, cal.MustNewDate(1, 1, 2).Month())
	c.Equal(2, cal.MustNewDate(2, 28, 4).Month())
	c.Equal(2, cal.MustNewDate(2, 29, 4).Month())
	c.Equal(3, cal.MustNewDate(3, 1, 4).Month())

	c.Equal(2, cal.MustNewDate(2, 29, -1).Month())
	c.Equal(2, cal.MustNewDate(2, 28, -2).Month())
}

func TestDayInMonth(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	c.Equal(1, cal.MustNewDate(1, 1, 1).DayInMonth())
	c.Equal(31, cal.MustNewDate(1, 31, 1).DayInMonth())
	c.Equal(1, cal.MustNewDate(2, 1, 1).DayInMonth())
	c.Equal(28, cal.MustNewDate(2, 28, 1).DayInMonth())
	c.Equal(1, cal.MustNewDate(3, 1, 1).DayInMonth())
	c.Equal(31, cal.MustNewDate(12, 31, 1).DayInMonth())
	c.Equal(1, cal.MustNewDate(1, 1, 2).DayInMonth())
	c.Equal(28, cal.MustNewDate(2, 28, 2).DayInMonth())
	c.Equal(1, cal.MustNewDate(3, 1, 2).DayInMonth())
	c.Equal(28, cal.MustNewDate(2, 28, 4).DayInMonth())
	c.Equal(29, cal.MustNewDate(2, 29, 4).DayInMonth())
	c.Equal(1, cal.MustNewDate(3, 1, 4).DayInMonth())

	c.Equal(29, cal.MustNewDate(2, 29, -1).DayInMonth())
	c.Equal(28, cal.MustNewDate(2, 28, -2).DayInMonth())
}

func TestDateToString(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	c.Equal("1/1/1", cal.MustNewDate(1, 1, 1).String())
	c.Equal("12/31/1", cal.MustNewDate(12, 31, 1).String())
	c.Equal("1/1/2", cal.MustNewDate(1, 1, 2).String())
	c.Equal("1/1/2017", cal.MustNewDate(1, 1, 2017).String())
	c.Equal("9/22/2017", cal.MustNewDate(9, 22, 2017).String())

	c.Equal("1/1/1 BC", cal.MustNewDate(1, 1, -1).String())
	c.Equal("12/31/1 BC", cal.MustNewDate(12, 31, -1).String())
	c.Equal("1/1/2 BC", cal.MustNewDate(1, 1, -2).String())
	c.Equal("12/31/2 BC", cal.MustNewDate(12, 31, -2).String())
	c.Equal("12/31/3 BC", cal.MustNewDate(12, 31, -3).String())
}

func TestWeekDay(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	c.Equal(1, cal.MustNewDate(1, 1, 1).WeekDay())
	c.Equal(4, cal.MustNewDate(1, 4, 1).WeekDay())
	c.Equal(1, cal.MustNewDate(1, 8, 1).WeekDay())
	c.Equal(0, cal.MustNewDate(12, 31, -1).WeekDay())
	c.Equal(6, cal.MustNewDate(12, 30, -1).WeekDay())
	c.Equal(0, cal.MustNewDate(12, 24, -1).WeekDay())
	c.Equal(6, cal.MustNewDate(1, 1, 2000).WeekDay())
	c.Equal(1, cal.MustNewDate(9, 3, 2018).WeekDay())
}

func TestFormat(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	d := cal.MustNewDate(9, 22, 2017)
	c.Equal("9/22/2017", d.Format(calendar.ShortFormat))
	c.Equal("Sep 22, 2017", d.Format(calendar.MediumFormat))
	c.Equal("September 22, 2017", d.Format(calendar.LongFormat))
	c.Equal("Friday, September 22, 2017", d.Format(calendar.FullFormat))
	c.Equal("%Fri%", d.Format("%%%w%%"))
	c.Equal("Friday, September 22, 2017 AD", d.Format("%W, %M %D, %y"))

	d = cal.MustNewDate(9, 22, -1)
	c.Equal("9/22/1 BC", d.Format(calendar.ShortFormat))
	c.Equal("Sep 22, 1 BC", d.Format(calendar.MediumFormat))
	c.Equal("September 22, 1 BC", d.Format(calendar.LongFormat))
	c.Equal("Friday, September 22, 1 BC", d.Format(calendar.FullFormat))
	c.Equal("%Fri%", d.Format("%%%w%%"))
	c.Equal("Friday, September 22, 1 BC", d.Format("%W, %M %D, %y"))
}

// TestFormatRepeatedDirectivesConsistent verifies that resolving the date once for the whole layout (instead of once
// per directive) does not let one directive's transformation leak into another. In particular %y flips the sign of a
// negative year for a distinct-era display, which must not disturb the signed year a later %z or %Y reads from the same
// resolved value, and a directive that appears several times must yield the same value each time.
func TestFormatRepeatedDirectivesConsistent(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian() // distinct eras: AD / BC

	d := cal.MustNewDate(9, 22, -1) // year -1
	// %z is the raw signed year; %y and %Y render the previous era as "1 BC". Each %z must still report -1 even though
	// it follows a %y that negates its own copy of the year for display.
	c.Equal("-1|1 BC|-1|1 BC|-1", d.Format("%z|%y|%z|%Y|%z"))
	// Month and day directives repeated across the layout must each resolve to the same value.
	c.Equal("September 22 September 22 9 22", d.Format("%M %D %M %D %N %D"))

	d = cal.MustNewDate(9, 22, 2017) // positive year keeps its sign through %y
	c.Equal("2017|2017 AD|2017", d.Format("%z|%y|%z"))
}

// TestMediumFormatRoundTripsThroughParse verifies the abbreviated month name that %m emits is exactly the abbreviation
// monthFromText accepts when parsing, so MediumFormat output parses back to the same date for every month. Both the
// emit and parse sides are driven by abbreviatedNameLength, so this round-trip holds by construction; the test guards
// against the two widths drifting apart again.
func TestMediumFormatRoundTripsThroughParse(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	for month := 1; month <= 12; month++ {
		want := cal.MustNewDate(month, 15, 2017)
		formatted := want.Format(calendar.MediumFormat) // uses %m, the abbreviated month name
		got, err := cal.ParseDate(formatted)
		c.NoError(err, "month %d formatted as %q", month, formatted)
		c.Equal(want, got, "month %d round-trip via %q", month, formatted)
	}
}

// Era labels that Valid accepts but that a single alphabetic word cannot match, used by the parse tests below.
const (
	dottedEra     = "A.D."
	dottedPrevEra = "B.C."
	wordyEra      = "Age of Light"
	wordyPrevEra  = "Age of Dark"
)

// maximalCalendar builds a calendar with a single month holding the most days Config.Valid accepts, with or without a
// leap rule (every second year, the densest allowed). Its years are the longest possible, so its first day of year
// math.MinInt32 and last day of year math.MaxInt32 are the farthest any Date can lie from 1/1/1.
func maximalCalendar(c check.Checker, leap bool) *calendar.Calendar {
	cfg := &calendar.Config{
		WeekDays: []string{"A"},
		Months:   []calendar.Month{{Name: "Only", Days: 1 << 30}},
	}
	if leap {
		cfg.Months[0].Days--
		cfg.LeapYear = &calendar.LeapYear{Month: 1, Every: 2}
	}
	cal, err := calendar.New(cfg)
	c.NoError(err)
	return cal
}

// eraTestCalendar builds a small but valid calendar whose only interesting variation is its era pair.
func eraTestCalendar(era, previousEra string) (*calendar.Calendar, error) {
	return calendar.New(&calendar.Config{
		DayZeroWeekDay: 0,
		WeekDays:       []string{"A", "B", "C", "D", "E", "F", "G"},
		Months: []calendar.Month{
			{Name: "Janus", Days: 30},
			{Name: "Febris", Days: 30},
			{Name: "Martis", Days: 30},
		},
		Seasons:     []calendar.Season{{Name: "S", StartMonth: 1, StartDay: 1, EndMonth: 3, EndDay: 30}},
		Era:         era,
		PreviousEra: previousEra,
	})
}

// TestEraDisplayModel pins the single era model (eraForYear) that %z, %Y, %y and Date.Era all build on, across the
// three era configurations: distinct eras (the label carries the sign), a single shared label (the sign stays on the
// number), and no eras at all. The Gregorian-only TestFormat does not cover the shared-label or empty configurations.
func TestEraDisplayModel(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct { //nolint:govet // Not concerned with pointer bytes in tests
		era, prev                    string
		year                         int
		wantZ, wantY, wanty, wantEra string
	}{
		// Distinct eras: %Y is terse (no label on a non-negative year), %y always labels, and a negative year shows its
		// magnitude beside the previous-era label.
		{"AD", "BC", 2017, "2017", "2017", "2017 AD", "AD"},
		{"AD", "BC", -5, "-5", "5 BC", "5 BC", "BC"},
		// A single shared era label: %Y and %y agree, and the sign stays on the number rather than the label.
		{"AR", "AR", 2017, "2017", "2017 AR", "2017 AR", "AR"},
		{"AR", "AR", -5, "-5", "-5 AR", "-5 AR", "AR"},
		// No eras: nothing but the signed year, whichever directive is used.
		{"", "", 2017, "2017", "2017", "2017", ""},
		{"", "", -5, "-5", "-5", "-5", ""},
	} {
		cal, err := eraTestCalendar(tc.era, tc.prev)
		c.NoError(err)
		d := cal.MustNewDate(2, 15, tc.year)
		c.Equal(tc.wantZ, d.Format("%z"), "directive z, eras %q/%q, year %d", tc.era, tc.prev, tc.year)
		c.Equal(tc.wantY, d.Format("%Y"), "directive Y, eras %q/%q, year %d", tc.era, tc.prev, tc.year)
		c.Equal(tc.wanty, d.Format("%y"), "directive y, eras %q/%q, year %d", tc.era, tc.prev, tc.year)
		c.Equal(tc.wantEra, d.Era(), "Era(), eras %q/%q, year %d", tc.era, tc.prev, tc.year)
	}
}

// TestEraRoundTripsThroughParse verifies eraForYear and resolveEraSuffix are exact inverses: a date rendered with the
// era-bearing %y directive parses back to the same date for every era configuration and sign, so the format and parse
// sides of the era model cannot drift apart.
func TestEraRoundTripsThroughParse(t *testing.T) {
	c := check.New(t)
	// The last two pairs carry punctuation and spaces, which Valid permits and which the parse patterns must therefore
	// match in full rather than as a generic single word.
	for _, eras := range [][2]string{{"AD", "BC"}, {"AR", "AR"}, {"", ""}, {dottedEra, dottedPrevEra}, {wordyEra, wordyPrevEra}} {
		cal, err := eraTestCalendar(eras[0], eras[1])
		c.NoError(err)
		for _, year := range []int{2017, 5, 1, -1, -5, -2017} {
			want := cal.MustNewDate(2, 15, year)
			formatted := want.Format("%M %D, %y")
			var got calendar.Date
			got, err = cal.ParseDate(formatted)
			c.NoError(err, "eras %q/%q year %d formatted as %q", eras[0], eras[1], year, formatted)
			c.Equal(want, got, "eras %q/%q: %q did not round-trip", eras[0], eras[1], formatted)
		}
	}
}

func TestFormatZeroPadded(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	c.Equal("01/01", cal.MustNewDate(1, 1, 2017).Format("%n/%d"))
	c.Equal("09/22", cal.MustNewDate(9, 22, 2017).Format("%n/%d"))
	c.Equal("02/29", cal.MustNewDate(2, 29, 2016).Format("%n/%d"))
	// December is the last month; %d used to index past the end of the months
	// slice and panic.
	c.Equal("12/05", cal.MustNewDate(12, 5, 2017).Format("%n/%d"))
	c.Equal("12/31", cal.MustNewDate(12, 31, 2017).Format("%n/%d"))
}

func TestFormatDayWidthConsistent(t *testing.T) {
	c := check.New(t)

	// A calendar whose months have very different lengths. The zero-padded day (%d) width must be consistent across
	// every month, sized to the calendar's longest month rather than the month being formatted.
	cal, err := calendar.New(&calendar.Config{
		WeekDays:       []string{"A", "B", "C"},
		DayZeroWeekDay: 0,
		Months:         []calendar.Month{{Name: "Short", Days: 5}, {Name: "Long", Days: 40}},
		Seasons:        []calendar.Season{{Name: "All", StartMonth: 1, StartDay: 1, EndMonth: 2, EndDay: 40}},
	})
	c.NoError(err)
	// Day 3 of the short month previously rendered as "3" (width 1) while the long month rendered "03" (width 2); both
	// must now be "03".
	c.Equal("03", cal.MustNewDate(1, 3, 1).Format("%d"))
	c.Equal("03", cal.MustNewDate(2, 3, 1).Format("%d"))
	c.Equal("40", cal.MustNewDate(2, 40, 1).Format("%d"))

	// The leap month's extra day is accounted for, so the width is also consistent between leap and non-leap years.
	var leapCal *calendar.Calendar
	leapCal, err = calendar.New(&calendar.Config{
		WeekDays:       []string{"A", "B"},
		DayZeroWeekDay: 0,
		Months:         []calendar.Month{{Name: "M", Days: 9}},
		Seasons:        []calendar.Season{{Name: "All", StartMonth: 1, StartDay: 1, EndMonth: 1, EndDay: 9}},
		LeapYear:       &calendar.LeapYear{Month: 1, Every: 2},
	})
	c.NoError(err)
	c.Equal("03", leapCal.MustNewDate(1, 3, 1).Format("%d"))  // year 1: 9-day month
	c.Equal("03", leapCal.MustNewDate(1, 3, 2).Format("%d"))  // year 2 (leap): 10-day month
	c.Equal("10", leapCal.MustNewDate(1, 10, 2).Format("%d")) // the leap day itself
}

func TestTextMultiByteWeekDayNames(t *testing.T) {
	c := check.New(t)
	cfg := calendar.Gregorian().Config()
	// Week day names whose first rune is multi-byte in UTF-8.
	cfg.WeekDays = []string{"Étoile", "Понедельник", "Δευτέρα", "三", "Mercredi", "木曜日", "Saturn"}
	cal, err := calendar.New(cfg)
	c.NoError(err)

	var buf bytes.Buffer
	c.NoError(cal.Text(2017, &buf))
	out := buf.String()
	c.True(utf8.ValidString(out), "calendar text must remain valid UTF-8")
	// The week day legend abbreviates each name to its first rune, not its first byte.
	c.True(strings.Contains(out, "(É)"), "expected first-rune abbreviation '(É)' in:\n%s", out)
	c.True(strings.Contains(out, "(三)"), "expected first-rune abbreviation '(三)' in:\n%s", out)

	buf.Reset()
	cal.MustNewDate(1, 1, 2017).TextCalendarMonth(&buf)
	c.True(utf8.ValidString(buf.String()), "month text must remain valid UTF-8")
}

func TestTextEmptyWeekDayNameDoesNotPanic(t *testing.T) {
	c := check.New(t)
	cfg := calendar.Gregorian().Config()
	// An empty week day name previously caused weekday[:1] to panic.
	cfg.WeekDays = []string{"", "B", "C", "D", "E", "F", "G"}
	_, err := calendar.New(cfg)
	c.HasError(err)
}

func TestParseDate(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	targetDate := cal.MustNewDate(9, 22, 2017)
	date, err := cal.ParseDate("A long, rambling prefix September 22, 2017 and a long suffix")
	c.NoError(err)
	c.Equal(targetDate, date)
	date, err = cal.ParseDate("Friday, September 22, 2017")
	c.NoError(err)
	c.Equal(targetDate, date)
	date, err = cal.ParseDate("September 22, 2017")
	c.NoError(err)
	c.Equal(targetDate, date)
	date, err = cal.ParseDate("9/22/2017")
	c.NoError(err)
	c.Equal(targetDate, date)
	date, err = cal.ParseDate("what 9/22/2017 how?")
	c.NoError(err)
	c.Equal(targetDate, date)
	_, err = cal.ParseDate("9/22")
	c.HasError(err)
	_, err = cal.ParseDate("9/666/2017")
	c.HasError(err)
	_, err = cal.ParseDate("13/22/2017")
	c.HasError(err)
	date, err = cal.ParseDate("September 22, 2017 AD")
	c.NoError(err)
	c.Equal(targetDate, date)
	date, err = cal.ParseDate("September 22, 1 BC")
	c.NoError(err)
	c.Equal(cal.MustNewDate(9, 22, -1), date)

	targetDate = cal.MustNewDate(9, 22, -2017)
	date, err = cal.ParseDate("9/22/-2017")
	c.NoError(err)
	c.Equal(targetDate, date)
	date, err = cal.ParseDate("September 22, -2017")
	c.NoError(err)
	c.Equal(targetDate, date)
	date, err = cal.ParseDate("9/22/2017 BC")
	c.NoError(err)
	c.Equal(targetDate, date)

	// A negative year combined with a previous-era suffix is contradictory and must not double-negate back into
	// the current era.
	_, err = cal.ParseDate("9/22/-2017 BC")
	c.HasError(err)
	_, err = cal.ParseDate("September 22, -2017 BC")
	c.HasError(err)

	// A negative year combined with the current-era suffix is contradictory in the same way: the minus sign places
	// the year in the previous era while the suffix names the current era. Reject it, symmetrically with the
	// previous-era case above. A non-negative year with the current-era suffix remains valid.
	_, err = cal.ParseDate("9/22/-5 AD")
	c.HasError(err)
	_, err = cal.ParseDate("September 22, -5 AD")
	c.HasError(err)
	date, err = cal.ParseDate("9/22/2017 AD")
	c.NoError(err)
	c.Equal(cal.MustNewDate(9, 22, 2017), date)
}

// TestParseDateEraNamesBeyondSingleWord guards the era suffix against being matched as a generic alphabetic word. With
// eras "A.D." and "B.C." the previous-era date 2/15/-5 renders as "2/15/5 B.C.", and a word-only pattern captured just
// the "B", ignored it as unrecognized, and returned year +5 with no error. The patterns are now built from the
// calendar's own era names, so any label Valid accepts round-trips with its sign intact.
func TestParseDateEraNamesBeyondSingleWord(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		era, prev string
		text      string
		year      int
		rendered  bool // text is exactly what the date renders as in "%N/%D/%y"
	}{
		{dottedEra, dottedPrevEra, "2/15/5 B.C.", -5, true},
		{dottedEra, dottedPrevEra, "2/15/5 A.D.", 5, true},
		{dottedEra, dottedPrevEra, "Febris 15, 5 B.C.", -5, false},
		{dottedEra, dottedPrevEra, "2/15/5 b.c.", -5, false}, // era names match without regard to case
		{wordyEra, wordyPrevEra, "2/15/5 Age of Dark", -5, true},
		{wordyEra, wordyPrevEra, "Febris 15, 5 Age of Dark", -5, false},
		{wordyEra, wordyPrevEra, "The battle of 2/15/5 Age of Dark, as told later", -5, false},
		{wordyEra, wordyPrevEra, "2/15/5 Age of Light", 5, true},
		// An era name is only recognized as a whole: a longer word that merely begins with it is not the era, so it
		// is ignored like any other unrecognized suffix rather than silently flipping the year's sign.
		{"AD", "BC", "2/15/5 BCE", 5, false},
		{dottedEra, dottedPrevEra, "2/15/5 B.C.E.", 5, false},
		{"AD", "BC", "2/15/5 BC.", -5, false}, // trailing punctuation does not stop the era from matching
		{"AD", "BC", "2/15/5BC", -5, false},   // nor does the absence of a space, as before
	} {
		cal, err := eraTestCalendar(tc.era, tc.prev)
		c.NoError(err)
		want := cal.MustNewDate(2, 15, tc.year)
		if tc.rendered {
			c.Equal(tc.text, want.Format("%N/%D/%y"), "eras %q/%q year %d", tc.era, tc.prev, tc.year)
		}
		got, err := cal.ParseDate(tc.text)
		c.NoError(err, "eras %q/%q: %q", tc.era, tc.prev, tc.text)
		c.Equal(want, got, "eras %q/%q: %q", tc.era, tc.prev, tc.text)
	}
}

// TestParseDateMonthNamesBeyondSingleWord guards the month name against being matched as a generic alphabetic word.
// Valid permits month names with spaces, punctuation and non-ASCII letters, and every format that emits a month name
// (LongFormat, FullFormat and the abbreviating MediumFormat) must parse back to the same date on such a calendar. The
// issue's own example: with months "New Moon" and "Old Moon", "New Moon 5, 100" failed with 'invalid month text "Moon"'.
func TestParseDateMonthNamesBeyondSingleWord(t *testing.T) {
	c := check.New(t)
	cal, err := calendar.New(&calendar.Config{
		WeekDays: []string{"Firstday", "Lastday"},
		Months: []calendar.Month{
			{Name: "New Moon", Days: 30},
			{Name: "Old Moon", Days: 30},
			{Name: "Harvest's End", Days: 30},
			{Name: "Frühling", Days: 30},
		},
	})
	c.NoError(err)
	date, err := cal.ParseDate("New Moon 5, 100")
	c.NoError(err)
	c.Equal(cal.MustNewDate(1, 5, 100), date)
	for month := 1; month <= 4; month++ {
		want := cal.MustNewDate(month, 5, 100)
		for _, layout := range []string{calendar.LongFormat, calendar.FullFormat, calendar.MediumFormat} {
			formatted := want.Format(layout)
			var got calendar.Date
			got, err = cal.ParseDate(formatted)
			c.NoError(err, "month %d formatted as %q", month, formatted)
			c.Equal(want, got, "month %d round-trip via %q", month, formatted)
		}
	}
	// Dates remain extractable from prose, and the name matches without regard to case.
	date, err = cal.ParseDate("It was Harvest's End 12, 100 when the caravan arrived.")
	c.NoError(err)
	c.Equal(cal.MustNewDate(3, 12, 100), date)
	date, err = cal.ParseDate("old moon 5, 100")
	c.NoError(err)
	c.Equal(cal.MustNewDate(2, 5, 100), date)
	// A fragment of a name is not a month.
	_, err = cal.ParseDate("Moon 5, 100")
	c.HasError(err)
}

// TestParseDateNamesWithRegexMetacharacters guards the quoting in the per-calendar parse patterns. Month and era names
// are interpolated into a regular expression, so any metacharacters in them must be matched literally: a month named
// ".*" must not become a wildcard that swallows the surrounding text, "(" must not break compilation, and "|" must not
// split a name into two alternatives.
func TestParseDateNamesWithRegexMetacharacters(t *testing.T) {
	c := check.New(t)
	var cal *calendar.Calendar
	var err error
	c.NotPanics(func() {
		cal, err = calendar.New(&calendar.Config{
			WeekDays: []string{"A", "B"},
			Months: []calendar.Month{
				{Name: ".*", Days: 30},
				{Name: "(unbalanced", Days: 30},
				{Name: "X|Bar", Days: 30},
				{Name: `back\\slash [a-z]+ $^?`, Days: 30},
			},
			Era:         "A.D.",
			PreviousEra: "(B|C)+",
		})
	})
	c.NoError(err)
	for month := 1; month <= 4; month++ {
		want := cal.MustNewDate(month, 5, 100)
		for _, layout := range []string{calendar.LongFormat, calendar.MediumFormat, "%M %D, %y"} {
			formatted := want.Format(layout)
			var got calendar.Date
			got, err = cal.ParseDate(formatted)
			c.NoError(err, "month %d formatted as %q", month, formatted)
			c.Equal(want, got, "month %d round-trip via %q", month, formatted)
		}
	}
	// Each metacharacter name matches only itself, never what it would mean as a pattern.
	_, err = cal.ParseDate("anything 5, 100") // ".*" is not a wildcard
	c.HasError(err)
	_, err = cal.ParseDate("X 5, 100") // "X|Bar" is one name, not two
	c.HasError(err)
	_, err = cal.ParseDate("Bar 5, 100")
	c.HasError(err)
	_, err = cal.ParseDate("backslash 5, 100") // the literal backslash is required
	c.HasError(err)
	got, err := cal.ParseDate("1/5/100 BCB") // "(B|C)+" is a literal label, not a repeated group
	c.NoError(err)
	c.Equal(100, got.Year())
	got, err = cal.ParseDate("1/5/100 (B|C)+")
	c.NoError(err)
	c.Equal(-100, got.Year())
}

func TestParseDateSharedEraSuffix(t *testing.T) {
	c := check.New(t)
	// When a calendar uses the same name for both eras (as the Pathfinder calendars do), the suffix cannot
	// disambiguate the year, so a negative year combined with that suffix is NOT a contradiction and must still
	// parse, unlike the distinct-era cases rejected in TestParseDate.
	cfg := calendar.Gregorian().Config()
	cfg.Era = "AR"
	cfg.PreviousEra = "AR"
	cal, err := calendar.New(cfg)
	c.NoError(err)
	date, err := cal.ParseDate("9/22/-5 AR")
	c.NoError(err)
	c.Equal(cal.MustNewDate(9, 22, -5), date)
	date, err = cal.ParseDate("9/22/2017 AR")
	c.NoError(err)
	c.Equal(cal.MustNewDate(9, 22, 2017), date)
}

func TestParseDateAmbiguousMonthAbbreviation(t *testing.T) {
	c := check.New(t)
	// "Marbol" and "Martok" share the first three letters, so the abbreviation "Mar" cannot identify either one.
	cal, err := calendar.New(&calendar.Config{
		WeekDays: []string{"One", "Two", "Three", "Four", "Five"},
		Months: []calendar.Month{
			{Name: "Marbol", Days: 30},
			{Name: "Martok", Days: 30},
			{Name: "June", Days: 30},
		},
	})
	c.NoError(err)

	// The ambiguous abbreviation must be rejected rather than silently resolving to the first match.
	_, err = cal.ParseDate("Mar 5, 1200")
	c.HasError(err)
	c.True(strings.Contains(err.Error(), "ambiguous"), "expected an ambiguity error, got: %v", err)

	// Full names disambiguate, each resolving to its own month.
	date, err := cal.ParseDate("Marbol 5, 1200")
	c.NoError(err)
	c.Equal(1, date.Month())
	date, err = cal.ParseDate("Martok 5, 1200")
	c.NoError(err)
	c.Equal(2, date.Month())

	// An abbreviation that is unique still works, as does the full name.
	date, err = cal.ParseDate("Jun 5, 1200")
	c.NoError(err)
	c.Equal(3, date.Month())
	date, err = cal.ParseDate("June 5, 1200")
	c.NoError(err)
	c.Equal(3, date.Month())
}

func TestMarshaling(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	date := cal.MustNewDate(9, 22, 2017)
	text, err := date.MarshalText()
	c.NoError(err)
	c.Equal("9/22/2017", string(text))

	type embedded struct {
		Date calendar.Date
	}
	embeddedDate := embedded{Date: date}
	text, err = json.Marshal(&embeddedDate)
	c.NoError(err)
	c.Equal(`{"Date":"9/22/2017"}`, string(text))

	text, err = yaml.Marshal(&embeddedDate)
	c.NoError(err)
	c.Equal("date: 9/22/2017\n", string(text))

	type embeddedPtr struct {
		Date *calendar.Date
	}
	embeddedPtrDate := embeddedPtr{Date: &date}
	text, err = json.Marshal(&embeddedPtrDate)
	c.NoError(err)
	c.Equal(`{"Date":"9/22/2017"}`, string(text))

	text, err = yaml.Marshal(&embeddedPtrDate)
	c.NoError(err)
	c.Equal("date: 9/22/2017\n", string(text))
}

func TestZeroValueDate(t *testing.T) {
	c := check.New(t)

	// A zero-value Date has no associated calendar; accessors and formatting must fall back to
	// Default rather than panicking with a nil pointer dereference (mirroring UnmarshalText).
	var date calendar.Date
	c.NotPanics(func() {
		c.Equal(1, date.Year())
		c.Equal(1, date.Month())
		c.Equal(1, date.DayInMonth())
		c.Equal("1/1/1", date.String())
	})

	text, err := date.MarshalText()
	c.NoError(err)
	c.Equal("1/1/1", string(text))

	// json.Marshal of a struct holding an unset Date field must not panic.
	type embedded struct {
		Date calendar.Date
	}
	text, err = json.Marshal(&embedded{})
	c.NoError(err)
	c.Equal(`{"Date":"1/1/1"}`, string(text))
}

func TestUnmarshaling(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	target := cal.MustNewDate(9, 22, 2017)
	var date calendar.Date
	c.NoError(date.UnmarshalText([]byte("9/22/2017")))
	c.Equal(target, date)

	type embedded struct {
		Date calendar.Date
	}
	var embeddedDate embedded
	c.NoError(json.Unmarshal([]byte(`{"Date":"9/22/2017"}`), &embeddedDate))
	c.Equal(target, embeddedDate.Date)

	c.NoError(yaml.Unmarshal([]byte(`date: 9/22/2017`), &embeddedDate))
	c.Equal(target, embeddedDate.Date)

	type embeddedPtr struct {
		Date *calendar.Date
	}
	var embeddedPtrDate embeddedPtr
	c.NoError(json.Unmarshal([]byte(`{"Date":"9/22/2017"}`), &embeddedPtrDate))
	c.Equal(target, *embeddedPtrDate.Date)

	c.NoError(yaml.Unmarshal([]byte(`date: 9/22/2017`), &embeddedPtrDate))
	c.Equal(target, *embeddedPtrDate.Date)

	cal = calendar.PathfinderAbsalomReckoning()
	savedDefault := calendar.Default()
	defer func() { calendar.SetDefault(savedDefault) }()
	calendar.SetDefault(cal)
	date = calendar.Date{}
	target = cal.MustNewDate(9, 22, 2017)
	c.NoError(date.UnmarshalText([]byte("9/22/2017 AR")))
	c.Equal(target, date)
}

// TestAddSaturates exercises Date.Add's documented saturation contract directly, in both directions. A Date is
// confined to the years IsValidYear accepts, so the limits are the first day of year math.MinInt32 and the last day of
// year math.MaxInt32 on the date's own calendar. A sum that wraps around int64 lands on the far side of zero, and Add
// once compared that wrapped sum against the limit before checking for the wrap, so a negative overflow saturated at
// the upper limit rather than the lower.
func TestAddSaturates(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	c.Equal(int64(10), cal.NewDateByDays(3).Add(7).Days())
	c.Equal(int64(-4), cal.NewDateByDays(3).Add(-7).Days())
	c.Equal(int64(5), calendar.Date{}.Add(5).Days())

	// The limits themselves are representable; one day past either is not.
	last := cal.MustNewDate(12, 31, math.MaxInt32).Days()
	first := cal.MustNewDate(1, 1, math.MinInt32).Days()
	c.Equal(last, cal.NewDateByDays(last-1).Add(1).Days())
	c.Equal(last, cal.NewDateByDays(last).Add(1).Days())
	c.Equal(last, cal.NewDateByDays(0).Add(last+1).Days())
	c.Equal(first, cal.NewDateByDays(first+1).Add(-1).Days())
	c.Equal(first, cal.NewDateByDays(first).Add(-1).Days())
	c.Equal(first, cal.NewDateByDays(0).Add(first-1).Days())

	// A sum that wraps around int64 saturates toward the sign of the delta, not the sign of the wrapped result.
	c.Equal(last, cal.NewDateByDays(1).Add(math.MaxInt64).Days())
	c.Equal(last, cal.NewDateByDays(last).Add(math.MaxInt64).Days())
	c.Equal(first, cal.NewDateByDays(-1).Add(math.MinInt64).Days())
	c.Equal(first, cal.NewDateByDays(first).Add(math.MinInt64).Days())

	// A delta large enough to cross the whole span without wrapping saturates at the far limit.
	c.Equal(first, cal.NewDateByDays(last).Add(math.MinInt64).Days())
	c.Equal(last, cal.NewDateByDays(first).Add(math.MaxInt64).Days())

	// The limits belong to the calendar: one with a different leap rule ends year math.MaxInt32 on a different day
	// count, but still saturates at 12/31 of that year.
	other := calendar.PathfinderAbsalomReckoning()
	c.True(other.NewDateByDays(math.MaxInt64).Days() != last)
	for _, each := range []*calendar.Calendar{cal, other} {
		c.Equal("12/31/2147483647", each.NewDateByDays(math.MaxInt64).Format("%N/%D/%z"))
		c.Equal("1/1/-2147483648", each.NewDateByDays(math.MinInt64).Format("%N/%D/%z"))
	}

	// The calendar with the longest years Config.Valid allows saturates at the farthest days any Date can reach.
	huge := maximalCalendar(c, false)
	c.Equal(int64(-1<<61), huge.NewDateByDays(math.MinInt64).Days())
	c.Equal(int64(-1<<61), huge.NewDateByDays(-1).Add(math.MinInt64).Days())
	c.Equal(math.MinInt32, huge.NewDateByDays(math.MinInt64).Year())
	c.Equal(int64(1<<61-1<<30-1), huge.NewDateByDays(math.MaxInt64).Days())
	c.Equal(math.MaxInt32, huge.NewDateByDays(math.MaxInt64).Year())
}

// TestDatesConfinedToValidYears verifies that no Date can leave the years IsValidYear accepts, which is what lets Year
// return an int32: stepping past the last day of year math.MaxInt32 or before the first day of year math.MinInt32
// saturates at that boundary day, and the boundary days resolve, format, parse back and render without panicking.
// Dates beyond the int32 years were once reachable through Add, and their leap status was then misjudged by the
// range-checked IsLeapYear, so 2/29 of year 2^31 rendered as 3/1 and its last day panicked with "unable to determine
// month".
func TestDatesConfinedToValidYears(t *testing.T) {
	c := check.New(t)
	for _, cal := range []*calendar.Calendar{calendar.Gregorian(), calendar.PathfinderAbsalomReckoning()} {
		last := cal.MustNewDate(12, 31, math.MaxInt32)
		c.Equal(last, last.Add(1))
		c.Equal(last, last.Add(60))
		c.Equal(last, last.Add(366))
		c.Equal(math.MaxInt32, last.Add(1).Year())
		c.Equal(last.Days(), cal.NewDateByDays(math.MaxInt64).Days())
		first := cal.MustNewDate(1, 1, math.MinInt32)
		c.Equal(first, first.Add(-1))
		c.Equal(first, first.Add(-60))
		c.Equal(first, first.Add(-366))
		c.Equal(math.MinInt32, first.Add(-1).Year())
		c.Equal(first.Days(), cal.NewDateByDays(math.MinInt64).Days())
		for _, d := range []calendar.Date{last, first} {
			text := d.Format(calendar.ShortFormat)
			c.NotPanics(func() {
				c.True(d.Month() >= 1 && d.Month() <= 12, "month of %s", text)
				c.True(d.DayInMonth() >= 1 && d.DayInMonth() <= d.DaysInMonth(), "day of %s", text)
				_ = d.DayInYear()
				_ = d.WeekDay()
				_, _ = d.Season()
				_ = d.Format(calendar.FullFormat)
				var buf bytes.Buffer
				d.TextCalendarMonth(&buf)
			}, "%s", text)
			// The boundary years round-trip through their own text, including the previous-era magnitude 2147483648
			// that only fits an int32 once its sign is restored.
			got, err := cal.ParseDate(text)
			c.NoError(err, "%s", text)
			c.True(got.Equal(d), "%s parsed back as %s", text, got.Format(calendar.ShortFormat))
		}
	}
	greg := calendar.Gregorian()
	c.Equal("1/1/2147483648 BC", greg.MustNewDate(1, 1, math.MinInt32).String())
	_, err := greg.ParseDate("1/1/2147483648")
	c.HasError(err)
	_, err = greg.ParseDate("1/1/2147483649 BC")
	c.HasError(err)
	_, err = greg.ParseDate("1/1/-2147483649")
	c.HasError(err)
	// The full year at each end renders.
	var buf bytes.Buffer
	c.NoError(greg.Text(math.MaxInt32, &buf))
	c.True(strings.HasPrefix(buf.String(), "Year 2147483647\n"))
	buf.Reset()
	c.NoError(greg.Text(math.MinInt32, &buf))
	c.True(strings.HasPrefix(buf.String(), "Year 2147483648 BC\n"))
}

// TestFormatPassesThroughNonDirectives verifies that a % which does not introduce a directive is emitted as written,
// along with the character after it, and that a trailing % is not lost. The directive switch once had no default case,
// so "50% off" rendered as "50off" and "100%" as "100".
func TestFormatPassesThroughNonDirectives(t *testing.T) {
	c := check.New(t)
	d := calendar.Gregorian().MustNewDate(9, 22, 2017)
	c.Equal("50% off", d.Format("50% off"))
	c.Equal("100%", d.Format("100%"))
	c.Equal("%Q", d.Format("%Q"))
	c.Equal("%", d.Format("%"))
	c.Equal("%é", d.Format("%é")) // a multi-byte character after the % survives intact
	c.Equal("%x9", d.Format("%x%N"))
	c.Equal("9/22/2017%", d.Format("%N/%D/%Y%"))
	c.Equal("%%", d.Format("%%%")) // an escaped % followed by a trailing one
	var buf strings.Builder
	d.WriteFormat(&buf, "%Y%")
	c.Equal("2017%", buf.String())
}

// TestDateEqual pins Date.Equal, which exists because == cannot compare a zero Date{} with the same day obtained from
// Default(): the former carries no calendar reference and the latter does, though both resolve to the default calendar.
func TestDateEqual(t *testing.T) {
	c := check.New(t)
	greg := calendar.Gregorian()
	d := greg.MustNewDate(9, 22, 2017)
	same := d
	c.True(d.Equal(same))
	c.True(d.Equal(greg.NewDateByDays(d.Days())))
	c.False(d.Equal(d.Add(1)))
	c.False(d.Add(1).Equal(d))
	// The same day count on a different calendar is a different date.
	c.False(d.Equal(calendar.PathfinderAbsalomReckoning().NewDateByDays(d.Days())))

	zero := calendar.Date{}
	fromDefault := calendar.Default().NewDateByDays(0)
	c.Equal(zero.String(), fromDefault.String())
	c.True(zero != fromDefault, "== distinguishes the two, which is the trap Equal exists for")
	c.True(zero.Equal(fromDefault))
	c.True(fromDefault.Equal(zero))
	c.True(zero.Equal(calendar.Date{}))
	c.True(zero.Equal(zero.Add(0)))
	c.False(zero.Equal(zero.Add(1)))
	c.False(zero.Equal(fromDefault.Add(-1)))
}
