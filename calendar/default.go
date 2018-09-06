package calendar

// Current is the current calendar.
var Current = Default()

// Default returns a new default calendar.
func Default() *Calendar {
	return &Calendar{
		FirstWeekDayOfFirstYear: 1,
		WeekDays:                []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"},
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
		YearSuffix:       "AD",
		YearBeforeSuffix: "BC",
		LeapYear: &LeapYear{
			Month:  2,
			Every:  4,
			Except: 100,
			Unless: 400,
		},
	}
}
