// Copyright (c) 2017-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calendar_test

import (
	"testing"

	"github.com/richardwilkes/rpgtools/calendar"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestSeasonLeapDayBoundary(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian() // February (month 2) is the leap month; its base length is 28.

	// A season may legitimately start or end on the 29th of the leap month, which is a real calendar day in leap years
	// even though February's base length is 28.
	c.NoError((&calendar.Season{Name: "Winter", StartMonth: 1, StartDay: 1, EndMonth: 2, EndDay: 29}).Valid(cal))
	c.NoError((&calendar.Season{Name: "Thaw", StartMonth: 2, StartDay: 29, EndMonth: 3, EndDay: 1}).Valid(cal))

	// One day past the leap day is still rejected...
	c.HasError((&calendar.Season{Name: "Winter", StartMonth: 1, StartDay: 1, EndMonth: 2, EndDay: 30}).Valid(cal))
	c.HasError((&calendar.Season{Name: "Thaw", StartMonth: 2, StartDay: 30, EndMonth: 3, EndDay: 1}).Valid(cal))
	// ...and a non-leap month gains no extra day, so March (31 days) still rejects the 32nd.
	c.HasError((&calendar.Season{Name: "Spring", StartMonth: 3, StartDay: 1, EndMonth: 3, EndDay: 32}).Valid(cal))
}
