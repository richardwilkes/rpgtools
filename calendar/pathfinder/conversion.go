package pathfinder

import "github.com/richardwilkes/rpgtools/calendar"

// AbsalomToImperial converts an Absalom date to an Imperial one.
func AbsalomToImperial(d calendar.Date) calendar.Date {
	year := d.Year()
	month := d.Month()
	day := d.DayInMonth()
	d2, err := calendar.NewDate(month, day, year+2500)
	if err != nil {
		// Must have hit a leap year variance
		d2 = calendar.MustNewDate(month+1, 1, year+2500)
	}
	return d2
}

// ImperialToAbsalom converts an Imperial date to an Absalom one.
func ImperialToAbsalom(d calendar.Date) calendar.Date {
	year := d.Year()
	month := d.Month()
	day := d.DayInMonth()
	d2, err := calendar.NewDate(month, day, year-2500)
	if err != nil {
		// Must have hit a leap year variance
		d2 = calendar.MustNewDate(month+1, 1, year-2500)
	}
	return d2
}
