// Copyright (c) 2017-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calendar

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/richardwilkes/toolbox/v2/xstrings"
)

// Predefined formats.
const (
	FullFormat   = "%W, %M %D, %Y"
	LongFormat   = "%M %D, %Y"
	MediumFormat = "%m %D, %Y"
	ShortFormat  = "%N/%D/%Y"
)

// Date holds a calendar date. If the date was not initialized (i.e. not obtained from a Calendar), the default calendar
// will be used when doing any operation that needs to know which calendar it was a part of. Day counts (Days, Add and
// NewDateByDays) are int64; everything else is an int, kept within 32 bits on every target by IsValidYear for years
// and by the caps Config.Valid enforces for months, days and weekdays.
//
// Compare Dates with Equal rather than ==. Date is a comparable struct, but == also compares the calendar reference a
// Date carries, and a zero Date{} carries none while the same day obtained from Default() (or from Add on a zero Date)
// carries a reference to the default calendar, so == reports the two as different even though both represent 1/1/1 on
// the default calendar and format identically.
type Date struct {
	cal  *Calendar
	days int64
}

// calendar returns the calendar associated with the date, falling back to Default for a zero-value Date that was never
// associated with a calendar.
func (date Date) calendar() *Calendar {
	if date.cal == nil || date.cal.cfg == nil {
		return Default()
	}
	return date.cal
}

// Add delta days to the date and return a new Date. The result saturates at the calendar's earliest and latest
// representable days -- the first day of year math.MinInt32 and the last day of year math.MaxInt32 -- on overflow or
// when the sum would fall outside that span. Confining every Date to the years IsValidYear accepts is what keeps Year
// within the int32 range on every target.
func (date Date) Add(delta int64) Date {
	cal := date.calendar()
	days := date.days + delta
	// Test for the sum having wrapped around int64 before comparing its magnitude: a wrapped sum lands on the far side
	// of zero, so checking it against the limits first would saturate toward the wrong one.
	switch {
	case delta > 0 && days < date.days: // wrapped past math.MaxInt64
		days = cal.lastDay
	case delta < 0 && days > date.days: // wrapped past math.MinInt64
		days = cal.firstDay
	case days > cal.lastDay:
		days = cal.lastDay
	case days < cal.firstDay:
		days = cal.firstDay
	}
	return Date{
		cal:  cal,
		days: days,
	}
}

// Days is the number of days since 1/1/1 in the calendar. Note that the value -1 refers to the last day of the year -1,
// not year 0, as there is no year 0.
func (date Date) Days() int64 {
	return date.days
}

// Equal reports whether the two dates represent the same day on the same calendar. Unlike ==, a zero Date{} compares
// equal to the same day obtained from Default(), since both resolve to the default calendar. Dates on different
// calendars are never equal, even when their Days match.
func (date Date) Equal(other Date) bool {
	return date.days == other.days && date.calendar() == other.calendar()
}

