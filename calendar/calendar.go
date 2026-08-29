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
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

var (
	defaultCalendarLock sync.RWMutex
	defaultCalendar     *Calendar
)

// Assigned in init rather than at declaration: newCalendar's helpers reference Default, so a package-level initializer
// would form an initialization cycle.
func init() {
	defaultCalendar = Gregorian()
}

// abbreviatedNameLength is the number of leading characters in an abbreviated month or weekday name. Both the %m and %w
// directives and monthFromText use it, so an emitted abbreviation parses back to the same month.
const abbreviatedNameLength = 3

// Calendar holds the data for a calendar.
type Calendar struct {
	cfg            *Config
	numericDate    *regexp.Regexp // "9/22/2017" or "9/22/2017 AD"; see Config.dateRegexes
	namedDate      *regexp.Regexp // "September 22, 2017 AD", "September 22, 2017", "Sep 22, 2017 AD", or "Sep 22, 2017"
	minDaysPerYear int            // sum of every month's Days
	firstDay       int64          // Days of the earliest representable Date; see Date.Add
	lastDay        int64          // Days of the latest representable Date; see Date.Add
}

// New creates a new Calendar from the given Config. An error is returned if the Config is not Valid.
func New(cfg *Config) (*Calendar, error) {
	if err := cfg.Valid(); err != nil {
		return nil, err
	}
	return newCalendar(cfg.Clone()), nil
}

// newCalendar wraps an already-validated Config, which it takes as-is without cloning, and precomputes the values that
// depend only on it.
func newCalendar(cfg *Config) *Calendar {
	c := &Calendar{cfg: cfg}
	for i := range cfg.Months {
		c.minDaysPerYear += cfg.Months[i].Days
	}
	// Dates are confined to the years IsValidYear accepts; maxDaysPerYear keeps that span within an int64.
	c.firstDay = c.yearToDays(math.MinInt32)
	c.lastDay = c.yearToDays(math.MaxInt32) + int64(c.Days(math.MaxInt32)) - 1
	c.numericDate, c.namedDate = cfg.dateRegexes()
	return c
}

// dateRegexes compiles the two patterns ParseDate recognizes, built from the Config's own month and era names (quoted
// literally, matched case-insensitively, longest first) so that names with spaces or punctuation parse back whole. The
// month group (1) also accepts each name's abbreviatedNameLength prefix, which monthFromText resolves. The era group
// (4) must be followed by the end of the text or a non-alphanumeric character so "BC" cannot match inside "BCE"; with
// no eras the group is still present, so a match always has the same shape. maxNamesLength keeps the patterns well
// inside regexp's size limit, the only way compiling them could fail, so MustCompile is safe.
func (c *Config) dateRegexes() (numeric, named *regexp.Regexp) {
	months := make([]string, 0, len(c.Months)*2)
	for i := range c.Months {
		months = append(months, c.Months[i].Name, xstrings.FirstN(c.Months[i].Name, abbreviatedNameLength))
	}
	era := "()"
	if eras := alternation(c.Era, c.PreviousEra); eras != "" {
		era = "(?: *(" + eras + ")(?:$|[^\\pL\\pN]))?"
	}
	numeric = regexp.MustCompile("(?i)([[:digit:]]+)/([[:digit:]]+)/(-?[[:digit:]]+)" + era)
	named = regexp.MustCompile("(?i)(" + alternation(months...) + ") *([[:digit:]]+), *(-?[[:digit:]]+)" + era)
	return numeric, named
}

// alternation joins the distinct, non-empty names into a regular expression alternation of literals, longest first, or
// "" when there are none.
func alternation(names ...string) string {
	quoted := make([]string, 0, len(names))
	for _, name := range names {
		if name != "" {
			quoted = append(quoted, regexp.QuoteMeta(name))
		}
	}
	slices.SortStableFunc(quoted, func(a, b string) int { return len(b) - len(a) })
	return strings.Join(slices.Compact(quoted), "|")
}

// Default returns the default Calendar that will be used if one isn't explicitly used (for example, if you create a
// Date directly via Date{} rather than via a Calendar).
func Default() *Calendar {
	defaultCalendarLock.RLock()
	defer defaultCalendarLock.RUnlock()
	return defaultCalendar
}

