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

// The default calendar is assigned in init rather than in its declaration: newCalendar computes each calendar's day
// span through methods that fall back to Default for a nil Calendar, so a package-level initializer reaching Default
// through Gregorian would form an initialization cycle.
func init() {
	defaultCalendar = Gregorian()
}

// abbreviatedNameLength is the number of leading characters used to abbreviate a month or weekday name. The %m and %w
// format directives emit this many characters, and monthFromText accepts a month abbreviation of this length, so an
// emitted short month name parses back to the same month (e.g. MediumFormat round-trips through ParseDate). Keeping
// both sides driven by this single constant prevents the emit and parse widths from silently drifting apart.
const abbreviatedNameLength = 3

// Calendar holds the data for a calendar.
type Calendar struct {
	cfg            *Config
	numericDate    *regexp.Regexp // "9/22/2017" or "9/22/2017 AD"; see Config.dateRegexes
	namedDate      *regexp.Regexp // "September 22, 2017 AD", "September 22, 2017", "Sep 22, 2017 AD", or "Sep 22, 2017"
	minDaysPerYear int            // sum of every month's Days; a pure function of the immutable cfg, cached at construction
	firstDay       int64          // Days of the earliest Date the calendar can represent; see Date.Add
	lastDay        int64          // Days of the latest Date the calendar can represent; see Date.Add
}

// New creates a new Calendar from the given Config. An error is returned if the Config is not Valid.
func New(cfg *Config) (*Calendar, error) {
	if err := cfg.Valid(); err != nil {
		return nil, err
	}
	return newCalendar(cfg.Clone()), nil
}

// newCalendar wraps an already-validated (or built-in) Config, precomputing minDaysPerYear so Year and the date
// accessors that lean on it do not re-sum every month on each call, the span of days a Date on the calendar may occupy,
// and the ParseDate patterns that depend on the Config's month and era names. The cfg is taken as-is and not cloned
// again; callers pass a Config they own (New clones first, the built-ins pass a fresh literal).
func newCalendar(cfg *Config) *Calendar {
	c := &Calendar{cfg: cfg}
	for i := range cfg.Months {
		c.minDaysPerYear += cfg.Months[i].Days
	}
	// A Date is confined to the years IsValidYear accepts so that Date.Year stays within the int32 range on every
	// target; maxDaysPerYear keeps that whole span within an int64 on every calendar.
	c.firstDay = c.yearToDays(math.MinInt32)
	c.lastDay = c.yearToDays(math.MaxInt32) + int64(c.Days(math.MaxInt32)) - 1
	c.numericDate, c.namedDate = cfg.dateRegexes()
	return c
}

// dateRegexes compiles the two patterns ParseDate recognizes, tailored to the Config's own month and era names so that
// whatever Date.Format emits parses back. A generic word pattern cannot do that: Valid permits names containing spaces,
// punctuation and non-ASCII letters (a month "New Moon", an era "A.D."), and a generic pattern captures only a fragment
// of such a name, rejecting the month or, worse, silently dropping the era and with it the year's sign. Every name is
// quoted literally and matched case-insensitively. The month group (1) also accepts each name's abbreviatedNameLength
// prefix, which is what %m emits, leaving monthFromText to resolve it and to reject an ambiguous one. An era (group 4)
// must be followed by the end of the text or a non-alphanumeric character so a label cannot match the start of a longer
// word ("BC" inside "BCE"); when the Config names no era the group is present but can never match, so the parts a match
// yields always have the same shape. Alternatives are ordered longest first so a name that is a prefix of another is
// never preferred over the longer match. Every name is quoted literally and the patterns have no nesting, so the only
// way compilation could fail is by exceeding the size a regular expression may have, and maxNamesLength keeps the
// patterns for any Config Valid accepts far inside that (the built-in Configs are smaller still), which is what lets
// them be compiled with MustCompile.
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

// alternation joins the distinct, non-empty names into a regular expression alternation of literal matches, longest
// first. It returns "" when there is nothing to match.
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
// zero-value Calendar rather than one obtained from New or one of the built-in constructors, since such a Calendar
// could not serve as the fallback for a zero-value Date.
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

