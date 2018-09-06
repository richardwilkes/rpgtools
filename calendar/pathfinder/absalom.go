package pathfinder

import "github.com/richardwilkes/rpgtools/calendar"

// Absalom returns a new Pathfinder Absalom calendar.
func Absalom() *calendar.Calendar {
	return &calendar.Calendar{
		FirstWeekDayOfFirstYear: 0,
		WeekDays:                []string{"Moonday", "Toilday", "Wealday", "Oathday", "Fireday", "Starday", "Sunday"},
		Months: []calendar.Month{
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
		},
		Seasons: []calendar.Season{
			{
				Name:       "Winter",
				StartMonth: 12,
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
		},
		YearSuffix:       "AR",
		YearBeforeSuffix: "AR",
		LeapYear: &calendar.LeapYear{
			Month: 2,
			Every: 8,
		},
	}
}