// Year returns the year of the date. Add confines every Date to the years IsValidYear accepts, so the year always lies
// within the int32 range, even where int is wider.
func (date Date) Year() int {
	cal := date.calendar()
	minDays := int64(cal.MinDaysPerYear())
	// The search works in int64 because its upper bound can lie past math.MaxInt32 for a date near the end of the span
	// even though the year it settles on never does.
	lo, hi := int64(1), date.days/minDays+1
	if date.days < 0 {
		lo, hi = date.days/minDays-1, -1
	}
	for lo < hi {
		mid := lo + (hi-lo+1)/2 // bias toward hi so lo still advances when only one candidate separates them
		if cal.yearToDaysWith(mid, minDays) <= date.days {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return int(lo)
}

// resolve returns the year, month (1-based), day within the month (1-based), and the number of days in that month from
// a single Year computation and a single walk over the months. The individual accessors delegate here so they do not
// each recompute the relatively expensive Year.
func (date Date) resolve() (year, month, dayInMonth, daysInMonth int) {
	cal := date.calendar()
	cfg := cal.config()
	year = date.Year()
	isLeapYear := cal.IsLeapYear(year)
	days := 1 + date.days - cal.yearToDays(int64(year)) // the day within the year, which maxDaysPerYear keeps in an int
	for i := range cfg.Months {
		amt := cfg.Months[i].Days
		if isLeapYear && cal.IsLeapMonth(i+1) {
			amt++
		}
		if days <= int64(amt) {
			return year, i + 1, int(days), amt
		}
		days -= int64(amt)
	}
	// If this is reached, the algorithm is wrong.
	panic("unable to determine month") // @allow
}

// Month returns the month of the date. Note that the first month is represented by 1, not 0.
func (date Date) Month() int {
	_, month, _, _ := date.resolve()
	return month
}

// MonthName returns the name of the month of the date.
func (date Date) MonthName() string {
	return date.calendar().config().Months[date.Month()-1].Name
}

// DayInYear returns the day within the year of the date. Note that the first day is represented by a 1, not 0.
func (date Date) DayInYear() int {
	return int(1 + date.days - date.calendar().yearToDays(int64(date.Year())))
}

// DayInMonth returns the day within the month of the date. Note that the first day is represented by a 1, not 0.
func (date Date) DayInMonth() int {
	_, _, dayInMonth, _ := date.resolve()
	return dayInMonth
}

// DaysInMonth returns the number of days in the month of the date.
func (date Date) DaysInMonth() int {
	_, _, _, daysInMonth := date.resolve()
	return daysInMonth
}

// WeekDay returns the weekday of the date.
func (date Date) WeekDay() int {
	cfg := date.calendar().config()
	weekDays := int64(len(cfg.WeekDays))
	weekday := date.days % weekDays
	if date.days < 0 {
		weekday += weekDays
	}
	return int((weekday + int64(cfg.DayZeroWeekDay)) % weekDays)
}

// WeekDayName returns the name of the weekday of the date.
func (date Date) WeekDayName() string {
	return date.calendar().config().WeekDays[date.WeekDay()]
}

// Season returns the season that contains the date and true, or a zero Season and false when no season covers it. When
// seasons overlap, the first one in declaration order that contains the date is returned. See Season for how a season's
// span (including one that wraps the year boundary) is interpreted.
func (date Date) Season() (Season, bool) {
	_, month, dayInMonth, _ := date.resolve()
	cal := date.calendar()
	cfg := cal.config()
	for i := range cfg.Seasons {
		s := &cfg.Seasons[i]
		// Whether the span wraps the year boundary is decided from the season as configured. The end-of-month
		// extension below is applied only afterwards: a season ending on the last day of the leap month and starting
		// on the leap day itself (1/31 through 1/30 on a 30-day leap month) wraps as configured, but testing the
		// extended end (1/31) against the start would instead read it as a one-day contiguous span.
		wraps := !onOrAfter(s.EndMonth, s.EndDay, s.StartMonth, s.StartDay)
		endDay := s.EndDay
		if s.EndMonth >= 1 && s.EndMonth <= len(cfg.Months) && endDay == cfg.Months[s.EndMonth-1].Days {
			endDay = cal.maxDaysInMonth(s.EndMonth)
		}
		afterStart := onOrAfter(month, dayInMonth, s.StartMonth, s.StartDay)
		beforeEnd := onOrAfter(s.EndMonth, endDay, month, dayInMonth)
		if wraps {
			if afterStart || beforeEnd { // start after end: the span wraps the year boundary
				return *s, true
			}
		} else if afterStart && beforeEnd { // start on or before end: a single contiguous span
			return *s, true
		}
	}
	return Season{}, false
}

func onOrAfter(m1, d1, m2, d2 int) bool {
	if m1 != m2 {
		return m1 > m2
	}
	return d1 >= d2
}

// Era returns the era suffix for the year.
func (date Date) Era() string {
	_, era := date.calendar().eraForYear(date.Year())
	return era
}

// String returns a date in the ShortFormat.
func (date Date) String() string {
	return date.Format(ShortFormat)
}

// MarshalText implements encoding.TextMarshaler.
func (date Date) MarshalText() ([]byte, error) {
	return []byte(date.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (date *Date) UnmarshalText(text []byte) error {
	d, err := date.calendar().ParseDate(string(text))
	if err != nil {
		return err
	}
	*date = d
	return nil
}

// Format returns a formatted version of the date. The layout is parsed as in WriteFormat().
func (date Date) Format(layout string) string {
	var buffer strings.Builder
	date.WriteFormat(&buffer, layout)
	return buffer.String()
}

// WriteFormat writes a formatted version of the date to the writer. The layout is parsed for directives and anything
// that is not a directive is passed through unchanged, including a % that is not followed by one of the directive
// characters below (so "50% off" and a trailing % are emitted as written). Valid directives:
//
//	%W  Full weekday, e.g. 'Friday'
//	%w  Short weekday, e.g. 'Fri'
//	%M  Full month name, e.g. 'September'
//	%m  Short month name, e.g. 'Sep'
//	%N  Month, e.g. '9'
//	%n  Month padded with zeroes, e.g. '09'
//	%D  Day, e.g. '2'
//	%d  Day padded with zeroes, e.g. '02'
//	%Y  Year, labeled only when the label carries the sign: with distinct eras (such as 'AD' and 'BC'), '2017' for a
//	    year in the current era and '2017 BC' for one in the previous era; when the eras match each other, this
//	    behaves the same as %y; when the eras are empty, the same as %z
//	%y  Year with era, e.g. '2017 AD' or '2017 BC'; when the eras match each other the label cannot carry the sign,
//	    so it stays on the year and a negative year results in '-2017 AR'; when the eras are empty there is no label
//	    to emit and the result is the same as %z, e.g. '-2017'
//	%z  Year without the era, e.g. '2017' or '-2017'
//	%%  %
func (date Date) WriteFormat(w io.Writer, layout string) {
	cal := date.calendar()
	cfg := cal.config()
	var year, month, dayInMonth int
	resolved := false
	resolve := func() {
		if !resolved {
			year, month, dayInMonth, _ = date.resolve()
			resolved = true
		}
	}
	cmd := false
	for _, r := range layout {
		switch {
		case cmd:
			cmd = false
			switch r {
			case 'W':
				fmt.Fprint(w, date.WeekDayName())
			case 'w':
				fmt.Fprint(w, xstrings.FirstN(date.WeekDayName(), abbreviatedNameLength))
			case 'M':
				resolve()
				fmt.Fprint(w, cfg.Months[month-1].Name)
			case 'm':
				resolve()
				fmt.Fprint(w, xstrings.FirstN(cfg.Months[month-1].Name, abbreviatedNameLength))
			case 'N':
				resolve()
				fmt.Fprint(w, month)
			case 'n':
				resolve()
				fmt.Fprintf(w, "%0[1]*[2]d", widthNeeded(len(cfg.Months)), month)
			case 'D':
				resolve()
				fmt.Fprint(w, dayInMonth)
			case 'd':
				resolve()
				fmt.Fprintf(w, "%0[1]*[2]d", widthNeeded(cal.mostDaysInMonth()), dayInMonth)
			case 'Y':
				resolve()
				displayYear, era := cal.eraForYear(year)
				if era != "" && (year < 0 || !cfg.distinctEras()) {
					fmt.Fprintf(w, "%d %s", displayYear, era)
				} else {
					fmt.Fprint(w, displayYear)
				}
			case 'y':
				resolve()
				displayYear, era := cal.eraForYear(year)
				if era != "" {
					fmt.Fprintf(w, "%d %s", displayYear, era)
				} else {
					fmt.Fprint(w, displayYear)
				}
			case 'z':
				resolve()
				fmt.Fprint(w, year)
			case '%':
				fmt.Fprint(w, "%")
			default: // not a directive: pass the % and the character that followed it through unchanged
				fmt.Fprintf(w, "%%%c", r)
			}
		case r == '%':
			cmd = true
		default:
			fmt.Fprintf(w, "%c", r)
		}
	}
	if cmd { // a trailing % has nothing to introduce, so it is not a directive either
		fmt.Fprint(w, "%")
	}
}

func widthNeeded(num int) int {
	return len(strconv.Itoa(num))
}

// TextCalendarMonth writes a text representation of the month.
func (date Date) TextCalendarMonth(w io.Writer) {
	date.textCalendarMonth(w, widthNeeded(date.calendar().mostDaysInMonth()))
}

func (date Date) textCalendarMonth(w io.Writer, width int) {
	cal := date.calendar()
	cfg := cal.config()
	_, month, dayInMonth, maximum := date.resolve()
	fmt.Fprintf(w, "%d: %s", month, cfg.Months[month-1].Name)
	lastDayOfWeek := len(cfg.WeekDays) - 1
	for i, weekday := range cfg.WeekDays {
		if i == 0 {
			fmt.Fprint(w, "\n")
		} else {
			fmt.Fprint(w, " ")
		}
		fmt.Fprint(w, strings.Repeat(" ", width-1))
		fmt.Fprint(w, xstrings.FirstN(weekday, 1))
	}
	// Step back to the first of the month from the date itself rather than rebuilding it through NewDate, which would
	// repeat the year lookup; every day of the month lies within the calendar's span, so this cannot saturate.
	firstDay := date.Add(int64(1 - dayInMonth))
	for i := 1; i <= maximum; i++ {
		weekDay := firstDay.Add(int64(i - 1)).WeekDay()
		if i == 1 || weekDay == 0 {
			fmt.Fprint(w, "\n")
		}
		if i == 1 && weekDay != 0 {
			fmt.Fprint(w, strings.Repeat(" ", weekDay*(width+1)))
		}
		fmt.Fprintf(w, "%[1]*d", width, i)
		if weekDay != lastDayOfWeek {
			fmt.Fprint(w, " ")
		}
	}
	fmt.Fprintln(w)
}