// NewDate creates a new date from the specified month, day and year. The year must be one IsValidYear accepts; every
// such year is representable in full on every calendar (see maxDaysPerYear).
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
// is no year 0 (year -1 immediately precedes year 1). Confining years to that range keeps every Date's year within a
// 32-bit int on any target, and bounds the day arithmetic (see maxDaysPerYear).
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
// recognized, each optionally followed by one of the calendar's era names: the numeric ShortFormat ("9/22/2017",
// "9/22/2017 AD") and the named LongFormat or MediumFormat ("September 22, 2017", "Sep 22, 2017 AD"), where the month
// is any of the calendar's month names or its abbreviation as %m emits it, matched without regard to case.
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
	// The era is applied, and the range checked, before the year is narrowed to an int, which may be only 32 bits wide:
	// year math.MinInt32 renders as "2147483648 BC", a magnitude that only fits once the previous-era suffix has
	// restored its sign.
	if year, err = c.resolveEraSuffix(year, yearText, eraText); err != nil {
		return Date{cal: c}, err
	}
	if year < math.MinInt32 || year > math.MaxInt32 {
		return Date{cal: c}, errs.Newf("year %d is invalid; must be in the range %d to %d", year, math.MinInt32,
			math.MaxInt32)
	}
	return c.NewDate(month, day, int(year))
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
// the sign, so the signed year is returned unchanged (-5 with eras "AR"/"AR" yields -5, "AR"). The display year is an
// int64 because the magnitude of math.MinInt32 does not fit a 32-bit int.
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

// distinctEras reports whether the calendar labels its two eras differently, so that the label can carry the sign of a
// year (see eraForYear). The labels are compared without regard to case because that is how ParseDate matches an era
// suffix: two labels that differ only in case read as distinct when formatting but cannot be told apart when parsing,
// so a formatted previous-era year would parse back into the current era. Valid rejects such a pair, so for any Config
// it accepts this agrees with a plain comparison, and this single definition keeps every user of the era model -- the
// format directives, Date.Era and resolveEraSuffix -- on the same side of that line.
func (c *Config) distinctEras() bool {
	return !strings.EqualFold(c.Era, c.PreviousEra)
}

// resolveEraSuffix folds a recognized era suffix into the sign of a parsed year, the parse-side inverse of eraForYear.
// A leading minus sign already places the year before the current era, so a recognized suffix must agree with it. When
// the calendar names its two eras distinctly, a previous-era suffix on a non-negative year selects the previous era,
// but on a negative year it merely repeats the sign, and a current-era suffix on a negative year flatly contradicts it;
// reject both rather than silently choosing an interpretation. An empty or unrecognized suffix is left alone so
// ParseDate can still find dates embedded in surrounding text. yearText and eraText are the original matched text,
// used only for the error messages. The year is an int64 so that a previous-era magnitude of 2147483648 can be
// negated into math.MinInt32 before parseDate narrows it.
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

// dateRegexes returns the ParseDate patterns for this calendar, falling back to the default calendar's for a nil or
// zero-value Calendar, as config does.
func (c *Calendar) dateRegexes() (numeric, named *regexp.Regexp) {
	if c != nil && c.cfg != nil {
		return c.numericDate, c.namedDate
	}
	def := Default()
	return def.numericDate, def.namedDate
}

// MinDaysPerYear returns the minimum number of days in a year. The sum is computed once when the Calendar is built (see
// newCalendar) and cached, since it is a pure function of the immutable Config yet is consulted on every Year lookup
// and date resolution.
func (c *Calendar) MinDaysPerYear() int {
	if c != nil && c.cfg != nil {
		return c.minDaysPerYear
	}
	return Default().minDaysPerYear
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

// Days returns the number of days contained in a specific year, or 0 for a year IsValidYear rejects, since no such year
// exists (IsLeapYear likewise reports false for it). The length follows the leap-year rule exactly as the internal date
// math treats it, so the result always agrees with the distance between consecutive first-of-year dates.
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

// IsLeapYear returns true if the year is a leap year. Note that valid years are constrained to not 0 and in the range
// math.MinInt32 to math.MaxInt32, so an invalid year will always return false.
func (c *Calendar) IsLeapYear(year int) bool {
	return IsValidYear(year) && c.isLeapYear(int64(year))
}

// isLeapYear reports the leap status of a year from the leap-year rule alone, for any int64 year. The internal date
// math (leapYearsSince and yearToDaysWith) needs the true leap status of every year Date.Year's binary search probes,
// and that search ranges past the int32 limits for a date whose year sits near them; treating those out-of-range
// probes as non-leap would undercount a year's length and let the search settle on the wrong year.
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

// leapYearsSince returns the number of leap years that have occurred between year 1 and the specified year, exclusive.
// It counts the true leap years for any int64 year, including those outside the int32 range, because Date.Year's
// search probes years beyond it (see isLeapYear). The calendar must have a leap rule: yearToDaysWith, the only caller,
// checks that before calling, and countLeaps relies on it.
func (c *Calendar) leapYearsSince(year int64) int64 {
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
// NewDate accepts. Write errors from w are not reported, matching WriteFormat and TextCalendarMonth.
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