// SetDefault sets the default Calendar to use. An error is returned, and the default left unchanged, if cal is nil or a
// zero-value Calendar rather than one obtained from New or one of the built-in constructors.
func SetDefault(cal *Calendar) error {
	if cal == nil || cal.cfg == nil {
		return errs.New("calendar must be obtained from New or one of the built-in constructors")
	}
	defaultCalendarLock.Lock()
	defaultCalendar = cal
	defaultCalendarLock.Unlock()
	return nil
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

// NewDate creates a new date from the specified month, day and year. The year must be one IsValidYear accepts.
func (c *Calendar) NewDate(month, day, year int) (Date, error) {
	if !IsValidYear(year) {
		return Date{cal: c}, errs.Newf("year %d is invalid; must be in the range %d to %d, not including 0", year,
			math.MinInt32, math.MaxInt32)
	}
	cfg := c.config()
	if month < 1 || month > len(cfg.Months) {
		return Date{cal: c}, errs.Newf("month %d is invalid; must be in the range 1 to %d", month, len(cfg.Months))
	}
	monthDays := cfg.Months[month-1].Days
	if c.IsLeapMonth(month) && c.IsLeapYear(year) {
		monthDays++
	}
	if day < 1 || day > monthDays {
		return Date{cal: c}, errs.Newf("day %d is invalid; must be in the range 1 to %d for month %d", day, monthDays,
			month)
	}
	days := c.yearToDays(int64(year)) + int64(day) - 1
	for i := 1; i < month; i++ {
		days += int64(cfg.Months[i-1].Days)
	}
	if c.IsLeapYear(year) && cfg.LeapYear.Month < month {
		days++
	}
	return c.NewDateByDays(days), nil
}

// IsValidYear returns true if the year is one that may be used: any year within the int32 range except 0, since there
// is no year 0 (year -1 immediately precedes year 1).
func IsValidYear(year int) bool {
	return year != 0 && year >= math.MinInt32 && year <= math.MaxInt32
}

// NewDateByDays creates a new date from a number of days, with 0 representing the date 1/1/1. The result saturates at
// the calendar's earliest and latest representable days, as Date.Add does.
func (c *Calendar) NewDateByDays(days int64) Date {
	d := Date{cal: c}
	return d.Add(days)
}

func (c *Calendar) yearToDays(year int64) int64 {
	return c.yearToDaysWith(year, int64(c.MinDaysPerYear()))
}

func (c *Calendar) yearToDaysWith(year, minDaysPerYear int64) int64 {
	var days int64
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

// ParseDate creates a new date from the specified text, which may be embedded in surrounding prose. Two forms are
// recognized, each optionally followed by one of the calendar's era names: the numeric ShortFormat ("9/22/2017 AD") and
// the named LongFormat or MediumFormat ("September 22, 2017", "Sep 22, 2017 AD"). Names are matched without regard to
// case.
func (c *Calendar) ParseDate(in string) (Date, error) {
	numeric, named := c.dateRegexes()
	if parts := numeric.FindStringSubmatch(in); parts != nil {
		month, err := strconv.Atoi(parts[1])
		if err != nil {
			return Date{cal: c}, errs.NewWithCausef(err, "invalid month text %q", parts[1])
		}
		return c.parseDate(month, parts[2], parts[3], parts[4])
	}
	if parts := named.FindStringSubmatch(in); parts != nil {
		month, err := c.monthFromText(parts[1])
		if err != nil {
			return Date{cal: c}, err
		}
		return c.parseDate(month, parts[2], parts[3], parts[4])
	}
	return Date{cal: c}, errs.Newf("invalid date text %q", in)
}

func (c *Calendar) parseDate(month int, dayText, yearText, eraText string) (Date, error) {
	year, err := strconv.ParseInt(yearText, 10, 64)
	if err != nil {
		return Date{cal: c}, errs.NewWithCausef(err, "invalid year text %q", yearText)
	}
	day, err := strconv.Atoi(dayText)
	if err != nil {
		return Date{cal: c}, errs.NewWithCausef(err, "invalid day text %q", dayText)
	}
	// Apply the era before narrowing to int: "2147483648 BC" only fits once its sign is restored.
	if year, err = c.resolveEraSuffix(year, yearText, eraText); err != nil {
		return Date{cal: c}, err
	}
	if year < math.MinInt32 || year > math.MaxInt32 {
		return Date{cal: c}, errs.Newf("year %d is invalid; must be in the range %d to %d", year, math.MinInt32,
			math.MaxInt32)
	}
	return c.NewDate(month, day, int(year))
}

// monthFromText resolves a month name to its 1-based index. A full-name match is preferred; failing that, an
// abbreviation is accepted only when it identifies a single month.
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

// eraForYear maps a signed year to the year and era label displayed for it; resolveEraSuffix is its inverse. A negative
// year belongs to the previous era. When the eras are distinct the label carries the sign, so the magnitude is returned
// (-5 with "AD"/"BC" yields 5, "BC"); otherwise the signed year is returned unchanged (-5 with "AR"/"AR" yields -5,
// "AR"). The result is an int64 because the magnitude of math.MinInt32 does not fit a 32-bit int.
func (c *Calendar) eraForYear(year int) (displayYear int64, era string) {
	cfg := c.config()
	era = cfg.Era
	if year < 0 {
		era = cfg.PreviousEra
	}
	displayYear = int64(year)
	if year < 0 && era != "" && cfg.distinctEras() {
		displayYear = -displayYear
	}
	return displayYear, era
}

// distinctEras reports whether the two era labels differ, compared without regard to case since that is how ParseDate
// matches them. Valid rejects labels that differ only in case, so this agrees with a plain comparison for any Config it
// accepts.
func (c *Config) distinctEras() bool {
	return !strings.EqualFold(c.Era, c.PreviousEra)
}

// resolveEraSuffix folds a recognized era suffix into the sign of a parsed year, the inverse of eraForYear. With
// distinct eras, a previous-era suffix negates a non-negative year, while either suffix on a negative year is rejected
// as redundant or contradictory. An empty or unrecognized suffix is ignored so dates embedded in prose still parse.
// yearText and eraText appear only in error messages.
func (c *Calendar) resolveEraSuffix(year int64, yearText, eraText string) (int64, error) {
	cfg := c.config()
	distinctEras := cfg.distinctEras()
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

// dateRegexes returns the ParseDate patterns, falling back to the default calendar's as config does.
func (c *Calendar) dateRegexes() (numeric, named *regexp.Regexp) {
	if c != nil && c.cfg != nil {
		return c.numericDate, c.namedDate
	}
	def := Default()
	return def.numericDate, def.namedDate
}

// MinDaysPerYear returns the minimum number of days in a year.
func (c *Calendar) MinDaysPerYear() int {
	if c != nil && c.cfg != nil {
		return c.minDaysPerYear
	}
	return Default().minDaysPerYear
}

// maxDaysInMonth returns the most days the given 1-based month can hold, including the leap day.
func (c *Calendar) maxDaysInMonth(month int) int {
	return c.config().maxDaysInMonth(month)
}

// mostDaysInMonth returns the most days any month can hold, including the leap day, for sizing day-of-month fields.
func (c *Calendar) mostDaysInMonth() int {
	cfg := c.config()
	most := 0
	for i := range cfg.Months {
		most = max(most, cfg.maxDaysInMonth(i+1))
	}
	return most
}

// Days returns the number of days in a specific year, or 0 for a year IsValidYear rejects.
func (c *Calendar) Days(year int) int {
	if !IsValidYear(year) {
		return 0
	}
	days := c.MinDaysPerYear()
	if c.isLeapYear(int64(year)) {
		days++
	}
	return days
}

// IsLeapYear returns true if the year is a leap year. A year IsValidYear rejects is never a leap year.
func (c *Calendar) IsLeapYear(year int) bool {
	return IsValidYear(year) && c.isLeapYear(int64(year))
}

// isLeapYear applies the leap-year rule to any int64 year, without range checking: Date.Year's search probes years past
// the int32 limits and needs their true leap status.
func (c *Calendar) isLeapYear(year int64) bool {
	cfg := c.config()
	if cfg.LeapYear == nil {
		return false
	}
	if year < 1 {
		year++ // account for gap, since there is no year 0
	}
	if year%int64(cfg.LeapYear.Every) != 0 {
		return false
	}
	if cfg.LeapYear.Except == 0 {
		return true
	}
	if year%int64(cfg.LeapYear.Except) != 0 {
		return true
	}
	if cfg.LeapYear.Unless == 0 {
		return false
	}
	return year%int64(cfg.LeapYear.Unless) == 0
}

// IsLeapMonth returns true if the month is the leap month.
func (c *Calendar) IsLeapMonth(month int) bool {
	return c.config().isLeapMonth(month)
}

// leapYearsSince returns the number of leap years strictly between year 1 and the given year, for any int64 year. The
// calendar must have a leap rule.
func (c *Calendar) leapYearsSince(year int64) int64 {
	if year >= 1 {
		return c.countLeaps(year - 1)
	}
	// There is no year 0, so the years strictly between year and 1 are year+1..-1, which isLeapYear treats as
	// magnitudes 0..(-year-2). countLeaps covers magnitudes 1 and up; magnitude 0 (year -1) depends on the
	// Except/Unless rule, so it is added separately.
	upper := -year - 2
	if upper < 0 {
		return 0 // year == -1
	}
	count := c.countLeaps(upper)
	if c.isLeapYear(-1) {
		count++
	}
	return count
}

// countLeaps returns the number of leap years with magnitude 1 through n inclusive; n must not be negative. Valid
// guarantees Unless is a multiple of Except, which is a multiple of Every, so each tier can be counted by division.
func (c *Calendar) countLeaps(n int64) int64 {
	cfg := c.config()
	count := n / int64(cfg.LeapYear.Every)
	if cfg.LeapYear.Except != 0 {
		count -= n / int64(cfg.LeapYear.Except)
		if cfg.LeapYear.Unless != 0 {
			count += n / int64(cfg.LeapYear.Unless)
		}
	}
	return count
}

// Text writes a text representation of the year. An error is returned, and nothing is written, if the year is not one
// NewDate accepts. Write errors from w are not reported.
func (c *Calendar) Text(year int, w io.Writer) error {
	date, err := c.NewDate(1, 1, year)
	if err != nil {
		return err
	}
	cfg := c.config()
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
	return nil
}
