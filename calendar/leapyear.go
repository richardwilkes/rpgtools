// Copyright ©2017-2020 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calendar

import "github.com/richardwilkes/toolbox/errs"

// LeapYear holds parameters for determining leap years.
type LeapYear struct {
	Month  int `json:"month"`
	Every  int `json:"every"`
	Except int `json:"except,omitempty" yaml:",omitempty"`
	Unless int `json:"unless,omitempty" yaml:",omitempty"`
}

// Valid returns nil if the leap year data is usable.
func (leapYear *LeapYear) Valid(cal *Calendar) error {
	if leapYear.Month < 1 || leapYear.Month > len(cal.Months) {
		return errs.New("LeapYear.Month must specify a valid month")
	}
	if leapYear.Every < 2 {
		return errs.New("LeapYear.Every may not be less than 2")
	}
	if leapYear.Except != 0 {
		if leapYear.Except <= leapYear.Every {
			return errs.New("LeapYear.Except must be greater than LeapYear.Every if not 0")
		}
		if (leapYear.Except/leapYear.Every)*leapYear.Every != leapYear.Except {
			return errs.New("LeapYear.Except must be a multiple of LeapYear.Every")
		}
	}
	if leapYear.Unless != 0 {
		if leapYear.Except == 0 {
			return errs.New("LeapYear.Unless may not be set if LeapYear.Except is 0")
		}
		if leapYear.Unless <= leapYear.Except {
			return errs.New("LeapYear.Unless must be greater than LeapYear.Except if not 0")
		}
		if (leapYear.Unless/leapYear.Except)*leapYear.Except != leapYear.Unless {
			return errs.New("LeapYear.Unless must be a multiple of LeapYear.Except")
		}
	}
	return nil
}

// Is returns true if the year is a leap year.
func (leapYear *LeapYear) Is(year int) bool {
	if year < 1 {
		year++ // account for gap, since there is no year 0
	}
	if year%leapYear.Every == 0 {
		if leapYear.Except != 0 {
			if year%leapYear.Except == 0 {
				if leapYear.Unless != 0 {
					return year%leapYear.Unless == 0
				}
				return false
			}
		}
		return true
	}
	return false
}

// Since returns the number of leap years that have occurred between year 1
// and the specified year, exclusive.
func (leapYear *LeapYear) Since(year int) int {
	if year == -1 {
		return 0
	}
	delta := year
	if delta < 1 {
		delta = -(delta + 1) // make it positive and account for gap, since there is no year 0
	}
	count := delta / leapYear.Every
	if leapYear.Except != 0 {
		count -= delta / leapYear.Except
		if leapYear.Unless != 0 {
			count += delta / leapYear.Unless
		}
	}
	if leapYear.Is(year) {
		count--
	}
	if year < -1 {
		count++
	}
	return count
}
