// Copyright (c) 2017-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

// Package calendar provides a customizable calendar for roleplaying games.
package calendar

import (
	"fmt"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

var (
	// Default is the default calendar that will be used by Date.UnmarshalText() if the date was not initialized. If you
	// modify this (either the pointer to the Calendar, or any contents within the Calendar), you are responsible for
	// ensuring it is done in a thread-safe context, as this code assumes it is effectively immutable when used.
	Default = Gregorian()
	// "9/22/2017" or "9/22/2017 AD"
	regexMMDDYYYY = regexp.MustCompile("([[:digit:]]+)/([[:digit:]]+)/(-?[[:digit:]]+) *([[:alpha:]]+)?")
	// "September 22, 2017 AD", "September 22, 2017", "Sep 22, 2017 AD", or "Sep 22, 2017"
	regexMonthDDYYYY = regexp.MustCompile("([[:alpha:]]+) *([[:digit:]]+), *(-?[[:digit:]]+) *([[:alpha:]]+)?")
)

// abbreviatedNameLength is the number of leading characters used to abbreviate a month or weekday name. The %m and %w
// format directives emit this many characters, and monthFromText accepts a month abbreviation of this length, so an
// emitted short month name parses back to the same month (e.g. MediumFormat round-trips through ParseDate). Keeping
// both sides driven by this single constant prevents the emit and parse widths from silently drifting apart.
const abbreviatedNameLength = 3

// Calendar holds the data for the calendar.
type Calendar struct {
	LeapYear       *LeapYear `json:"leapyear,omitempty" yaml:",omitempty"`
	Era            string    `json:"era,omitempty" yaml:",omitempty"`
	PreviousEra    string    `json:"previous_era,omitempty" yaml:"previous_era,omitempty"`
	WeekDays       []string  `json:"weekdays"`
	Months         []Month   `json:"months"`
	Seasons        []Season  `json:"seasons"`
	DayZeroWeekDay int       `json:"day_zero_weekday" yaml:"day_zero_weekday"`
}

// Valid returns nil if the calendar data is usable.
func (cal *Calendar) Valid() error {
	if err := cal.checkUsable(); err != nil {
		return err
	}
	if len(cal.Seasons) == 0 {
		return errs.New("Calendar must have at least one season")
	}
	if slices.Contains(cal.WeekDays, "") {
		return errs.New("Calendar week day names must not be empty")
	}
	for _, month := range cal.Months {
		if err := month.Valid(); err != nil {
			return err
		}
	}
	for i := range cal.Seasons {
		if err := cal.Seasons[i].Valid(cal); err != nil {
			return err
		}
	}
	return nil
}

// checkUsable returns an error unless the calendar is structurally safe for date math: a non-empty set of week days, an
// in-range day-zero week day, at least one month with at least one day, and a valid leap year rule when one is present.
func (cal *Calendar) checkUsable() error {
	if len(cal.WeekDays) == 0 {
		return errs.New("Calendar must have at least one week day")
	}
	if cal.DayZeroWeekDay < 0 || cal.DayZeroWeekDay >= len(cal.WeekDays) {
		return errs.New("Calendar's first week day of the first year must be a valid week day")
	}
	if cal.MinDaysPerYear() < 2 {
		return errs.New("Calendar must have at least one month and a total of at least two days")
	}
	if cal.LeapYear != nil {
		if err := cal.LeapYear.Valid(cal); err != nil {
			return err
		}
	}
	return nil
}

// mustBeUsable panics with the error checkUsable reports when the calendar cannot support date math. The constructors
// reject an unusable calendar up front, but a zero-value Date (or one produced by UnmarshalText) resolves through
// Default, which a caller can reassign to a deserialized, never-validated calendar. The Date accessors call this before
// they divide by the days-per-year or the week-day count, turning what would otherwise be an opaque "integer divide by
// zero" panic deep inside an accessor into the same actionable error the constructors already give.
func (cal *Calendar) mustBeUsable() {
	if err := cal.checkUsable(); err != nil {
		panic(err) // @allow
	}
}

// MustNewDate creates a new date from the specified month, day and year. Panics if the values are invalid.
func (cal *Calendar) MustNewDate(month, day, year int) Date {
	date, err := cal.NewDate(month, day, year)
	if err != nil {
		panic(err) // @allow
	}
	return date
}

// NewDate creates a new date from the specified month, day and year.
func (cal *Calendar) NewDate(month, day, year int) (Date, error) {
	if err := cal.checkUsable(); err != nil {
		return Date{cal: cal}, err
	}
	if year == 0 {
		return Date{cal: cal}, errs.New("year 0 is invalid")
	}
	if month < 1 || month > len(cal.Months) {
		return Date{cal: cal}, errs.Newf("month %d is invalid", month)
	}
	days := cal.Months[month-1].Days
	if cal.IsLeapMonth(month) && cal.IsLeapYear(year) {
		days++
	}
	if day < 1 || day > days {
		return Date{cal: cal}, errs.Newf("day %d is invalid", day)
	}
	days = cal.yearToDays(year) + day - 1
	for i := 1; i < month; i++ {
		days += cal.Months[i-1].Days
	}
	if cal.IsLeapYear(year) && cal.LeapYear.Month < month {
		days++
	}
	return Date{Days: days, cal: cal}, nil
}

// NewDateByDays creates a new date from a number of days, with 0 representing the date 1/1/1. It panics if the calendar
// is not usable for date math; call Valid in advance to check a calendar without risking a panic.
func (cal *Calendar) NewDateByDays(days int) Date {
	if err := cal.checkUsable(); err != nil {
		panic(err) // @allow
	}
	return Date{Days: days, cal: cal}
}

func (cal *Calendar) yearToDays(year int) int {
	return cal.yearToDaysWith(year, cal.MinDaysPerYear())
}

// yearToDaysWith is yearToDays with the minimum days per year supplied by the caller, so that callers iterating over
// candidate years (such as Date.Year) can hoist the O(months) MinDaysPerYear summation out of their loop.
func (cal *Calendar) yearToDaysWith(year, minDaysPerYear int) int {
	var days int
	if year > 1 {
		days = (year - 1) * minDaysPerYear
	} else if year < 0 {
		days = year * minDaysPerYear
	}
	if cal.LeapYear != nil {
		leaps := cal.LeapYear.Since(year)
		if year > 1 {
			days += leaps
		} else {
			days -= leaps
			if cal.LeapYear.Is(year) {
				days--
			}
		}
	}
	return days
}

// ParseDate creates a new date from the specified text.
func (cal *Calendar) ParseDate(in string) (Date, error) {
	if parts := regexMMDDYYYY.FindStringSubmatch(in); parts != nil {
		month, err := strconv.Atoi(parts[1])
		if err != nil {
			return Date{cal: cal}, errs.NewWithCausef(err, "invalid month text '%s'", parts[1])
		}
		return cal.parseDate(month, parts[2], parts[3], parts[4])
	}
	if parts := regexMonthDDYYYY.FindStringSubmatch(in); parts != nil {
		month, err := cal.monthFromText(parts[1])
		if err != nil {
			return Date{cal: cal}, err
		}
		return cal.parseDate(month, parts[2], parts[3], parts[4])
	}
	return Date{cal: cal}, errs.Newf("invalid date text '%s'", in)
}

// monthFromText resolves a month name to its 1-based index. A full-name match is preferred; failing that, a 3-letter
// abbreviation is accepted only when it unambiguously identifies a single month, since two months whose names share the
// same first three letters cannot otherwise be told apart.
func (cal *Calendar) monthFromText(text string) (int, error) {
	for i := range cal.Months {
		if strings.EqualFold(text, cal.Months[i].Name) {
			return i + 1, nil
		}
	}
	month := 0
	for i := range cal.Months {
		if strings.EqualFold(text, xstrings.FirstN(cal.Months[i].Name, abbreviatedNameLength)) {
			if month != 0 {
				return 0, errs.Newf("ambiguous month text '%s'", text)
			}
			month = i + 1
		}
	}
	if month == 0 {
		return 0, errs.Newf("invalid month text '%s'", text)
	}
	return month, nil
}

func (cal *Calendar) parseDate(month int, dayText, yearText, eraText string) (Date, error) {
	year, err := strconv.Atoi(yearText)
	if err != nil {
		return Date{cal: cal}, errs.NewWithCausef(err, "invalid year text '%s'", yearText)
	}
	day, err := strconv.Atoi(dayText)
	if err != nil {
		return Date{cal: cal}, errs.NewWithCausef(err, "invalid day text '%s'", dayText)
	}
	if year, err = cal.resolveEraSuffix(year, yearText, eraText); err != nil {
		return Date{cal: cal}, err
	}
	return cal.NewDate(month, day, year)
}

// eraForYear maps a signed internal year to the year value and era label that represent it for display. It is the
// single definition of the calendar's era model that Date.Era and the %y/%Y format directives build on, and
// resolveEraSuffix is its parse-side inverse. A negative year belongs to the previous era and a non-negative year to
// the current era. When the two eras are distinct the era label carries the sign, so the magnitude is returned (a year
// of -5 with eras "AD"/"BC" yields 5, "BC"); when the eras are empty or identical there is no distinct label to carry
// the sign, so the signed year is returned unchanged (-5 with eras "AR"/"AR" yields -5, "AR").
func (cal *Calendar) eraForYear(year int) (displayYear int, era string) {
	era = cal.Era
	if year < 0 {
		era = cal.PreviousEra
	}
	displayYear = year
	if year < 0 && era != "" && cal.Era != cal.PreviousEra {
		displayYear = -year
	}
	return displayYear, era
}

// resolveEraSuffix folds a recognized era suffix into the sign of a parsed year, the parse-side inverse of eraForYear.
// A leading minus sign already places the year before the current era, so a recognized suffix must agree with it. When
// the calendar names its two eras distinctly, a previous-era suffix on a non-negative year selects the previous era,
// but on a negative year it merely repeats the sign, and a current-era suffix on a negative year flatly contradicts it;
// reject both rather than silently choosing an interpretation. An empty or unrecognized suffix is left alone so
// ParseDate can still find dates embedded in surrounding prose. yearText and eraText are the original matched text,
// used only for the error messages.
func (cal *Calendar) resolveEraSuffix(year int, yearText, eraText string) (int, error) {
	distinctEras := cal.Era != cal.PreviousEra
	previousEraSuffix := eraText != "" && distinctEras && strings.EqualFold(cal.PreviousEra, eraText)
	currentEraSuffix := eraText != "" && distinctEras && strings.EqualFold(cal.Era, eraText)
	switch {
	case year < 0 && previousEraSuffix:
		return 0, errs.Newf("year '%s' and previous-era suffix '%s' both indicate the previous era", yearText, eraText)
	case year < 0 && currentEraSuffix:
		return 0, errs.Newf("negative year '%s' contradicts the current-era suffix '%s'", yearText, eraText)
	case previousEraSuffix:
		year = -year
	}
	return year, nil
}

// MinDaysPerYear returns the minimum number of days in a year.
func (cal *Calendar) MinDaysPerYear() int {
	days := 0
	for _, month := range cal.Months {
		days += month.Days
	}
	return days
}

// maxDaysInMonth returns the largest number of days the given 1-based month can hold, including the extra day the leap
// month gains in a leap year. It is the upper bound for any day-of-month within that month across all years, so a date
// or season boundary is valid as long as it does not exceed this.
func (cal *Calendar) maxDaysInMonth(month int) int {
	days := cal.Months[month-1].Days
	if cal.IsLeapMonth(month) {
		days++
	}
	return days
}

// mostDaysInMonth returns the largest number of days any single month can contain, including the extra day the leap
// month gains in a leap year. It is used to size day-of-month fields to a consistent width regardless of which month or
// year is being formatted.
func (cal *Calendar) mostDaysInMonth() int {
	most := 0
	for i := range cal.Months {
		most = max(most, cal.maxDaysInMonth(i+1))
	}
	return most
}

// Days returns the number of days contained in a specific year.
func (cal *Calendar) Days(year int) int {
	days := cal.MinDaysPerYear()
	if cal.IsLeapYear(year) {
		days++
	}
	return days
}

// IsLeapYear returns true if the year is a leap year.
func (cal *Calendar) IsLeapYear(year int) bool {
	return cal.LeapYear != nil && cal.LeapYear.Is(year)
}

// IsLeapMonth returns true if the month is the leap month.
func (cal *Calendar) IsLeapMonth(month int) bool {
	return cal.LeapYear != nil && cal.LeapYear.Month == month
}

// Text writes a text representation of the year.
func (cal *Calendar) Text(year int, w io.Writer) {
	date := cal.MustNewDate(1, 1, year)
	date.WriteFormat(w, "Year %Y\n")
	width := widthNeeded(cal.mostDaysInMonth())
	maximum := len(cal.Months)
	for i := 1; i <= maximum; i++ {
		fmt.Fprintln(w)
		cal.MustNewDate(i, 1, year).textCalendarMonth(w, width)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Seasons:")
	for i := range cal.Seasons {
		fmt.Fprintf(w, "  %v\n", &cal.Seasons[i])
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Week Days:")
	for i, weekday := range cal.WeekDays {
		fmt.Fprintf(w, "  %d: (%s) %s\n", i+1, xstrings.FirstN(weekday, 1), weekday)
	}
}
