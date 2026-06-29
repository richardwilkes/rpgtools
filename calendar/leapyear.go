// Copyright (c) 2017-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calendar

import "github.com/richardwilkes/toolbox/v2/errs"

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
		if leapYear.Except%leapYear.Every != 0 {
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
		if leapYear.Unless%leapYear.Except != 0 {
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

// Since returns the number of leap years that have occurred between year 1 and the specified year, exclusive.
func (leapYear *LeapYear) Since(year int) int {
	if year >= 1 {
		return leapYear.countLeaps(year - 1)
	}
	// There is no year 0, so the years strictly between year and 1 run year+1..-1. Is() derives a negative year's leap
	// status from the magnitude |year+1|, so those years map to magnitudes 0..(-year-2). countLeaps covers magnitudes 1
	// and up; magnitude 0 (year -1) is added on separately because whether it is a leap year depends on the
	// Except/Unless rule, which is exactly what Is(-1) reports.
	upper := -year - 2
	if upper < 0 {
		return 0 // year == -1: nothing lies strictly between it and year 1
	}
	count := leapYear.countLeaps(upper)
	if leapYear.Is(-1) {
		count++
	}
	return count
}

// countLeaps returns the number of leap years whose magnitude (distance from the leap pattern's origin) is 1 through n
// inclusive. The leap pattern is symmetric about the origin, so the same closed form serves positive years directly and
// negative years via their shifted magnitude. n must not be negative. Valid guarantees every multiple of Except is a
// multiple of Every and every multiple of Unless is a multiple of Except, so dividing counts each tier independently.
func (leapYear *LeapYear) countLeaps(n int) int {
	count := n / leapYear.Every
	if leapYear.Except != 0 {
		count -= n / leapYear.Except
		if leapYear.Unless != 0 {
			count += n / leapYear.Unless
		}
	}
	return count
}
