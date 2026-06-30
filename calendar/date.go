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

// DaysLimit is the limit to how many days a Date can represent on either side of the 1/1/1 date of a Calendar.
const DaysLimit = 1 << 61

// Date holds a calendar date. If the date was not initialized (i.e. not obtained from a Calendar), the default calendar
// will be used when doing any operation that needs to know which calendar it was a part of.
type Date struct {
	cal  *Calendar
	days int
}

// calendar returns the calendar associated with the date, falling back to Default for a zero-value Date that was never
// associated with a calendar.
func (date Date) calendar() *Calendar {
	if date.cal == nil || date.cal.cfg == nil {
		return Default()
	}
	return date.cal
}

// Add delta days to the date and return a new Date, saturating to [-DaysLimit, DaysLimit] on overflow or when the sum
// exceeds the range.
func (date Date) Add(delta int) Date {
	days := date.days + delta
	if days > DaysLimit || (date.days > 0 && delta > 0 && days < 0) {
		days = DaysLimit
	} else if days < -DaysLimit || (date.days < 0 && delta < 0 && days > 0) {
		days = -DaysLimit
	}
	return Date{
		cal:  date.calendar(),
		days: days,
	}
}

// Days is the number of days since 1/1/1 in the calendar. Note that the value -1 refers to the last day of the year -1,
// not year 0, as there is no year 0.
func (date Date) Days() int {
	return date.days
}

// Year returns the year of the date.
func (date Date) Year() int {
	cal := date.calendar()
	minDays := cal.MinDaysPerYear()
	lo, hi := 1, date.days/minDays+1
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
	return lo
}

// resolve returns the year, month (1-based), day within the month (1-based), and the number of days in that month from
// a single Year computation and a single walk over the months. The individual accessors delegate here so they do not
// each recompute the relatively expensive Year.
func (date Date) resolve() (year, month, dayInMonth, daysInMonth int) {
	cal := date.calendar()
	cfg := cal.config()
	year = date.Year()
	isLeapYear := cal.IsLeapYear(year)
	days := 1 + date.days - cal.yearToDays(year)
	for i := range cfg.Months {
		amt := cfg.Months[i].Days
		if isLeapYear && cal.IsLeapMonth(i+1) {
			amt++
		}
		if days <= amt {
			return year, i + 1, days, amt
		}
		days -= amt
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
	return 1 + date.days - date.calendar().yearToDays(date.Year())
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
	weekday := date.days % len(cfg.WeekDays)
	if date.days < 0 {
		weekday += len(cfg.WeekDays)
	}
	return (weekday + cfg.DayZeroWeekDay) % len(cfg.WeekDays)
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
		endDay := cfg.Seasons[i].EndDay
		if cfg.Seasons[i].EndMonth >= 1 && cfg.Seasons[i].EndMonth <= len(cfg.Months) &&
			endDay == cfg.Months[cfg.Seasons[i].EndMonth-1].Days {
			endDay = cal.maxDaysInMonth(cfg.Seasons[i].EndMonth)
		}
		afterStart := onOrAfter(month, dayInMonth, cfg.Seasons[i].StartMonth, cfg.Seasons[i].StartDay)
		beforeEnd := onOrAfter(cfg.Seasons[i].EndMonth, endDay, month, dayInMonth)
		if onOrAfter(cfg.Seasons[i].EndMonth, endDay, cfg.Seasons[i].StartMonth, cfg.Seasons[i].StartDay) {
			if afterStart && beforeEnd { // start on or before end: a single contiguous span
				return cfg.Seasons[i], true
			}
		} else if afterStart || beforeEnd { // start after end: the span wraps the year boundary
			return cfg.Seasons[i], true
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
// that is not a directive is passed through unchanged. Valid directives:
//
//	%W  Full weekday, e.g. 'Friday'
//	%w  Short weekday, e.g. 'Fri'
//	%M  Full month name, e.g. 'September'
//	%m  Short month name, e.g. 'Sep'
//	%N  Month, e.g. '9'
//	%n  Month padded with zeroes, e.g. '09'
//	%D  Day, e.g. '2'
//	%d  Day padded with zeroes, e.g. '02'
//	%Y  Year, e.g. '2017' if positive, '2017 BC' if negative; however, if the eras aren't empty and match each other,
//	    then this will behave the same as %y
//	%y  Year with era, e.g. '2017 AD'; however, if the eras are empty or they match each other, then negative years
//	    will result in '-2017 AD'
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
				if era != "" && era == cfg.PreviousEra {
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
			}
		case r == '%':
			cmd = true
		default:
			fmt.Fprintf(w, "%c", r)
		}
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
	year, month, _, maximum := date.resolve()
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
	firstDay := cal.MustNewDate(month, 1, year)
	for i := 1; i <= maximum; i++ {
		weekDay := firstDay.Add(i - 1).WeekDay()
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
