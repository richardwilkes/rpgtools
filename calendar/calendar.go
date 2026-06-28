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
	// Default is the default calendar that will be used by Date.UnmarshalText() if the date was not initialized.
	Default = Gregorian()
	// "9/22/2017" or "9/22/2017 AD"
	regexMMDDYYY = regexp.MustCompile("([[:digit:]]+)/([[:digit:]]+)/(-?[[:digit:]]+) *([[:alpha:]]+)?")
	// "September 22, 2017 AD", "September 22, 2017", "Sep 22, 2017 AD", or "Sep 22, 2017"
	regexMonthDDYYYY = regexp.MustCompile("([[:alpha:]]+) *([[:digit:]]+), *(-?[[:digit:]]+) *([[:alpha:]]+)?")
)

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
	if len(cal.WeekDays) == 0 {
		return errs.New("Calendar must have at least one week day")
	}
	if len(cal.Months) == 0 {
		return errs.New("Calendar must have at least one month")
	}
	if len(cal.Seasons) == 0 {
		return errs.New("Calendar must have at least one season")
	}
	if cal.DayZeroWeekDay < 0 || cal.DayZeroWeekDay >= len(cal.WeekDays) {
		return errs.New("Calendar's first week day of the first year must be a valid week day")
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
	if cal.LeapYear != nil {
		if err := cal.LeapYear.Valid(cal); err != nil {
			return err
		}
	}
	return nil
}

// checkUsable returns an error if the calendar lacks the minimum structure required for date math: at least one week
// day and at least one month with at least one day. This is a deliberately narrow subset of Valid that ignores cosmetic
// concerns (such as empty week day names or missing seasons) so that dates cannot be created against a calendar whose
// accessors would later panic with a divide-by-zero, while still accepting the calendars the rest of the package is
// expected to tolerate.
func (cal *Calendar) checkUsable() error {
	if len(cal.WeekDays) == 0 {
		return errs.New("Calendar must have at least one week day")
	}
	if cal.MinDaysPerYear() < 1 {
		return errs.New("Calendar must have at least one month with at least one day")
	}
	return nil
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
	if parts := regexMMDDYYY.FindStringSubmatch(in); parts != nil {
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
		if strings.EqualFold(text, xstrings.FirstN(cal.Months[i].Name, 3)) {
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
	if cal.PreviousEra != "" && cal.PreviousEra != cal.Era && strings.EqualFold(cal.PreviousEra, eraText) {
		// A leading minus sign and the previous-era suffix both place the year before the current era, so
		// accepting both would double-negate the year back into the current era. Reject the contradiction rather
		// than silently pick one interpretation.
		if year < 0 {
			return Date{cal: cal}, errs.Newf("year '%s' and previous-era suffix '%s' both indicate the previous era",
				yearText, eraText)
		}
		year = -year
	}
	return cal.NewDate(month, day, year)
}

// MinDaysPerYear returns the minimum number of days in a year.
func (cal *Calendar) MinDaysPerYear() int {
	days := 0
	for _, month := range cal.Months {
		days += month.Days
	}
	return days
}

// mostDaysInMonth returns the largest number of days any single month can contain, including the extra day the leap
// month gains in a leap year. It is used to size day-of-month fields to a consistent width regardless of which month or
// year is being formatted.
func (cal *Calendar) mostDaysInMonth() int {
	most := 0
	for i, month := range cal.Months {
		days := month.Days
		if cal.IsLeapMonth(i + 1) {
			days++
		}
		most = max(most, days)
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
	maximum := len(cal.Months)
	for i := 1; i <= maximum; i++ {
		fmt.Fprintln(w)
		cal.MustNewDate(i, 1, year).TextCalendarMonth(w)
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
