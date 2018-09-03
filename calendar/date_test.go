package calendar_test

import (
	"encoding/json"
	"testing"

	"github.com/richardwilkes/rpgtools/calendar"
	"github.com/stretchr/testify/assert"
)

func TestYear(t *testing.T) {
	daysPerYear := calendar.Current.DaysPerYear()
	assert.Equal(t, -3, calendar.Date(-daysPerYear-1).Year(), "Last day of year -3")
	assert.Equal(t, -2, calendar.Date(-daysPerYear).Year(), "First day of year -2")
	assert.Equal(t, -2, calendar.Date(-1).Year(), "Last day of year -2")
	assert.Equal(t, -1, calendar.Date(0).Year(), "First day of year -1")
	assert.Equal(t, -1, calendar.Date(daysPerYear-1).Year(), "Last day of year -1")
	assert.Equal(t, 1, calendar.Date(daysPerYear).Year(), "First day of year 1")
	assert.Equal(t, 1, calendar.Date(daysPerYear*2-1).Year(), "Last day of year 1")
	assert.Equal(t, 2, calendar.Date(daysPerYear*2).Year(), "First day of year 2")
}

func TestDayInYear(t *testing.T) {
	daysPerYear := calendar.Current.DaysPerYear()
	assert.Equal(t, daysPerYear, calendar.Date(-daysPerYear-1).DayInYear(), "Last day of year -3")
	assert.Equal(t, 1, calendar.Date(-daysPerYear).DayInYear(), "First day of year -2")
	assert.Equal(t, daysPerYear, calendar.Date(-1).DayInYear(), "Last day of year -2")
	assert.Equal(t, 1, calendar.Date(0).DayInYear(), "First day of year -1")
	assert.Equal(t, daysPerYear, calendar.Date(daysPerYear-1).DayInYear(), "Last day of year -1")
	assert.Equal(t, 1, calendar.Date(daysPerYear).DayInYear(), "First day of year 1")
	assert.Equal(t, daysPerYear, calendar.Date(daysPerYear*2-1).DayInYear(), "Last day of year 1")
	assert.Equal(t, 1, calendar.Date(daysPerYear*2).DayInYear(), "First day of year 2")
}

func TestMonth(t *testing.T) {
	assert.Equal(t, 1, calendar.Date(0).Month(), "First day of first month")
	assert.Equal(t, 1, calendar.Date(30).Month(), "Last day of first month")
	assert.Equal(t, 2, calendar.Date(31).Month(), "First day of second month")
	assert.Equal(t, 2, calendar.Date(58).Month(), "Last day of second month")
	assert.Equal(t, 3, calendar.Date(59).Month(), "First day of third month")
	assert.Equal(t, 12, calendar.Date(364).Month(), "Last day of last month")
	assert.Equal(t, 1, calendar.Date(365).Month(), "First day of first month of year 1")
}

func TestDayInMonth(t *testing.T) {
	assert.Equal(t, 1, calendar.Date(0).DayInMonth(), "First day of first month")
	assert.Equal(t, 31, calendar.Date(30).DayInMonth(), "Last day of first month")
	assert.Equal(t, 1, calendar.Date(31).DayInMonth(), "First day of second month")
	assert.Equal(t, 28, calendar.Date(58).DayInMonth(), "Last day of second month")
	assert.Equal(t, 1, calendar.Date(59).DayInMonth(), "First day of third month")
	assert.Equal(t, 31, calendar.Date(364).DayInMonth(), "Last day of last month")
	assert.Equal(t, 1, calendar.Date(365).DayInMonth(), "First day of first month of year 1")
}

func TestWeekDay(t *testing.T) {
	assert.Equal(t, 6, calendar.Date(0).WeekDay(), "Saturday, first day of year 0")
	assert.Equal(t, 2, calendar.Date(3).WeekDay(), "Tuesday, fourth day of year 0")
	assert.Equal(t, 6, calendar.Date(7).WeekDay(), "Saturday, first day of year 0 plus one week")
	assert.Equal(t, 5, calendar.Date(-1).WeekDay(), "Friday, last day of year -1")
	assert.Equal(t, 4, calendar.Date(-2).WeekDay(), "Thursday, second-to-last day of year -1")
	assert.Equal(t, 5, calendar.Date(-8).WeekDay(), "Friday, last day of year -1 minus one week")
	assert.Equal(t, 0, calendar.Date(736205).WeekDay(), "Sunday, 1/1/2017")
}

func TestDateToString(t *testing.T) {
	assert.Equal(t, "1/1/2017", calendar.Date(736205).String())
	assert.Equal(t, "9/22/2017", calendar.Date(736469).String())
	assert.Equal(t, "12/31/-1", calendar.Date(364).String())
	assert.Equal(t, "12/31/-2", calendar.Date(-1).String())
	assert.Equal(t, "1/1/-1", calendar.Date(0).String())
	assert.Equal(t, "1/1/1", calendar.Date(365).String())
	assert.Equal(t, "1/1/1", calendar.Date(365).String())
}

