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

// Date holds a calendar date. This is the number of days since 1/1/1 in the calendar. Note that the value -1 refers to
// the last day of the year -1, not year 0, as there is no year 0.
type Date struct {
	cal  *Calendar
	Days int
}

// calendar returns the calendar associated with the date, falling back to Default for a zero-value Date that was never
// associated with a calendar. This mirrors the behavior of UnmarshalText.
func (date Date) calendar() *Calendar {
	if date.cal == nil {
		return Default
	}
	return date.cal
}

// Year returns the year of the date.
func (date Date) Year() int {
	cal := date.calendar()
	minDays := cal.MinDaysPerYear()
	estimate := date.Days / minDays
	if date.Days < 0 {
		estimate--
		for date.Days >= cal.yearToDaysWith(estimate+1, minDays) {
			estimate++
		}
	} else {
		estimate++
		for date.Days < cal.yearToDaysWith(estimate, minDays) {
			estimate--
		}
	}
	return estimate
}

// resolve returns the year, month (1-based), day within the month (1-based), and the number of days in that month from
// a single Year computation and a single walk over the months. The individual accessors delegate here so they do not
// each recompute the relatively expensive Year.
func (date Date) resolve() (year, month, dayInMonth, daysInMonth int) {
	cal := date.calendar()
	year = date.Year()
	isLeapYear := cal.IsLeapYear(year)
	days := 1 + date.Days - cal.yearToDays(year)
	for i := range cal.Months {
		amt := cal.Months[i].Days
		if isLeapYear && cal.IsLeapMonth(i+1) {
			amt++
		}
		if days <= amt {
			return year, i + 1, days, amt
		}
		days -= amt
	}
	// If this is reached, the algorithm is wrong.
	panic("Unable to determine month") // @allow
}

// Month returns the month of the date. Note that the first month is represented by 1, not 0.
func (date Date) Month() int {
	_, month, _, _ := date.resolve()
	return month
}

// MonthName returns the name of the month of the date.
func (date Date) MonthName() string {
	return date.calendar().Months[date.Month()-1].Name
}

// DayInYear returns the day within the year of the date. Note that the first day is represented by a 1, not 0.
func (date Date) DayInYear() int {
	return 1 + date.Days - date.calendar().yearToDays(date.Year())
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
	cal := date.calendar()
	weekday := date.Days % len(cal.WeekDays)
	if date.Days < 0 {
		weekday += len(cal.WeekDays)
	}
	return (weekday + cal.DayZeroWeekDay) % len(cal.WeekDays)
}

// WeekDayName returns the name of the weekday of the date.
func (date Date) WeekDayName() string {
	return date.calendar().WeekDays[date.WeekDay()]
}

// Era returns the era suffix for the year.
func (date Date) Era() string {
	cal := date.calendar()
	if date.Year() < 0 {
		return cal.PreviousEra
	}
	return cal.Era
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
	cmd := false
	for _, r := range layout {
		switch {
		case cmd:
			cmd = false
			switch r {
			case 'W':
				fmt.Fprint(w, date.WeekDayName())
			case 'w':
				fmt.Fprint(w, xstrings.FirstN(date.WeekDayName(), 3))
			case 'M':
				fmt.Fprint(w, date.MonthName())
			case 'm':
				fmt.Fprint(w, xstrings.FirstN(date.MonthName(), 3))
			case 'N':
				fmt.Fprint(w, date.Month())
			case 'n':
				fmt.Fprintf(w, "%0[1]*[2]d", widthNeeded(len(cal.Months)), date.Month())
			case 'D':
				fmt.Fprint(w, date.DayInMonth())
			case 'd':
				fmt.Fprintf(w, "%0[1]*[2]d", widthNeeded(cal.mostDaysInMonth()), date.DayInMonth())
			case 'Y':
				year := date.Year()
				if cal.PreviousEra != "" {
					switch {
					case cal.Era == cal.PreviousEra:
						fmt.Fprintf(w, "%d %s", year, cal.PreviousEra)
					case year < 0:
						fmt.Fprintf(w, "%d %s", -year, cal.PreviousEra)
					default:
						fmt.Fprint(w, year)
					}
				} else {
					fmt.Fprint(w, year)
				}
			case 'y':
				era := date.Era()
				year := date.Year()
				if year < 0 && era != "" && cal.Era != cal.PreviousEra {
					year = -year
				}
				if era != "" {
					fmt.Fprintf(w, "%d %s", year, era)
				} else {
					fmt.Fprint(w, year)
				}
			case 'z':
				fmt.Fprint(w, date.Year())
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

func widthNeeded(count int) int {
	needed := 1
	for count > 9 {
		count /= 10
		needed++
	}
	return needed
}

// TextCalendarMonth writes a text representation of the month.
func (date Date) TextCalendarMonth(w io.Writer) {
	cal := date.calendar()
	year, month, _, maximum := date.resolve()
	fmt.Fprintf(w, "%d: %s", month, cal.Months[month-1].Name)
	lastDayOfWeek := len(cal.WeekDays) - 1
	width := widthNeeded(cal.mostDaysInMonth())
	for i, weekday := range cal.WeekDays {
		if i == 0 {
			fmt.Fprint(w, "\n")
		} else {
			fmt.Fprint(w, " ")
		}
		for j := 0; j < width-1; j++ {
			fmt.Fprint(w, " ")
		}
		fmt.Fprint(w, xstrings.FirstN(weekday, 1))
	}
	// Consecutive days differ only by one in Days, so derive the first day once and increment rather than rebuilding a
	// fresh Date (with its year convergence and month summation) for every day of the month.
	firstDay := cal.MustNewDate(month, 1, year).Days
	numFmt := fmt.Sprintf("%%%dd", width)
	for i := 1; i <= maximum; i++ {
		weekDay := Date{Days: firstDay + i - 1, cal: cal}.WeekDay()
		if i == 1 || weekDay == 0 {
			fmt.Fprint(w, "\n")
		}
		if i == 1 && weekDay != 0 {
			for j := 0; j < weekDay*(width+1); j++ {
				fmt.Fprint(w, " ")
			}
		}
		fmt.Fprintf(w, numFmt, i)
		if weekDay != lastDayOfWeek {
			fmt.Fprint(w, " ")
		}
	}
	fmt.Fprintln(w)
}
