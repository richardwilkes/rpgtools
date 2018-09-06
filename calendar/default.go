package calendar

// Gregorian returns a new calendar which mimics the Gregorian calendar,
// although not precisely, as the real-world calendar has a lot of
// irregularities to it prior to the 1600's. If you want a more precise
// real-world calendar, use Go's time.Time instead.
func Gregorian() *Calendar {
	return &Calendar{
		DayZeroWeekDay: 1,
		WeekDays:       []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"},
		Months: []Month{
			{
				Name: "January",
				Days: 31,
			},
			{
				Name: "February",
				Days: 28,
			},
			{
				Name: "March",
				Days: 31,
			},
			{
				Name: "April",
				Days: 30,
			},
			{
				Name: "May",
				Days: 31,
			},
			{
				Name: "June",
				Days: 30,
			},
			{
				Name: "July",
				Days: 31,
			},
			{
				Name: "August",
				Days: 31,
			},
			{
				Name: "September",
				Days: 30,
			},
			{
				Name: "October",
				Days: 31,
			},
			{
				Name: "November",
				Days: 30,
			},
			{
				Name: "December",
				Days: 31,
			},
		},
		Seasons: []Season{
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
		},
		Era:         "AD",
		PreviousEra: "BC",
		LeapYear: &LeapYear{
			Month:  2,
			Every:  4,
			Except: 100,
			Unless: 400,
		},
	}
}