func TestFormat(t *testing.T) {
	d := calendar.MustNewDate(9, 22, 2017)
	assert.Equal(t, "September 22, 2017", d.Format(false, false, false))
	assert.Equal(t, "September 22, 2017 AD", d.Format(false, false, true))
	assert.Equal(t, "Sep 22, 2017", d.Format(false, true, false))
	assert.Equal(t, "Sep 22, 2017 AD", d.Format(false, true, true))
	assert.Equal(t, "Friday, September 22, 2017", d.Format(true, false, false))
	assert.Equal(t, "Friday, September 22, 2017 AD", d.Format(true, false, true))
	assert.Equal(t, "Friday, Sep 22, 2017", d.Format(true, true, false))
	assert.Equal(t, "Friday, Sep 22, 2017 AD", d.Format(true, true, true))
}

func TestNewDate(t *testing.T) {
	date, err := calendar.NewDate(1, 1, 1)
	assert.NoError(t, err)
	assert.Equal(t, 365, int(date))
	_, err = calendar.NewDate(1, 1, 0)
	assert.Error(t, err)
	date, err = calendar.NewDate(9, 22, 2017)
	assert.NoError(t, err)
	assert.Equal(t, 736469, int(date))
	_, err = calendar.NewDate(13, 22, 2017)
	assert.Error(t, err)
	_, err = calendar.NewDate(9, 888, 2017)
	assert.Error(t, err)
}

func TestParseDate(t *testing.T) {
	targetDate := calendar.MustNewDate(9, 22, 2017)
	date, err := calendar.ParseDate("A long, rambling prefix September 22, 2017 and a long suffix")
	assert.NoError(t, err)
	assert.Equal(t, targetDate, date)
	date, err = calendar.ParseDate("Friday, September 22, 2017")
	assert.NoError(t, err)
	assert.Equal(t, targetDate, date)
	date, err = calendar.ParseDate("September 22, 2017")
	assert.NoError(t, err)
	assert.Equal(t, targetDate, date)
	date, err = calendar.ParseDate("9/22/2017")
	assert.NoError(t, err)
	assert.Equal(t, targetDate, date)
	date, err = calendar.ParseDate("what 9/22/2017 how?")
	assert.NoError(t, err)
	assert.Equal(t, targetDate, date)
	_, err = calendar.ParseDate("9/22")
	assert.Error(t, err)
	_, err = calendar.ParseDate("9/666/2017")
	assert.Error(t, err)
	_, err = calendar.ParseDate("13/22/2017")
	assert.Error(t, err)
	date, err = calendar.ParseDate("September 22, 2017 AD")
	assert.NoError(t, err)
	assert.Equal(t, targetDate, date)
	date, err = calendar.ParseDate("September 22, 1 BC")
	assert.NoError(t, err)
	assert.Equal(t, calendar.MustNewDate(9, 22, -1), date)

	targetDate = calendar.MustNewDate(9, 22, -2017)
	date, err = calendar.ParseDate("9/22/-2017")
	assert.NoError(t, err)
	assert.Equal(t, targetDate, date)
	date, err = calendar.ParseDate("September 22, -2017")
	assert.NoError(t, err)
	assert.Equal(t, targetDate, date)
}

func TestMarshaling(t *testing.T) {
	date := calendar.MustNewDate(9, 22, 2017)
	text, err := date.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, "9/22/2017", string(text))

	type embedded struct {
		Date calendar.Date
	}
	embeddedDate := embedded{Date: date}
	text, err = json.Marshal(&embeddedDate)
	assert.NoError(t, err)
	assert.Equal(t, `{"Date":"9/22/2017"}`, string(text))

	type embeddedPtr struct {
		Date *calendar.Date
	}
	embeddedPtrDate := embeddedPtr{Date: &date}
	text, err = json.Marshal(&embeddedPtrDate)
	assert.NoError(t, err)
	assert.Equal(t, `{"Date":"9/22/2017"}`, string(text))
}

func TestUnmarshaling(t *testing.T) {
	target := calendar.MustNewDate(9, 22, 2017)
	var date calendar.Date
	assert.NoError(t, date.UnmarshalText([]byte("9/22/2017")))
	assert.Equal(t, target, date)

	type embedded struct {
		Date calendar.Date
	}
	var embeddedDate embedded
	assert.NoError(t, json.Unmarshal([]byte(`{"Date":"9/22/2017"}`), &embeddedDate))
	assert.Equal(t, target, embeddedDate.Date)

	type embeddedPtr struct {
		Date *calendar.Date
	}
	var embeddedPtrDate embeddedPtr
	assert.NoError(t, json.Unmarshal([]byte(`{"Date":"9/22/2017"}`), &embeddedPtrDate))
	assert.Equal(t, target, *embeddedPtrDate.Date)
}
