package calendar

// Date holds a calendar date. This is the number of days since 1/1/1 in the
// calendar. Note that the value -1 refers to the last day of the year -1, not
// year 0, as there is no year 0.
type Date struct {
	Days int
	cal  *Calendar
}

// Year returns the year of the date.
func (date Date) Year() int {
	estimate := date.Days / date.cal.MinDaysPerYear()
	if date.Days < 0 {
		estimate--
		for date.Days >= date.cal.yearToDays(estimate+1) {
			estimate++
		}
	} else {
		estimate++
		for date.Days < date.cal.yearToDays(estimate) {
			estimate--
		}
	}
	return estimate
}

// Month returns the month of the date. Note that the first month is
// represented by 1, not 0.
func (date Date) Month() int {
	isLeapYear := date.cal.IsLeapYear(date.Year())
	days := date.DayInYear()
	for i, month := range date.cal.Months {
		amt := month.Days
		if isLeapYear && date.cal.IsLeapMonth(i+1) {
			amt++
		}
		if days <= amt {
			return i + 1
		}
		days -= amt
	}
	// If this is reached, the algorithm is wrong.
	panic("Unable to determine month") // @allow
}

// MonthName returns the name of the month of the date.
func (date Date) MonthName() string {
	return date.cal.Months[date.Month()-1].Name
}

// DayInYear returns the day within the year of the date. Note that the first
// day is represented by a 1, not 0.
func (date Date) DayInYear() int {
	return 1 + date.Days - date.cal.yearToDays(date.Year())
}

// DayInMonth returns the day within the month of the date. Note that the
// first day is represented by a 1, not 0.
func (date Date) DayInMonth() int {
	isLeapYear := date.cal.IsLeapYear(date.Year())
	days := date.DayInYear()
	for i, month := range date.cal.Months {
		amt := month.Days
		if isLeapYear && date.cal.IsLeapMonth(i+1) {
			amt++
		}
		if days <= amt {
			return days
		}
		days -= amt
	}
	// If this is reached, the algorithm is wrong.
	panic("Unable to determine day in month") // @allow
}

// DaysInMonth returns the number of days in the month of the date.
func (date Date) DaysInMonth() int {
	return date.cal.Months[date.Month()-1].Days
}

// WeekDay returns the weekday of the date.
func (date Date) WeekDay() int {
	weekday := date.Days % len(date.cal.WeekDays)
	if date.Days < 0 {
		weekday += len(date.cal.WeekDays)
	}
	return (weekday + date.cal.DayZeroWeekDay) % len(date.cal.WeekDays)
}

// WeekDayName returns the name of the weekday of the date.
func (date Date) WeekDayName() string {
	return date.cal.WeekDays[date.WeekDay()]
}

// Era returns the era suffix for the year.
func (date Date) Era() string {
	if date.Year() < 0 {
		return date.cal.PreviousEra
	}
	return date.cal.Era
}

// String returns a date in the ShortFormat.
func (date Date) String() string {
	return date.Format(ShortFormat)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (date *Date) MarshalText() ([]byte, error) {
	return []byte(date.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (date *Date) UnmarshalText(text []byte) error {
	cal := date.cal
	if cal == nil {
		cal = Default
	}
	d, err := cal.ParseDate(string(text))
	if err != nil {
		return err
	}
	*date = d
	return nil
}
