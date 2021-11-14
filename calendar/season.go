// Copyright ©2017-2020 by Richard A. Wilkes. All rights reserved.
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

	"github.com/richardwilkes/toolbox/errs"
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

// Valid returns nil if the season data is usable for the given calendar.
func (season *Season) Valid(cal *Calendar) error {
	if season.Name == "" {
		return errs.New("Calendar season names must not be empty")
	}
	if season.StartMonth < 1 || season.StartMonth > len(cal.Months) {
		return errs.New("Calendar seasons must start in a valid month")
	}
	if season.StartDay < 1 || season.StartDay > cal.Months[season.StartMonth-1].Days {
		return errs.New("Calendar seasons must start in a valid day within the month")
	}
	if season.EndMonth < 1 || season.EndMonth > len(cal.Months) {
		return errs.New("Calendar seasons must end in a valid month")
	}
	if season.EndDay < 1 || season.EndDay > cal.Months[season.EndMonth-1].Days {
		return errs.New("Calendar seasons must end in a valid day within the month")
	}
	return nil
}
