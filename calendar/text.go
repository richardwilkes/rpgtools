package calendar

import (
	"fmt"
	"io"
)

// WriteYearBlock writes a text representation of the year.
func (date Date) WriteYearBlock(w io.Writer) {
	date.WriteFormat(w, "Year %Y\n")
	year := date.Year()
	max := len(date.cal.Months)
	for i := 1; i <= max; i++ {
		fmt.Fprintln(w)
		date.cal.MustNewDate(i, 1, year).WriteMonthBlock(w)
	}
	fmt.Fprintln(w)
	date.cal.WriteSeasonsBlock(w)
	fmt.Println()
	date.cal.WriteWeekDaysBlock(w)
}

// WriteMonthBlock writes a text representation of the month for the given
// date.
func (date Date) WriteMonthBlock(w io.Writer) {
	mostDays := 0
	for _, m := range date.cal.Months {
		if mostDays < m.Days {
			mostDays = m.Days
		}
	}
	fmt.Fprintf(w, "%d: %s", date.Month(), date.MonthName())
	lastDayOfWeek := len(date.cal.WeekDays) - 1
	width := len(fmt.Sprintf("%d", mostDays))
	for i, weekday := range date.cal.WeekDays {
		if i == 0 {
			fmt.Fprint(w, "\n")
		} else {
			fmt.Fprint(w, " ")
		}
		for j := 0; j < width-1; j++ {
			fmt.Fprint(w, " ")
		}
		fmt.Fprint(w, weekday[:1])
	}
	max := date.DaysInMonth()
	year := date.Year()
	month := date.Month()
	numFmt := fmt.Sprintf("%%%dd", width)
	for i := 1; i <= max; i++ {
		d := date.cal.MustNewDate(month, i, year)
		weekDay := d.WeekDay()
		if i == 1 || weekDay == 0 {
			fmt.Fprint(w, "\n")
		}
		if i == 1 && weekDay != 0 {
			for j := 0; j < weekDay*(width+1); j++ {
				fmt.Fprint(w, " ")
			}
		}
		fmt.Fprintf(w, numFmt, i)
		if weekDay != lastDayOfWeek {
			fmt.Fprint(w, " ")
		}
	}
	fmt.Fprintln(w)
}

// WriteSeasonsBlock writes a text representation of the seasons.
func (cal *Calendar) WriteSeasonsBlock(w io.Writer) {
	fmt.Fprintln(w, "Seasons:")
	for i := range cal.Seasons {
		fmt.Fprintf(w, "  %v\n", &cal.Seasons[i])
	}
}

// WriteWeekDaysBlock writes a text representation of the week days.
func (cal *Calendar) WriteWeekDaysBlock(w io.Writer) {
	fmt.Fprintln(w, "Week Days:")
	for i, weekday := range cal.WeekDays {
		fmt.Fprintf(w, "  %d: (%s) %s\n", i+1, weekday[:1], weekday)
	}
}
