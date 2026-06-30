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
	"slices"
	"strings"

	"github.com/richardwilkes/toolbox/v2/errs"
)

var (
	absalom   = newPathfinderCalendar("AR")
	imperial  = newPathfinderCalendar("IC")
	gregorian = &Calendar{
		cfg: &Config{
			DayZeroWeekDay: 1,
			WeekDays:       []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"},
			Months: []Month{
				{Name: "January", Days: 31},
				{Name: "February", Days: 28},
				{Name: "March", Days: 31},
				{Name: "April", Days: 30},
				{Name: "May", Days: 31},
				{Name: "June", Days: 30},
				{Name: "July", Days: 31},
				{Name: "August", Days: 31},
				{Name: "September", Days: 30},
				{Name: "October", Days: 31},
				{Name: "November", Days: 30},
				{Name: "December", Days: 31},
			},
			Seasons: []Season{
				{Name: "Winter", StartMonth: 11, StartDay: 1, EndMonth: 2, EndDay: 28},
				{Name: "Spring", StartMonth: 3, StartDay: 1, EndMonth: 5, EndDay: 31},
				{Name: "Summer", StartMonth: 6, StartDay: 1, EndMonth: 8, EndDay: 31},
				{Name: "Fall", StartMonth: 9, StartDay: 1, EndMonth: 10, EndDay: 31},
			},
			Era:         "AD",
			PreviousEra: "BC",
			LeapYear:    &LeapYear{Month: 2, Every: 4, Except: 100, Unless: 400},
		},
	}
)

// Season defines a seasonal period in the calendar. A season is positional within a year and ignores the year itself. A
// season whose start falls on or before its end covers the single contiguous span between them; a season whose start
// falls after its end (such as a winter running 11/1 through 2/28) wraps the year boundary and covers everything from
// the start through the end of the year plus the beginning of the year through the end. A season whose EndDay is the
// last day of a non-leap month is treated as running through the end of that month, so e.g. a winter ending 2/28 still
// contains 2/29 in a leap year (in a non-leap year that day simply does not occur, so the extension is harmless).
type Season struct {
	Name       string `json:"name"`
	StartMonth int    `json:"start_month" yaml:"start_month"`
	StartDay   int    `json:"start_day" yaml:"start_day"`
	EndMonth   int    `json:"end_month" yaml:"end_month"`
	EndDay     int    `json:"end_day" yaml:"end_day"`
}

// DateRange returns the range of dates a season spans as text.
func (s *Season) DateRange() string {
	if s.StartMonth == s.EndMonth && s.StartDay == s.EndDay {
		return fmt.Sprintf("%d/%d", s.StartMonth, s.StartDay)
	}
	return fmt.Sprintf("%d/%d-%d/%d", s.StartMonth, s.StartDay, s.EndMonth, s.EndDay)
}

// Month holds information about a month within a calendar.
type Month struct {
	Name string `json:"name"`
	Days int    `json:"days"`
}

// LeapYear holds parameters for determining leap years.
type LeapYear struct {
	Month  int `json:"month"`
	Every  int `json:"every"`
	Except int `json:"except,omitempty" yaml:",omitempty"`
	Unless int `json:"unless,omitempty" yaml:",omitempty"`
}

// Config holds the configuration data for a Calendar. Seasons may be empty. A season whose start falls after its end is
// permitted: it is interpreted as wrapping the year boundary (see Date.Season). Seasons are likewise permitted to
// overlap one another or to leave gaps in the year; neither is treated as an error.
type Config struct {
	LeapYear       *LeapYear `json:"leapyear,omitempty" yaml:",omitempty"`
	Era            string    `json:"era,omitempty" yaml:",omitempty"`
	PreviousEra    string    `json:"previous_era,omitempty" yaml:"previous_era,omitempty"`
	WeekDays       []string  `json:"weekdays"`
	Months         []Month   `json:"months"`
	Seasons        []Season  `json:"seasons,omitempty"`
	DayZeroWeekDay int       `json:"day_zero_weekday" yaml:"day_zero_weekday"`
}

// Clone this configuration.
func (c *Config) Clone() *Config {
	other := *c
	other.WeekDays = slices.Clone(c.WeekDays)
	other.Months = slices.Clone(c.Months)
	other.Seasons = slices.Clone(c.Seasons)
	if c.LeapYear != nil {
		leapYear := *c.LeapYear
		other.LeapYear = &leapYear
	}
	return &other
}

