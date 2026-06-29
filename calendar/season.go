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

	"github.com/richardwilkes/toolbox/v2/errs"
)

// Season defines a seasonal period in the calendar.
type Season struct {
	Name       string `json:"name"`
	StartMonth int    `json:"start_month" yaml:"start_month"`
	StartDay   int    `json:"start_day" yaml:"start_day"`
	EndMonth   int    `json:"end_month" yaml:"end_month"`
	EndDay     int    `json:"end_day" yaml:"end_day"`
}

func (season *Season) String() string {
	if season.StartMonth == season.EndMonth && season.StartDay == season.EndDay {
		return fmt.Sprintf("%s (%d/%d)", season.Name, season.StartMonth, season.StartDay)
	}
	return fmt.Sprintf("%s (%d/%d-%d/%d)", season.Name, season.StartMonth, season.StartDay, season.EndMonth, season.EndDay)
}

// containsDate reports whether the day-of-year position (month, day) falls within the season for the given calendar.
// The comparison is positional within a year and ignores the year itself. A season whose start falls on or before its
// end covers the single contiguous span between them; a season whose start falls after its end (such as a winter
// running 11/1 through 2/28) wraps the year boundary and covers everything from the start through the end of the year
// plus the beginning of the year through the end. A season whose EndDay is the last day of a non-leap month is treated
// as running through the end of that month, so e.g. a winter ending 2/28 still contains 2/29 in a leap year (in a
// non-leap year that day simply does not occur, so the extension is harmless).
func (season *Season) containsDate(cal *Calendar, month, day int) bool {
	endDay := season.EndDay
	// Extend an end-of-month boundary through the leap day, guarding the month index so a season with an out-of-range
	// EndMonth (possible on a calendar usable for date math but never run through Valid) degrades to a plain positional
	// comparison rather than panicking.
	if season.EndMonth >= 1 && season.EndMonth <= len(cal.Months) && endDay == cal.Months[season.EndMonth-1].Days {
		endDay = cal.maxDaysInMonth(season.EndMonth)
	}
	afterStart := onOrAfter(month, day, season.StartMonth, season.StartDay)
	beforeEnd := onOrAfter(season.EndMonth, endDay, month, day)
	if onOrAfter(season.EndMonth, endDay, season.StartMonth, season.StartDay) {
		return afterStart && beforeEnd // start on or before end: a single contiguous span
	}
	return afterStart || beforeEnd // start after end: the span wraps the year boundary
}

// onOrAfter reports whether the day-of-year position (m1, d1) falls on or after (m2, d2), comparing month first and the
// day within the month second.
func onOrAfter(m1, d1, m2, d2 int) bool {
	if m1 != m2 {
		return m1 > m2
	}
	return d1 >= d2
}

// seasonFor returns the first season, in declaration order, that contains the day-of-year position (month, day), and
// true; or a zero Season and false when no season contains it. Seasons are allowed to overlap, in which case the first
// matching one wins; they are also allowed to leave gaps, in which case a date in a gap yields false.
func (cal *Calendar) seasonFor(month, day int) (Season, bool) {
	for i := range cal.Seasons {
		if cal.Seasons[i].containsDate(cal, month, day) {
			return cal.Seasons[i], true
		}
	}
	return Season{}, false
}

// Valid returns nil if the season data is usable for the given calendar. A season whose start falls after its end is
// permitted: it is interpreted as wrapping the year boundary (see Date.Season). Seasons are likewise permitted to
// overlap one another or to leave gaps in the year; neither is treated as an error.
func (season *Season) Valid(cal *Calendar) error {
	if season.Name == "" {
		return errs.New("Calendar season names must not be empty")
	}
	if season.StartMonth < 1 || season.StartMonth > len(cal.Months) {
		return errs.New("Calendar seasons must start in a valid month")
	}
	if season.StartDay < 1 || season.StartDay > cal.maxDaysInMonth(season.StartMonth) {
		return errs.New("Calendar seasons must start in a valid day within the month")
	}
	if season.EndMonth < 1 || season.EndMonth > len(cal.Months) {
		return errs.New("Calendar seasons must end in a valid month")
	}
	if season.EndDay < 1 || season.EndDay > cal.maxDaysInMonth(season.EndMonth) {
		return errs.New("Calendar seasons must end in a valid day within the month")
	}
	return nil
}
