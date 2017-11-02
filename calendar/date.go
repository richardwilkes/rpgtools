package calendar

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/richardwilkes/toolbox/errs"
)

var (
	// 9/22/2017
	regexMMDDYYY = regexp.MustCompile("([[:digit:]]+)/([[:digit:]]+)/(-?[[:digit:]]+)")
	// September 22, 2017
	regexMonthDDYYYY = regexp.MustCompile("([[:alpha:]]+) *([[:digit:]]+), *(-?[[:digit:]]+)")
)

// Date holds a calendar date.
type Date int

// MustNewDate creates a new date from the specified month, day and year.
// Panics if the values are invalid.
func MustNewDate(month, day, year int) Date {
	date, err := NewDate(month, day, year)
	if err != nil {
		panic(err)
	}
	return date
}

// NewDate creates a new date from the specified month, day and year.
func NewDate(month, day, year int) (Date, error) {
	if month < 1 || month > len(Current.Months) {
		return 0, errs.New("month out of range")
	}
	if day < 1 || day > Current.Months[month-1].Days {
		return 0, errs.New("day out of range")
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
			return 0, errs.NewWithCause("invalid month text", err)
		}
		return parseDate(month, parts[2], parts[3])
	}
	if parts := regexMonthDDYYYY.FindStringSubmatch(in); parts != nil {
		for i, month := range Current.Months {
			if strings.EqualFold(parts[1], month.Name) {
				return parseDate(i+1, parts[2], parts[3])
			}
		}
		return 0, errs.New("invalid month text")
	}
	return 0, errs.New("invalid date text")
}

func parseDate(month int, dayText, yearText string) (Date, error) {
	year, err := strconv.Atoi(yearText)
	if err != nil {
		return 0, errs.NewWithCause("invalid year text", err)
	}
	day, err := strconv.Atoi(dayText)
	if err != nil {
		return 0, errs.NewWithCause("invalid day text", err)
	}
	return NewDate(month, day, year)
}

// Year returns the year of the date.
func (date Date) Year() int {
	daysPerYear := Current.DaysPerYear()
	days := int(date)
	if date < 0 {
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
	panic("Unable to determine month")
}

// MonthName returns the name of the month of the date.
func (date Date) MonthName() string {
	return Current.Months[date.Month()-1].Name
}

// DayInYear returns the day within the year of the date. Note that the first
// day is represented by a 1, not 0.
func (date Date) DayInYear() int {
	return 1 + int(date) - (date.Year() * Current.DaysPerYear())
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
	panic("Unable to determine day in month")
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
	return (weekday + Current.FirstWeekDayOfZeroYear) % len(Current.WeekDays)
}

// WeekDayName returns the name of the weekday of the date.
func (date Date) WeekDayName() string {
	return Current.WeekDays[date.WeekDay()]
}

// String returns a date in the format "9/22/2017".
func (date Date) String() string {
	return fmt.Sprintf("%d/%d/%d", date.Month(), date.DayInMonth(), date.Year())
}

// MediumString returns a date in the format "September 22, 2017".
func (date Date) MediumString() string {
	return fmt.Sprintf("%s %d, %d", Current.Months[date.Month()-1].Name, date.DayInMonth(), date.Year())
}

// LongString returns a date in the format "Friday, September 22, 2017".
func (date Date) LongString() string {
	return fmt.Sprintf("%s, %s %d, %d", Current.WeekDays[date.WeekDay()], Current.Months[date.Month()-1].Name, date.DayInMonth(), date.Year())
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
