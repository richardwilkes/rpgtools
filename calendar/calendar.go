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
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

var (
	defaultCalendarLock sync.RWMutex
	defaultCalendar     = Gregorian()
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

// Calendar holds the data for a calendar.
type Calendar struct {
	cfg *Config
}

// New creates a new Calendar from the given Config.
func New(cfg *Config) (*Calendar, error) {
	if err := cfg.Valid(); err != nil {
		return nil, err
	}
	return &Calendar{cfg: cfg.Clone()}, nil
}

// Default returns the default Calendar that will be used if one isn't explicitly used (for example, if you create a
// Date directly via Date{} rather than via a Calendar).
func Default() *Calendar {
	defaultCalendarLock.RLock()
	defer defaultCalendarLock.RUnlock()
	return defaultCalendar
}

// SetDefault sets the default Calendar to use.
func SetDefault(cal *Calendar) {
	if cal != nil && cal.cfg != nil {
		defaultCalendarLock.Lock()
		defaultCalendar = cal
		defaultCalendarLock.Unlock()
	}
}

// Config returns a clone of this Calendar's Config.
func (c *Calendar) Config() *Config {
	return c.config().Clone()
}

func (c *Calendar) config() *Config {
	if c != nil && c.cfg != nil {
		return c.cfg
	}
	return Default().cfg
}

// MustNewDate creates a new date from the specified month, day and year. Panics if the values are invalid.
func (c *Calendar) MustNewDate(month, day, year int) Date {
	date, err := c.NewDate(month, day, year)
	if err != nil {
		panic(err) // @allow
	}
	return date
}

// NewDate creates a new date from the specified month, day and year.
func (c *Calendar) NewDate(month, day, year int) (Date, error) {
	if !isValidYear(year) {
		return Date{cal: c}, errs.Newf("year %d is invalid; must be in the range %d to %d, not including 0", year,
			math.MinInt32, math.MaxInt32)
	}
	cfg := c.config()
	if month < 1 || month > len(cfg.Months) {
		return Date{cal: c}, errs.Newf("month %d is invalid; must be in the range 1 to %d", month, len(cfg.Months))
	}
	days := cfg.Months[month-1].Days
	if c.IsLeapMonth(month) && c.IsLeapYear(year) {
		days++
	}
	if day < 1 || day > days {
		return Date{cal: c}, errs.Newf("day %d is invalid; must be in the range 1 to %d for month %d", day, days, month)
	}
	days = c.yearToDays(year) + day - 1
	for i := 1; i < month; i++ {
		days += cfg.Months[i-1].Days
	}
	if c.IsLeapYear(year) && cfg.LeapYear.Month < month {
		days++
	}
	return c.NewDateByDays(days), nil
}

func isValidYear(year int) bool {
	return year != 0 && year >= math.MinInt32 && year <= math.MaxInt32
}

// NewDateByDays creates a new date from a number of days, with 0 representing the date 1/1/1.
func (c *Calendar) NewDateByDays(days int) Date {
	d := Date{cal: c}
	return d.Add(days)
}

func (c *Calendar) yearToDays(year int) int {
	return c.yearToDaysWith(year, c.MinDaysPerYear())
}

func (c *Calendar) yearToDaysWith(year, minDaysPerYear int) int {
	var days int
	if year > 1 {
		days = (year - 1) * minDaysPerYear
	} else if year < 0 {
		days = year * minDaysPerYear
	}
	if c.config().LeapYear != nil {
		leaps := c.leapYearsSince(year)
		if year > 1 {
			days += leaps
		} else {
			days -= leaps
			if c.isLeapYear(year) {
				days--
			}
		}
	}
	return days
}

// ParseDate creates a new date from the specified text.
func (c *Calendar) ParseDate(in string) (Date, error) {
	if parts := regexMMDDYYYY.FindStringSubmatch(in); parts != nil {
		month, err := strconv.Atoi(parts[1])
		if err != nil {
			return Date{cal: c}, errs.NewWithCausef(err, "invalid month text %q", parts[1])
		}
		return c.parseDate(month, parts[2], parts[3], parts[4])
	}
	if parts := regexMonthDDYYYY.FindStringSubmatch(in); parts != nil {
		month, err := c.monthFromText(parts[1])
		if err != nil {
			return Date{cal: c}, err
		}
		return c.parseDate(month, parts[2], parts[3], parts[4])
	}
	return Date{cal: c}, errs.Newf("invalid date text %q", in)
}

func (c *Calendar) parseDate(month int, dayText, yearText, eraText string) (Date, error) {
	year, err := strconv.Atoi(yearText)
	if err != nil {
		return Date{cal: c}, errs.NewWithCausef(err, "invalid year text %q", yearText)
	}
	day, err := strconv.Atoi(dayText)
	if err != nil {
		return Date{cal: c}, errs.NewWithCausef(err, "invalid day text %q", dayText)
	}
	if year, err = c.resolveEraSuffix(year, yearText, eraText); err != nil {
		return Date{cal: c}, err
	}
	return c.NewDate(month, day, year)
}

// monthFromText resolves a month name to its 1-based index. A full-name match is preferred; failing that, a 3-letter
// abbreviation is accepted only when it unambiguously identifies a single month, since two months whose names share the
// same first three letters cannot otherwise be told apart.
func (c *Calendar) monthFromText(text string) (int, error) {
	cfg := c.config()
	for i := range cfg.Months {
		if strings.EqualFold(text, cfg.Months[i].Name) {
			return i + 1, nil
		}
	}
	month := 0
	for i := range cfg.Months {
		if strings.EqualFold(text, xstrings.FirstN(cfg.Months[i].Name, abbreviatedNameLength)) {
			if month != 0 {
				return 0, errs.Newf("ambiguous month text %q", text)
			}
			month = i + 1
		}
	}
	if month == 0 {
		return 0, errs.Newf("invalid month text %q", text)
	}
	return month, nil
}

// eraForYear maps a signed internal year to the year value and era label that represent it for display. It is the
// single definition of the calendar's era model that Date.Era and the %y/%Y format directives build on, and
// resolveEraSuffix is its parse-side inverse. A negative year belongs to the previous era and a non-negative year to
// the current era. When the two eras are distinct the era label carries the sign, so the magnitude is returned (a year
// of -5 with eras "AD"/"BC" yields 5, "BC"); when the eras are empty or identical there is no distinct label to carry
// the sign, so the signed year is returned unchanged (-5 with eras "AR"/"AR" yields -5, "AR").
func (c *Calendar) eraForYear(year int) (displayYear int, era string) {
	cfg := c.config()
	era = cfg.Era
	if year < 0 {
		era = cfg.PreviousEra
	}
	displayYear = year
	if year < 0 && era != "" && cfg.Era != cfg.PreviousEra {
		displayYear = -year
	}
	return displayYear, era
}

// resolveEraSuffix folds a recognized era suffix into the sign of a parsed year, the parse-side inverse of eraForYear.
// A leading minus sign already places the year before the current era, so a recognized suffix must agree with it. When
// the calendar names its two eras distinctly, a previous-era suffix on a non-negative year selects the previous era,
// but on a negative year it merely repeats the sign, and a current-era suffix on a negative year flatly contradicts it;
// reject both rather than silently choosing an interpretation. An empty or unrecognized suffix is left alone so
// ParseDate can still find dates embedded in surrounding text. yearText and eraText are the original matched text,
// used only for the error messages.
func (c *Calendar) resolveEraSuffix(year int, yearText, eraText string) (int, error) {
	cfg := c.config()
	distinctEras := cfg.Era != cfg.PreviousEra
	previousEraSuffix := eraText != "" && distinctEras && strings.EqualFold(cfg.PreviousEra, eraText)
	currentEraSuffix := eraText != "" && distinctEras && strings.EqualFold(cfg.Era, eraText)
	switch {
	case year < 0 && previousEraSuffix:
		return 0, errs.Newf("year %q and previous-era suffix %q both indicate the previous era", yearText, eraText)
	case year < 0 && currentEraSuffix:
		return 0, errs.Newf("negative year %q contradicts the current-era suffix %q", yearText, eraText)
	case previousEraSuffix:
		year = -year
	}
	return year, nil
}

// MinDaysPerYear returns the minimum number of days in a year.
func (c *Calendar) MinDaysPerYear() int {
	cfg := c.config()
	days := 0
	for _, month := range cfg.Months {
		days += month.Days
	}
	return days
}

// maxDaysInMonth returns the largest number of days the given 1-based month can hold, including the extra day the leap
// month gains in a leap year. It is the upper bound for any day-of-month within that month across all years, so a date
// or season boundary is valid as long as it does not exceed this.
func (c *Calendar) maxDaysInMonth(month int) int {
	return c.config().maxDaysInMonth(month)
}

// mostDaysInMonth returns the largest number of days any single month can contain, including the extra day the leap
// month gains in a leap year. It is used to size day-of-month fields to a consistent width regardless of which month or
// year is being formatted.
func (c *Calendar) mostDaysInMonth() int {
	cfg := c.config()
	most := 0
	for i := range cfg.Months {
		most = max(most, cfg.maxDaysInMonth(i+1))
	}
	return most
}

// Days returns the number of days contained in a specific year.
func (c *Calendar) Days(year int) int {
	days := c.MinDaysPerYear()
	if c.IsLeapYear(year) {
		days++
	}
	return days
}

// IsLeapYear returns true if the year is a leap year. Note that valid years are constrained to not 0 and in the range
// math.MinInt32 to math.MaxInt32, so an invalid year will always return false.
func (c *Calendar) IsLeapYear(year int) bool {
	return isValidYear(year) && c.isLeapYear(year)
}

// isLeapYear reports the leap status of a year from the leap-year rule alone, without the isValidYear range check the
// public IsLeapYear applies. The internal date math (leapYearsSince and yearToDaysWith) needs the true leap status of
// every year Date.Year's binary search probes, and that search ranges past the public [math.MinInt32, math.MaxInt32]
// limits for a date whose year sits near them; treating those out-of-range probes as non-leap would undercount a
// year's length and let the search settle on the wrong year.
func (c *Calendar) isLeapYear(year int) bool {
	cfg := c.config()
	if cfg.LeapYear == nil {
		return false
	}
	if year < 1 {
		year++ // account for gap, since there is no year 0
	}
	if year%cfg.LeapYear.Every != 0 {
		return false
	}
	if cfg.LeapYear.Except == 0 {
		return true
	}
	if year%cfg.LeapYear.Except != 0 {
		return true
	}
	if cfg.LeapYear.Unless == 0 {
		return false
	}
	return year%cfg.LeapYear.Unless == 0
}

// IsLeapMonth returns true if the month is the leap month.
func (c *Calendar) IsLeapMonth(month int) bool {
	return c.config().isLeapMonth(month)
}

// leapYearsSince returns the number of leap years that have occurred between year 1 and the specified year, exclusive.
// It counts the true leap years for any year, including those outside the public valid range, because Date.Year's
// search probes years beyond it (see isLeapYear).
func (c *Calendar) leapYearsSince(year int) int {
	if c.config().LeapYear == nil {
		return 0
	}
	if year >= 1 {
		return c.countLeaps(year - 1)
	}
	// There is no year 0, so the years strictly between year and 1 run year+1..-1. isLeapYear derives a negative
	// year's leap status from the magnitude |year+1|, so those years map to magnitudes 0..(-year-2). countLeaps()
	// covers magnitudes 1 and up; magnitude 0 (year -1) is added on separately because whether it is a leap year
	// depends on the Except/Unless rule, which is exactly what isLeapYear(-1) reports.
	upper := -year - 2
	if upper < 0 {
		return 0 // year == -1: nothing lies strictly between it and year 1
	}
	count := c.countLeaps(upper)
	if c.isLeapYear(-1) {
		count++
	}
	return count
}

// countLeaps returns the number of leap years whose magnitude (distance from the leap pattern's origin) is 1 through n
// inclusive. The leap pattern is symmetric about the origin, so the same closed form serves positive years directly and
// negative years via their shifted magnitude. n must not be negative. Config.Valid() guarantees every multiple of
// Except is a multiple of Every and every multiple of Unless is a multiple of Except, so dividing counts each tier
// independently.
func (c *Calendar) countLeaps(n int) int {
	cfg := c.config()
	count := n / cfg.LeapYear.Every
	if cfg.LeapYear.Except != 0 {
		count -= n / cfg.LeapYear.Except
		if cfg.LeapYear.Unless != 0 {
			count += n / cfg.LeapYear.Unless
		}
	}
	return count
}

// Text writes a text representation of the year.
func (c *Calendar) Text(year int, w io.Writer) {
	cfg := c.config()
	date := c.MustNewDate(1, 1, year)
	date.WriteFormat(w, "Year %Y\n")
	width := widthNeeded(c.mostDaysInMonth())
	maximum := len(cfg.Months)
	for i := 1; i <= maximum; i++ {
		fmt.Fprintln(w)
		c.MustNewDate(i, 1, year).textCalendarMonth(w, width)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Seasons:")
	width = 0
	for i := range cfg.Seasons {
		width = max(utf8.RuneCountInString(cfg.Seasons[i].Name), width)
	}
	for i := range cfg.Seasons {
		fmt.Fprintf(w, "  %-[1]*s (%s)\n", width, cfg.Seasons[i].Name, cfg.Seasons[i].DateRange())
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Week Days:")
	for i, weekday := range cfg.WeekDays {
		fmt.Fprintf(w, "  %[1]*d: (%s) %s\n", widthNeeded(len(cfg.WeekDays)), i+1, xstrings.FirstN(weekday, 1), weekday)
	}
}