// Valid returns nil if the data is usable.
func (c *Config) Valid() error {
	if c == nil {
		return errs.New("configuration may not be nil")
	}
	if len(c.WeekDays) == 0 {
		return errs.New("must have at least one week day")
	}
	for _, weekDay := range c.WeekDays {
		if weekDay == "" {
			return errs.New("week day names must not be empty")
		}
		if weekDay != strings.TrimSpace(weekDay) {
			return errs.New("week day names may not begin or end with whitespace")
		}
	}
	if c.DayZeroWeekDay < 0 || c.DayZeroWeekDay >= len(c.WeekDays) {
		return errs.New("DayZeroWeekDay must specify a valid week day")
	}
	if len(c.Months) == 0 {
		return errs.New("must have at least one month")
	}
	for i := range c.Months {
		if c.Months[i].Name == "" {
			return errs.New("month names must not be empty")
		}
		if c.Months[i].Name != strings.TrimSpace(c.Months[i].Name) {
			return errs.New("month names may not begin or end with whitespace")
		}
		if c.Months[i].Days < 1 {
			return errs.New("months must contain at least 1 day")
		}
	}
	if c.LeapYear != nil {
		if c.LeapYear.Month < 1 || c.LeapYear.Month > len(c.Months) {
			return errs.New("LeapYear.Month must specify a valid month")
		}
		if c.LeapYear.Every < 2 {
			return errs.New("LeapYear.Every may not be less than 2")
		}
		if c.LeapYear.Except != 0 {
			if c.LeapYear.Except <= c.LeapYear.Every {
				return errs.New("LeapYear.Except must be greater than LeapYear.Every if not 0")
			}
			if c.LeapYear.Except%c.LeapYear.Every != 0 {
				return errs.New("LeapYear.Except must be a multiple of LeapYear.Every")
			}
		}
		if c.LeapYear.Unless != 0 {
			if c.LeapYear.Except == 0 {
				return errs.New("LeapYear.Unless may not be set if LeapYear.Except is 0")
			}
			if c.LeapYear.Unless <= c.LeapYear.Except {
				return errs.New("LeapYear.Unless must be greater than LeapYear.Except if not 0")
			}
			if c.LeapYear.Unless%c.LeapYear.Except != 0 {
				return errs.New("LeapYear.Unless must be a multiple of LeapYear.Except")
			}
		}
	}
	for i := range c.Seasons {
		if c.Seasons[i].Name == "" {
			return errs.New("season names must not be empty")
		}
		if c.Seasons[i].Name != strings.TrimSpace(c.Seasons[i].Name) {
			return errs.New("season names may not begin or end with whitespace")
		}
		if c.Seasons[i].StartMonth < 1 || c.Seasons[i].StartMonth > len(c.Months) {
			return errs.New("seasons must start in a valid month")
		}
		if c.Seasons[i].StartDay < 1 || c.Seasons[i].StartDay > c.maxDaysInMonth(c.Seasons[i].StartMonth) {
			return errs.New("seasons must start in a valid day within the month")
		}
		if c.Seasons[i].EndMonth < 1 || c.Seasons[i].EndMonth > len(c.Months) {
			return errs.New("seasons must end in a valid month")
		}
		if c.Seasons[i].EndDay < 1 || c.Seasons[i].EndDay > c.maxDaysInMonth(c.Seasons[i].EndMonth) {
			return errs.New("seasons must end in a valid day within the month")
		}
	}
	if c.Era != strings.TrimSpace(c.Era) {
		return errs.New("era may not begin or end with whitespace")
	}
	if c.PreviousEra != strings.TrimSpace(c.PreviousEra) {
		return errs.New("previous era may not begin or end with whitespace")
	}
	if (c.PreviousEra == "") != (c.Era == "") {
		return errs.New("era and previous era must either both be set or neither set")
	}
	return nil
}

func (c *Config) maxDaysInMonth(month int) int {
	if month < 1 || month > len(c.Months) {
		return 0
	}
	days := c.Months[month-1].Days
	if c.isLeapMonth(month) {
		days++
	}
	return days
}

func (c *Config) isLeapMonth(month int) bool {
	return c.LeapYear != nil && c.LeapYear.Month == month
}

// Gregorian returns the Gregorian calendar, although not precisely, as the real-world calendar has a lot of
// irregularities to it prior to the 1600's. If you want a more precise real-world calendar, use Go's time.Time instead.
func Gregorian() *Calendar {
	return gregorian
}

// PathfinderAbsalomReckoning returns the Pathfinder RPG Absalom Reckoning calendar.
func PathfinderAbsalomReckoning() *Calendar {
	return absalom
}

// PathfinderImperialCalendar returns the Pathfinder RPG Imperial Calendar.
func PathfinderImperialCalendar() *Calendar {
	return imperial
}

// newPathfinderCalendar builds the shared Pathfinder calendar structure, which differs between the variants only in the
// era name (the same name serves as both the current and previous era). Each call returns an independent Calendar built
// from fresh component slices.
func newPathfinderCalendar(era string) *Calendar {
	return &Calendar{
		cfg: &Config{
			WeekDays: []string{
				"Moonday",
				"Toilday",
				"Wealday",
				"Oathday",
				"Fireday",
				"Starday",
				"Sunday",
			},
			Months: []Month{
				{Name: "Abadius", Days: 31},
				{Name: "Calistril", Days: 28},
				{Name: "Pharast", Days: 31},
				{Name: "Gozran", Days: 30},
				{Name: "Desnus", Days: 31},
				{Name: "Sarenith", Days: 30},
				{Name: "Erastus", Days: 31},
				{Name: "Arodus", Days: 31},
				{Name: "Rova", Days: 30},
				{Name: "Lamashan", Days: 31},
				{Name: "Neth", Days: 30},
				{Name: "Kuthona", Days: 31},
			},
			Seasons: []Season{
				{Name: "Winter", StartMonth: 11, StartDay: 1, EndMonth: 2, EndDay: 28},
				{Name: "Spring", StartMonth: 3, StartDay: 1, EndMonth: 5, EndDay: 31},
				{Name: "Summer", StartMonth: 6, StartDay: 1, EndMonth: 8, EndDay: 31},
				{Name: "Fall", StartMonth: 9, StartDay: 1, EndMonth: 10, EndDay: 31},
			},
			Era:         era,
			PreviousEra: era,
			LeapYear:    &LeapYear{Month: 2, Every: 8},
		},
	}
}
