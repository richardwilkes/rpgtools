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

// Month holds information about a month within the calendar.
type Month struct {
	Name string `json:"name"`
	Days int    `json:"days"`
}

// Valid returns nil if the month data is usable.
func (month *Month) Valid() error {
	if month.Name == "" {
		return errs.New("Calendar month names must not be empty")
	}
	if month.Days < 1 {
		return errs.New("Calendar months must have at least 1 day")
	}
	return nil
}
