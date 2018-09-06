// Package calendar provides a customizable calendar for roleplaying games.
package calendar

import "github.com/richardwilkes/toolbox/errs"

// Calendar holds the data for the calendar.
type Calendar struct {
	FirstWeekDayOfFirstYear int       `json:"first_weekday_of_first_year" yaml:"first_weekday_of_first_year"`
	WeekDays                []string  `json:"weekdays"`
	Months                  []Month   `json:"months"`
	Seasons                 []Season  `json:"seasons"`
	YearSuffix              string    `json:"year_suffix,omitempty" yaml:",omitempty"`
	YearBeforeSuffix        string    `json:"year_before_suffix,omitempty" yaml:",omitempty"`
	LeapYear                *LeapYear `json:"leap_year,omitempty" yaml:",omitempty"`
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
	if cal.FirstWeekDayOfFirstYear < 0 || cal.FirstWeekDayOfFirstYear >= len(cal.WeekDays) {
		return errs.New("Calendar's first week day of the zero year must be a valid week day")
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
