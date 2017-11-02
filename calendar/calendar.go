// Package calendar provides a customizable calendar for roleplaying games.
package calendar

import "github.com/richardwilkes/toolbox/errs"

// Calendar holds the data for the calendar.
type Calendar struct {
	FirstWeekDayOfZeroYear int      `json:"first_weekday_of_zero_year" yaml:"first_weekday_of_zero_year"`
	WeekDays               []string `json:"weekdays" yaml:"weekdays"`
	Months                 []Month  `json:"months" yaml:"months"`
	Seasons                []Season `json:"seasons" yaml:"seasons"`
}

// DaysPerYear returns the number of days contained in a single year.
func (cal *Calendar) DaysPerYear() int {
	days := 0
	for _, month := range cal.Months {
		days += month.Days
	}
	return days
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
	if cal.FirstWeekDayOfZeroYear < 0 || cal.FirstWeekDayOfZeroYear >= len(cal.WeekDays) {
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
	for _, season := range cal.Seasons {
		if err := season.Valid(cal); err != nil {
			return err
		}
	}
	return nil
}
