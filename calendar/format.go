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
//   %Y  Year, e.g. '2017' if positive, '2017 BC' if negative; however, if the
//       dating suffixes aren't empty and match each other, then this will
//       behave the same as %y
//   %y  Year with dating suffix, e.g. '2017 AD'; however, if the dating
//       suffixes are empty or they match each other, then negative years will
//       result in '-2017 AD'
//   %z  Year without dating suffix, e.g. '2017' or '-2017'
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
				if Current.YearBeforeSuffix != "" {
					if Current.YearSuffix == Current.YearBeforeSuffix {
						fmt.Fprintf(w, "%d %s", year, Current.YearBeforeSuffix)
					} else if year < 0 {
						fmt.Fprintf(w, "%d %s", -year, Current.YearBeforeSuffix)
					} else {
						fmt.Fprint(w, year)
					}
				} else {
					fmt.Fprint(w, year)
				}
			case 'y':
				suffix := date.Suffix()
				year := date.Year()
				if year < 0 && suffix != "" && Current.YearSuffix != Current.YearBeforeSuffix {
					year = -year
				}
				if suffix != "" {
					fmt.Fprintf(w, "%d %s", year, suffix)
				} else {
					fmt.Fprint(w, year)
				}
			case 'z':
				fmt.Fprint(w, date.Year())
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
