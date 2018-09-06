package calendar_test

import (
	"encoding/json"
	"testing"

	"github.com/richardwilkes/rpgtools/calendar"
	"github.com/stretchr/testify/assert"
)

func TestNewDate(t *testing.T) {
	d, err := calendar.NewDate(1, 1, 1)
	assert.NoError(t, err)
	assert.Equal(t, calendar.Date(0), d)
	d, err = calendar.NewDate(12, 31, 1)
	assert.NoError(t, err)
	assert.Equal(t, calendar.Date(364), d)
	d, err = calendar.NewDate(1, 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, calendar.Date(365), d)

	d, err = calendar.NewDate(1, 1, -1)
	assert.NoError(t, err)
	assert.Equal(t, calendar.Date(-366), d)
	d, err = calendar.NewDate(12, 31, -1)
	assert.NoError(t, err)
	assert.Equal(t, calendar.Date(-1), d)
	d, err = calendar.NewDate(1, 1, -2)
	assert.NoError(t, err)
	assert.Equal(t, calendar.Date(-731), d)
	d, err = calendar.NewDate(12, 31, -2)
	assert.NoError(t, err)
	assert.Equal(t, calendar.Date(-367), d)
	d, err = calendar.NewDate(12, 31, -3)
	assert.NoError(t, err)
	assert.Equal(t, calendar.Date(-732), d)

	_, err = calendar.NewDate(1, 1, 0)
	assert.Error(t, err)
	_, err = calendar.NewDate(13, 22, 2017)
	assert.Error(t, err)
	_, err = calendar.NewDate(9, 888, 2017)
	assert.Error(t, err)
}

func TestYear(t *testing.T) {
	assert.Equal(t, 1, calendar.Date(0).Year(), "First day of year 1")
	assert.Equal(t, 1, calendar.Date(364).Year(), "Last day of year 1")
	assert.Equal(t, 2, calendar.Date(365).Year(), "First day of year 2")

	assert.Equal(t, -1, calendar.Date(-366).Year(), "First day of year -1")
	assert.Equal(t, -1, calendar.Date(-1).Year(), "Last day of year -1")
	assert.Equal(t, -2, calendar.Date(-731).Year(), "First day of year -2")
	assert.Equal(t, -2, calendar.Date(-367).Year(), "Last day of year -2")
	assert.Equal(t, -3, calendar.Date(-732).Year(), "Last day of year -3")

	for year := 1; year < 5000; year++ {
		assert.Equal(t, year, calendar.MustNewDate(1, 1, year).Year())
		assert.Equal(t, year, calendar.MustNewDate(12, 31, year).Year())
	}

	for year := -1; year > -5000; year-- {
		assert.Equal(t, year, calendar.MustNewDate(1, 1, year).Year())
		assert.Equal(t, year, calendar.MustNewDate(12, 31, year).Year())
	}
}

func TestDayInYear(t *testing.T) {
	assert.Equal(t, 1, calendar.MustNewDate(1, 1, 1).DayInYear())
	assert.Equal(t, 365, calendar.MustNewDate(12, 31, 1).DayInYear())
	assert.Equal(t, 1, calendar.MustNewDate(1, 1, 2).DayInYear())
	assert.Equal(t, 1, calendar.MustNewDate(1, 1, 4).DayInYear())
	assert.Equal(t, 366, calendar.MustNewDate(12, 31, 4).DayInYear())

	assert.Equal(t, 1, calendar.MustNewDate(1, 1, -1).DayInYear())
	assert.Equal(t, 366, calendar.MustNewDate(12, 31, -1).DayInYear())
	assert.Equal(t, 1, calendar.MustNewDate(1, 1, -2).DayInYear())
	assert.Equal(t, 365, calendar.MustNewDate(12, 31, -2).DayInYear())
	assert.Equal(t, 1, calendar.MustNewDate(1, 1, -5).DayInYear())
	assert.Equal(t, 366, calendar.MustNewDate(12, 31, -5).DayInYear())
}

func TestMonth(t *testing.T) {
	assert.Equal(t, 1, calendar.MustNewDate(1, 1, 1).Month())
	assert.Equal(t, 1, calendar.MustNewDate(1, 31, 1).Month())
	assert.Equal(t, 2, calendar.MustNewDate(2, 1, 1).Month())
	assert.Equal(t, 2, calendar.MustNewDate(2, 28, 1).Month())
	assert.Equal(t, 3, calendar.MustNewDate(3, 1, 1).Month())
	assert.Equal(t, 12, calendar.MustNewDate(12, 31, 1).Month())
	assert.Equal(t, 1, calendar.MustNewDate(1, 1, 2).Month())
	assert.Equal(t, 2, calendar.MustNewDate(2, 28, 4).Month())
	assert.Equal(t, 2, calendar.MustNewDate(2, 29, 4).Month())
	assert.Equal(t, 3, calendar.MustNewDate(3, 1, 4).Month())

	assert.Equal(t, 2, calendar.MustNewDate(2, 29, -1).Month())
	assert.Equal(t, 2, calendar.MustNewDate(2, 28, -2).Month())
}

