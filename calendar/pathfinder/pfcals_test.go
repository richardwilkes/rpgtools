// Copyright (c) 2017-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package pathfinder_test

import (
	"testing"

	"github.com/richardwilkes/rpgtools/calendar/pathfinder"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestPathfinderCalendars(t *testing.T) {
	c := check.New(t)
	ar := pathfinder.AbsalomReckoning()
	ic := pathfinder.ImperialCalendar()

	// Both calendars must be structurally valid.
	c.NoError(ar.Valid())
	c.NoError(ic.Valid())

	// Each names its single era after itself (Era == PreviousEra), and the two calendars use distinct era names.
	c.Equal(pathfinder.AbsalomReckoningEra, ar.Era)
	c.Equal(pathfinder.AbsalomReckoningEra, ar.PreviousEra)
	c.Equal(pathfinder.ImperialCalendarEra, ic.Era)
	c.Equal(pathfinder.ImperialCalendarEra, ic.PreviousEra)
	c.True(ar.Era != ic.Era, "the two Pathfinder calendars must use distinct era names")

	// The era is the only difference: every other component is identical between the two variants.
	c.Equal(ar.WeekDays, ic.WeekDays)
	c.Equal(ar.Months, ic.Months)
	c.Equal(ar.Seasons, ic.Seasons)
	c.Equal(ar.LeapYear, ic.LeapYear)

	// Each call returns an independent calendar built from fresh slices, so mutating one must not affect another.
	ar2 := pathfinder.AbsalomReckoning()
	ar.Months[0].Name = "Changed"
	c.Equal("Abadius", ar2.Months[0].Name, "calendars from separate calls must not share month storage")
}
