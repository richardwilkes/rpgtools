package calendar

import (
	"fmt"
	"io"
	"strings"

	"github.com/richardwilkes/toolbox/txt"
)

// Predefined formats.
const (
	FullFormat   = "%W, %M %D, %Y"
	LongFormat   = "%M %D, %Y"
	MediumFormat = "%m %D, %Y"
	ShortFormat  = "%N/%D/%Y"
)

// Format returns a formatted version of the date. The layout is parsed as in
// WriteFormat().
func (date Date) Format(layout string) string {
	var buffer strings.Builder
	date.WriteFormat(&buffer, layout)
	return buffer.String()
}

// WriteFormat writes a formatted version of the date to the writer. The
// layout is parsed for directives and anything that is not a directive is
// passed through unchanged. Valid directives:
//
//   %W  Full weekday, e.g. 'Friday'
//   %w  Short weekday, e.g. 'Fri'
//   %M  Full month name, e.g. 'September'
//   %m  Short month name, e.g. 'Sep'
//   %N  Month, e.g. '9'
//   %n  Month padded with zeroes, e.g. '09'
//   %D  Day, e.g. '2'
//   %d  Day padded with zeroes, e.g. '02'
//   %Y  Year, e.g. '2017' if positive, '2017 BC' if negative
//   %y  Year with dating suffix, e.g. '2017 AD'
//   %%  %
func (date Date) WriteFormat(w io.Writer, layout string) {
	cmd := false
	for _, r := range layout {
		switch {
		case cmd:
			cmd = false
			switch r {
			case 'W':
				fmt.Fprint(w, date.WeekDayName())
			case 'w':
				fmt.Fprint(w, txt.FirstN(date.WeekDayName(), 3))
			case 'M':
				fmt.Fprint(w, date.MonthName())
			case 'm':
				fmt.Fprint(w, txt.FirstN(date.MonthName(), 3))
			case 'N':
				fmt.Fprint(w, date.Month())
			case 'n':
				fmt.Fprintf(w, "%0[1]*[2]d", widthNeeded(len(Current.Months)), date.Month())
			case 'D':
				fmt.Fprint(w, date.DayInMonth())
			case 'd':
				fmt.Fprintf(w, "%0[1]*[2]d", widthNeeded(Current.Months[date.Month()].Days), date.DayInMonth())
			case 'Y':
				year := date.Year()
				if year < 0 && Current.YearBeforeSuffix != "" {
					fmt.Fprintf(w, "%d %s", -year, Current.YearBeforeSuffix)
				} else {
					fmt.Fprint(w, year)
				}
			case 'y':
				suffix := date.Suffix()
				year := date.Year()
				if year < 0 && suffix != "" {
					year = -year
				}
				if suffix != "" {
					fmt.Fprintf(w, "%d %s", year, suffix)
				} else {
					fmt.Fprint(w, year)
				}
			case '%':
				fmt.Fprint(w, "%")
			}
		case r == '%':
			cmd = true
		default:
			fmt.Fprintf(w, "%c", r)
		}
	}
}

func widthNeeded(count int) int {
	needed := 1
	for count > 9 {
		count /= 10
		needed++
	}
	return needed
}
