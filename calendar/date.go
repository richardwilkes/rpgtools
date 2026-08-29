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

// Date holds a calendar date. If the date was not obtained from a Calendar, the default calendar is used for any
// operation that needs one. Day counts are int64; every other value is an int that fits within 32 bits on every target.
//
// Compare Dates with Equal rather than ==: a zero Date{} carries no calendar reference while the same day obtained from
// Default() does, so == reports them as different even though both represent 1/1/1 on the default calendar.
type Date struct {
	cal  *Calendar
	days int64
}

// calendar returns the date's calendar, falling back to Default for a zero-value Date.
func (date Date) calendar() *Calendar {
	if date.cal == nil || date.cal.cfg == nil {
		return Default()
	}
	return date.cal
}

// Add delta days to the date and return a new Date. The result saturates at the calendar's earliest and latest
// representable days: the first day of year math.MinInt32 and the last day of year math.MaxInt32.
func (date Date) Add(delta int64) Date {
	cal := date.calendar()
	days := date.days + delta
	// A sum that wrapped around int64 lands on the far side of zero, so test for that before comparing to the limits.
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
// equal to the same day obtained from Default().
func (date Date) Equal(other Date) bool {
	return date.days == other.days && date.calendar() == other.calendar()
}

// Year returns the year of the date, which always lies within the int32 range.
func (date Date) Year() int {
	cal := date.calendar()
	minDays := int64(cal.MinDaysPerYear())
	// The bounds are int64 because hi can lie past math.MaxInt32 even though the result never does.
	lo, hi := int64(1), date.days/minDays+1
	if date.days < 0 {
		lo, hi = date.days/minDays-1, -1
	}
	for lo < hi {
		mid := lo + (hi-lo+1)/2 // biased toward hi so lo always advances
		if cal.yearToDaysWith(mid, minDays) <= date.days {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return int(lo)
}

// resolve returns the year, 1-based month, 1-based day within the month, and the number of days in that month from a
// single Year computation.
func (date Date) resolve() (year, month, dayInMonth, daysInMonth int) {
	cal := date.calendar()
	cfg := cal.config()
	year = date.Year()
	isLeapYear := cal.IsLeapYear(year)
	days := 1 + date.days - cal.yearToDays(int64(year)) // day within the year
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
// seasons overlap, the first in declaration order wins.
func (date Date) Season() (Season, bool) {
	_, month, dayInMonth, _ := date.resolve()
	cal := date.calendar()
	cfg := cal.config()
	for i := range cfg.Seasons {
		s := &cfg.Seasons[i]
		// Decide wrapping before extending the end to the leap day: a season from 1/31 through 1/30 on a 30-day leap
		// month wraps, but with the extended end (1/31) it would read as a one-day span.
		wraps := !onOrAfter(s.EndMonth, s.EndDay, s.StartMonth, s.StartDay)
		endDay := s.EndDay
		if s.EndMonth >= 1 && s.EndMonth <= len(cfg.Months) && endDay == cfg.Months[s.EndMonth-1].Days {
			endDay = cal.maxDaysInMonth(s.EndMonth)
		}
		afterStart := onOrAfter(month, dayInMonth, s.StartMonth, s.StartDay)
		beforeEnd := onOrAfter(s.EndMonth, endDay, month, dayInMonth)
		if wraps {
			if afterStart || beforeEnd {
				return *s, true
			}
		} else if afterStart && beforeEnd {
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

// WriteFormat writes a formatted version of the date to the writer. Anything in the layout that is not a directive is
// passed through unchanged, including a % not followed by a directive character. Valid directives:
//
//	%W  Full weekday, e.g. 'Friday'
//	%w  Short weekday, e.g. 'Fri'
//	%M  Full month name, e.g. 'September'
//	%m  Short month name, e.g. 'Sep'
//	%N  Month, e.g. '9'
//	%n  Month padded with zeroes, e.g. '09'
//	%D  Day, e.g. '2'
//	%d  Day padded with zeroes, e.g. '02'
//	%Y  Year, labeled only when the label carries the sign: with distinct eras, '2017' for the current era and
//	    '2017 BC' for the previous one; with matching eras, the same as %y; with empty eras, the same as %z
//	%y  Year with era, e.g. '2017 AD' or '2017 BC'; with matching eras the sign stays on the year, e.g. '-2017 AR';
//	    with empty eras, the same as %z
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
			default: // not a directive
				fmt.Fprintf(w, "%%%c", r)
			}
		case r == '%':
			cmd = true
		default:
			fmt.Fprintf(w, "%c", r)
		}
	}
	if cmd { // trailing %
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
	// Every day of the month lies within the calendar's span, so this cannot saturate.
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
