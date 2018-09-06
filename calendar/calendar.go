// Package calendar provides a customizable calendar for roleplaying games.
package calendar

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/txt"
)

var (
	// Default is the default calendar that will be used by Date.UnmarshalText()
	// if the date was not initialized.
	Default = Gregorian()
	// "9/22/2017" or "9/22/2017 AD"
	regexMMDDYYY = regexp.MustCompile("([[:digit:]]+)/([[:digit:]]+)/(-?[[:digit:]]+) *([[:alpha:]]+)?")
	// "September 22, 2017 AD", "September 22, 2017", "Sep 22, 2017 AD", or "Sep 22, 2017"
	regexMonthDDYYYY = regexp.MustCompile("([[:alpha:]]+) *([[:digit:]]+), *(-?[[:digit:]]+) *([[:alpha:]]+)?")
)

// Calendar holds the data for the calendar.
type Calendar struct {
	DayZeroWeekDay int       `json:"day_zero_weekday" yaml:"day_zero_weekday"`
	WeekDays       []string  `json:"weekdays"`
	Months         []Month   `json:"months"`
	Seasons        []Season  `json:"seasons"`
	Era            string    `json:"era,omitempty" yaml:",omitempty"`
	PreviousEra    string    `json:"previous_era,omitempty" yaml:"previous_era,omitempty"`
	LeapYear       *LeapYear `json:"leapyear,omitempty" yaml:",omitempty"`
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
	for _, weekday := range cal.WeekDays {
		if weekday == "" {
			return errs.New("Calendar week day names must not be empty")
		}
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

// MustNewDate creates a new date from the specified month, day and year.
// Panics if the values are invalid.
func (cal *Calendar) MustNewDate(month, day, year int) Date {
	date, err := cal.NewDate(month, day, year)
	if err != nil {
		panic(err) // @allow
	}
	return date
}

// NewDate creates a new date from the specified month, day and year.
func (cal *Calendar) NewDate(month, day, year int) (Date, error) {
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

// NewDateByDays creates a new date from a number of days, with 0 representing
// the date 1/1/1.
func (cal *Calendar) NewDateByDays(days int) Date {
	return Date{Days: days, cal: cal}
}

func (cal *Calendar) yearToDays(year int) int {
	var days int
	if year > 1 {
		days = (year - 1) * cal.MinDaysPerYear()
	} else if year < 0 {
		days = year * cal.MinDaysPerYear()
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
			return Date{cal: cal}, errs.NewfWithCause(err, "invalid month text '%s'", parts[1])
		}
		return cal.parseDate(month, parts[2], parts[3], parts[4])
	}
	if parts := regexMonthDDYYYY.FindStringSubmatch(in); parts != nil {
		for i, month := range cal.Months {
			if strings.EqualFold(parts[1], month.Name) || strings.EqualFold(parts[1], txt.FirstN(month.Name, 3)) {
				return cal.parseDate(i+1, parts[2], parts[3], parts[4])
			}
		}
		return Date{cal: cal}, errs.Newf("invalid month text '%s'", parts[1])
	}
	return Date{cal: cal}, errs.Newf("invalid date text '%s'", in)
}

func (cal *Calendar) parseDate(month int, dayText, yearText, eraText string) (Date, error) {
	year, err := strconv.Atoi(yearText)
	if err != nil {
		return Date{cal: cal}, errs.NewfWithCause(err, "invalid year text '%s'", yearText)
	}
	day, err := strconv.Atoi(dayText)
	if err != nil {
		return Date{cal: cal}, errs.NewfWithCause(err, "invalid day text '%s'", dayText)
	}
	if cal.PreviousEra != "" && strings.EqualFold(cal.PreviousEra, eraText) {
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
	if cal.LeapYear != nil {
		return cal.LeapYear.Is(year)
	}
	return false
}

// IsLeapMonth returns true if the month is the leap month.
func (cal *Calendar) IsLeapMonth(month int) bool {
	if cal.LeapYear != nil {
		return cal.LeapYear.Month == month
	}
	return false
}

// Text writes a text representation of the year.
func (cal *Calendar) Text(year int, w io.Writer) {
	date := cal.MustNewDate(1, 1, year)
	date.WriteFormat(w, "Year %Y\n")
	max := len(cal.Months)
	for i := 1; i <= max; i++ {
		fmt.Fprintln(w)
		cal.MustNewDate(i, 1, year).TextCalendarMonth(w)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Seasons:")
	for i := range cal.Seasons {
		fmt.Fprintf(w, "  %v\n", &cal.Seasons[i])
	}
	fmt.Println()
	fmt.Fprintln(w, "Week Days:")
	for i, weekday := range cal.WeekDays {
		fmt.Fprintf(w, "  %d: (%s) %s\n", i+1, weekday[:1], weekday)
	}
}
