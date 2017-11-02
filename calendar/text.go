package calendar

import (
	"fmt"
	"io"
)

// WriteYearBlock writes a text representation of the year.
func WriteYearBlock(year int, w io.Writer) {
	fmt.Fprintln(w, "Year", year)
	max := len(Current.Months)
	for i := 1; i <= max; i++ {
		fmt.Fprintln(w)
		WriteMonthBlock(MustNewDate(i, 1, year), w)
	}
	fmt.Fprintln(w)
	WriteSeasonsBlock(w)
	fmt.Println()
	WriteWeekDaysBlock(w)
}

// WriteMonthBlock writes a text representation of the month for the given
// date.
func WriteMonthBlock(date Date, w io.Writer) {
	mostDays := 0
	for _, m := range Current.Months {
		if mostDays < m.Days {
			mostDays = m.Days
		}
	}
	fmt.Fprintf(w, "%d: %s", date.Month(), date.MonthName())
	lastDayOfWeek := len(Current.WeekDays) - 1
	width := len(fmt.Sprintf("%d", mostDays))
	for i, weekday := range Current.WeekDays {
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
		d := MustNewDate(month, i, year)
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
func WriteSeasonsBlock(w io.Writer) {
	fmt.Fprintln(w, "Seasons:")
	for _, season := range Current.Seasons {
		fmt.Fprintf(w, "  %v\n", &season)
	}
}

// WriteWeekDaysBlock writes a text representation of the week days.
func WriteWeekDaysBlock(w io.Writer) {
	fmt.Fprintln(w, "Week Days:")
	for i, weekday := range Current.WeekDays {
		fmt.Fprintf(w, "  %d: (%s) %s\n", i+1, weekday[:1], weekday)
	}
}
