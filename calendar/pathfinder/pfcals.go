// Copyright ©2017-2022 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package pathfinder

import "github.com/richardwilkes/rpgtools/calendar"

// Era names
const (
	AbsalomReckoningEra = "AR"
	ImperialCalendarEra = "IC"
)

// AbsalomReckoning returns a new Pathfinder RPG Absalom Reckoning calendar.
func AbsalomReckoning() *calendar.Calendar {
	return &calendar.Calendar{
		WeekDays:    newWeekdays(),
		Months:      newMonths(),
		Seasons:     newSeasons(),
		Era:         AbsalomReckoningEra,
		PreviousEra: AbsalomReckoningEra,
		LeapYear:    newLeapYear(),
	}
}

// ImperialCalendar returns a new Pathfinder RPG Imperial Calendar.
func ImperialCalendar() *calendar.Calendar {
	return &calendar.Calendar{
		WeekDays:    newWeekdays(),
		Months:      newMonths(),
		Seasons:     newSeasons(),
		Era:         ImperialCalendarEra,
		PreviousEra: ImperialCalendarEra,
		LeapYear:    newLeapYear(),
	}
}

func newWeekdays() []string {
	return []string{
		"Moonday",
		"Toilday",
		"Wealday",
		"Oathday",
		"Fireday",
		"Starday",
		"Sunday",
	}
}

func newMonths() []calendar.Month {
	return []calendar.Month{
		{
			Name: "Abadius",
			Days: 31,
		},
		{
			Name: "Calistril",
			Days: 28,
		},
		{
			Name: "Pharast",
			Days: 31,
		},
		{
			Name: "Gozran",
			Days: 30,
		},
		{
			Name: "Desnus",
			Days: 31,
		},
		{
			Name: "Sarenith",
			Days: 30,
		},
		{
			Name: "Erastus",
			Days: 31,
		},
		{
			Name: "Arodus",
			Days: 31,
		},
		{
			Name: "Rova",
			Days: 30,
		},
		{
			Name: "Lamashan",
			Days: 31,
		},
		{
			Name: "Neth",
			Days: 30,
		},
		{
			Name: "Kuthona",
			Days: 31,
		},
	}
}

func newSeasons() []calendar.Season {
	return []calendar.Season{
		{
			Name:       "Winter",
			StartMonth: 11,
			StartDay:   1,
			EndMonth:   2,
			EndDay:     28,
		},
		{
			Name:       "Spring",
			StartMonth: 3,
			StartDay:   1,
			EndMonth:   5,
			EndDay:     31,
		},
		{
			Name:       "Summer",
			StartMonth: 6,
			StartDay:   1,
			EndMonth:   8,
			EndDay:     31,
		},
		{
			Name:       "Fall",
			StartMonth: 9,
			StartDay:   1,
			EndMonth:   10,
			EndDay:     31,
		},
	}
}

func newLeapYear() *calendar.LeapYear {
	return &calendar.LeapYear{
		Month: 2,
		Every: 8,
	}
}
