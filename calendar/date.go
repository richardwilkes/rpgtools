package calendar

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/txt"
)

var (
	// "9/22/2017" or "9/22/2017 AD"
	regexMMDDYYY = regexp.MustCompile("([[:digit:]]+)/([[:digit:]]+)/(-?[[:digit:]]+) *([[:alpha:]]+)?")
	// "September 22, 2017 AD", "September 22, 2017", "Sep 22, 2017 AD", or "Sep 22, 2017"
	regexMonthDDYYYY = regexp.MustCompile("([[:alpha:]]+) *([[:digit:]]+), *(-?[[:digit:]]+) *([[:alpha:]]+)?")
)

// Date holds a calendar date.
type Date int64

// MustNewDate creates a new date from the specified month, day and year.
// Panics if the values are invalid.
func MustNewDate(month, day, year int) Date {
	date, err := NewDate(month, day, year)
	if err != nil {
		panic(err) // @allow
	}
	return date
}

// NewDate creates a new date from the specified month, day and year.
func NewDate(month, day, year int) (Date, error) {
	if month < 1 || month > len(Current.Months) {
		return 0, errs.Newf("month %d is invalid", month)
	}
	if day < 1 || day > Current.Months[month-1].Days {
		return 0, errs.Newf("day %d is invalid", day)
	}
	if year == 0 {
		return 0, errs.New("year 0 is invalid")
	}
	if year < 0 {
		year++
	}
	date := year*Current.DaysPerYear() + day - 1
	for i := 1; i < month; i++ {
		date += Current.Months[i-1].Days
	}
	return Date(date), nil
}

// ParseDate creates a new date from the specified text.
func ParseDate(in string) (Date, error) {
	if parts := regexMMDDYYY.FindStringSubmatch(in); parts != nil {
		month, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, errs.NewfWithCause(err, "invalid month text '%s'", parts[1])
		}
		return parseDate(month, parts[2], parts[3], parts[4])
	}
	if parts := regexMonthDDYYYY.FindStringSubmatch(in); parts != nil {
		for i, month := range Current.Months {
			if strings.EqualFold(parts[1], month.Name) || strings.EqualFold(parts[1], txt.FirstN(month.Name, 3)) {
				return parseDate(i+1, parts[2], parts[3], parts[4])
			}
		}
		return 0, errs.Newf("invalid month text '%s'", parts[1])
	}
	return 0, errs.Newf("invalid date text '%s'", in)
}

func parseDate(month int, dayText, yearText, suffixText string) (Date, error) {
	year, err := strconv.Atoi(yearText)
	if err != nil {
		return 0, errs.NewfWithCause(err, "invalid year text '%s'", yearText)
	}
	day, err := strconv.Atoi(dayText)
	if err != nil {
		return 0, errs.NewfWithCause(err, "invalid day text '%s'", dayText)
	}
	if Current.YearBeforeSuffix != "" && strings.EqualFold(Current.YearBeforeSuffix, suffixText) {
		year = -year
	}
	return NewDate(month, day, year)
}

// Year returns the year of the date.
func (date Date) Year() int {
	daysPerYear := Current.DaysPerYear()
	days := int(date)
	if days < daysPerYear {
		days -= daysPerYear
		return -(1 - ((days + 1) / daysPerYear))
	}
	return days / daysPerYear
}

// Month returns the month of the date. Note that the first month is
// represented by 1, not 0.
func (date Date) Month() int {
	days := date.DayInYear()
	for i, month := range Current.Months {
		if days <= month.Days {
			return i + 1
		}
		days -= month.Days
	}
	// If this is reached, the algorithm is wrong.
	panic("Unable to determine month") // @allow
}

// MonthName returns the name of the month of the date.
func (date Date) MonthName() string {
	return Current.Months[date.Month()-1].Name
}

// DayInYear returns the day within the year of the date. Note that the first
// day is represented by a 1, not 0.
func (date Date) DayInYear() int {
	year := date.Year()
	if year < 0 {
		year++
	}
	return 1 + int(date) - (year * Current.DaysPerYear())
}

// DayInMonth returns the day within the month of the date. Note that the
// first day is represented by a 1, not 0.
func (date Date) DayInMonth() int {
	days := date.DayInYear()
	for _, month := range Current.Months {
		if days <= month.Days {
			return days
		}
		days -= month.Days
	}
	// If this is reached, the algorithm is wrong.
	panic("Unable to determine day in month") // @allow
}

// DaysInMonth returns the number of days in the month of the date.
func (date Date) DaysInMonth() int {
	return Current.Months[date.Month()-1].Days
}

// WeekDay returns the weekday of the date.
func (date Date) WeekDay() int {
	weekday := int(date) % len(Current.WeekDays)
	if date < 0 {
		weekday += len(Current.WeekDays)
	}
	return (weekday + Current.FirstWeekDayOfFirstYear) % len(Current.WeekDays)
}

// WeekDayName returns the name of the weekday of the date.
func (date Date) WeekDayName() string {
	return Current.WeekDays[date.WeekDay()]
}

// Suffix returns the suffix for the year.
func (date Date) Suffix() string {
	if date.Year() < 0 {
		return Current.YearBeforeSuffix
	}
	return Current.YearSuffix
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
	d, err := ParseDate(string(text))
	if err != nil {
		return err
	}
	*date = d
	return nil
}