func TestDayInMonth(t *testing.T) {
	assert.Equal(t, 1, calendar.MustNewDate(1, 1, 1).DayInMonth())
	assert.Equal(t, 31, calendar.MustNewDate(1, 31, 1).DayInMonth())
	assert.Equal(t, 1, calendar.MustNewDate(2, 1, 1).DayInMonth())
	assert.Equal(t, 28, calendar.MustNewDate(2, 28, 1).DayInMonth())
	assert.Equal(t, 1, calendar.MustNewDate(3, 1, 1).DayInMonth())
	assert.Equal(t, 31, calendar.MustNewDate(12, 31, 1).DayInMonth())
	assert.Equal(t, 1, calendar.MustNewDate(1, 1, 2).DayInMonth())
	assert.Equal(t, 28, calendar.MustNewDate(2, 28, 2).DayInMonth())
	assert.Equal(t, 1, calendar.MustNewDate(3, 1, 2).DayInMonth())
	assert.Equal(t, 28, calendar.MustNewDate(2, 28, 4).DayInMonth())
	assert.Equal(t, 29, calendar.MustNewDate(2, 29, 4).DayInMonth())
	assert.Equal(t, 1, calendar.MustNewDate(3, 1, 4).DayInMonth())

	assert.Equal(t, 29, calendar.MustNewDate(2, 29, -1).DayInMonth())
	assert.Equal(t, 28, calendar.MustNewDate(2, 28, -2).DayInMonth())
}

func TestDateToString(t *testing.T) {
	assert.Equal(t, "1/1/1", calendar.MustNewDate(1, 1, 1).String())
	assert.Equal(t, "12/31/1", calendar.MustNewDate(12, 31, 1).String())
	assert.Equal(t, "1/1/2", calendar.MustNewDate(1, 1, 2).String())
	assert.Equal(t, "1/1/2017", calendar.MustNewDate(1, 1, 2017).String())
	assert.Equal(t, "9/22/2017", calendar.MustNewDate(9, 22, 2017).String())

	assert.Equal(t, "1/1/1 BC", calendar.MustNewDate(1, 1, -1).String())
	assert.Equal(t, "12/31/1 BC", calendar.MustNewDate(12, 31, -1).String())
	assert.Equal(t, "1/1/2 BC", calendar.MustNewDate(1, 1, -2).String())
	assert.Equal(t, "12/31/2 BC", calendar.MustNewDate(12, 31, -2).String())
	assert.Equal(t, "12/31/3 BC", calendar.MustNewDate(12, 31, -3).String())
}

func TestWeekDay(t *testing.T) {
	assert.Equal(t, 1, calendar.MustNewDate(1, 1, 1).WeekDay())
	assert.Equal(t, 4, calendar.MustNewDate(1, 4, 1).WeekDay())
	assert.Equal(t, 1, calendar.MustNewDate(1, 8, 1).WeekDay())
	assert.Equal(t, 0, calendar.MustNewDate(12, 31, -1).WeekDay())
	assert.Equal(t, 6, calendar.MustNewDate(12, 30, -1).WeekDay())
	assert.Equal(t, 0, calendar.MustNewDate(12, 24, -1).WeekDay())
	assert.Equal(t, 6, calendar.MustNewDate(1, 1, 2000).WeekDay())
	assert.Equal(t, 1, calendar.MustNewDate(9, 3, 2018).WeekDay())
}

func TestFormat(t *testing.T) {
	d := calendar.MustNewDate(9, 22, 2017)
	assert.Equal(t, "9/22/2017", d.Format(calendar.ShortFormat))
	assert.Equal(t, "Sep 22, 2017", d.Format(calendar.MediumFormat))
	assert.Equal(t, "September 22, 2017", d.Format(calendar.LongFormat))
	assert.Equal(t, "Friday, September 22, 2017", d.Format(calendar.FullFormat))
	assert.Equal(t, "%Fri%", d.Format("%%%w%%"))
	assert.Equal(t, "Friday, September 22, 2017 AD", d.Format("%W, %M %D, %y"))

	d = calendar.MustNewDate(9, 22, -1)
	assert.Equal(t, "9/22/1 BC", d.Format(calendar.ShortFormat))
	assert.Equal(t, "Sep 22, 1 BC", d.Format(calendar.MediumFormat))
	assert.Equal(t, "September 22, 1 BC", d.Format(calendar.LongFormat))
	assert.Equal(t, "Friday, September 22, 1 BC", d.Format(calendar.FullFormat))
	assert.Equal(t, "%Fri%", d.Format("%%%w%%"))
	assert.Equal(t, "Friday, September 22, 1 BC", d.Format("%W, %M %D, %y"))
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
