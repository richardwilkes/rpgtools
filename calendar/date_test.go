// Copyright ©2017-2020 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calendar_test

import (
	"encoding/json"
	"testing"

	"github.com/richardwilkes/rpgtools/calendar"
	"github.com/stretchr/testify/assert"
)

func TestNewDate(t *testing.T) {
	cal := calendar.Gregorian()
	d, err := cal.NewDate(1, 1, 1)
	assert.NoError(t, err)
	assert.Equal(t, cal.NewDateByDays(0), d)
	d, err = cal.NewDate(12, 31, 1)
	assert.NoError(t, err)
	assert.Equal(t, cal.NewDateByDays(364), d)
	d, err = cal.NewDate(1, 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, cal.NewDateByDays(365), d)

	d, err = cal.NewDate(1, 1, -1)
	assert.NoError(t, err)
	assert.Equal(t, cal.NewDateByDays(-366), d)
	d, err = cal.NewDate(12, 31, -1)
	assert.NoError(t, err)
	assert.Equal(t, cal.NewDateByDays(-1), d)
	d, err = cal.NewDate(1, 1, -2)
	assert.NoError(t, err)
	assert.Equal(t, cal.NewDateByDays(-731), d)
	d, err = cal.NewDate(12, 31, -2)
	assert.NoError(t, err)
	assert.Equal(t, cal.NewDateByDays(-367), d)
	d, err = cal.NewDate(12, 31, -3)
	assert.NoError(t, err)
	assert.Equal(t, cal.NewDateByDays(-732), d)

	_, err = cal.NewDate(1, 1, 0)
	assert.Error(t, err)
	_, err = cal.NewDate(13, 22, 2017)
	assert.Error(t, err)
	_, err = cal.NewDate(9, 888, 2017)
	assert.Error(t, err)
}

func TestYear(t *testing.T) {
	cal := calendar.Gregorian()
	assert.Equal(t, 1, cal.NewDateByDays(0).Year(), "First day of year 1")
	assert.Equal(t, 1, cal.NewDateByDays(364).Year(), "Last day of year 1")
	assert.Equal(t, 2, cal.NewDateByDays(365).Year(), "First day of year 2")

	assert.Equal(t, -1, cal.NewDateByDays(-366).Year(), "First day of year -1")
	assert.Equal(t, -1, cal.NewDateByDays(-1).Year(), "Last day of year -1")
	assert.Equal(t, -2, cal.NewDateByDays(-731).Year(), "First day of year -2")
	assert.Equal(t, -2, cal.NewDateByDays(-367).Year(), "Last day of year -2")
	assert.Equal(t, -3, cal.NewDateByDays(-732).Year(), "Last day of year -3")

	for year := 1; year < 5000; year++ {
		assert.Equal(t, year, cal.MustNewDate(1, 1, year).Year())
		assert.Equal(t, year, cal.MustNewDate(12, 31, year).Year())
	}

	for year := -1; year > -5000; year-- {
		assert.Equal(t, year, cal.MustNewDate(1, 1, year).Year())
		assert.Equal(t, year, cal.MustNewDate(12, 31, year).Year())
	}
}

func TestDayInYear(t *testing.T) {
	cal := calendar.Gregorian()
	assert.Equal(t, 1, cal.MustNewDate(1, 1, 1).DayInYear())
	assert.Equal(t, 365, cal.MustNewDate(12, 31, 1).DayInYear())
	assert.Equal(t, 1, cal.MustNewDate(1, 1, 2).DayInYear())
	assert.Equal(t, 1, cal.MustNewDate(1, 1, 4).DayInYear())
	assert.Equal(t, 366, cal.MustNewDate(12, 31, 4).DayInYear())

	assert.Equal(t, 1, cal.MustNewDate(1, 1, -1).DayInYear())
	assert.Equal(t, 366, cal.MustNewDate(12, 31, -1).DayInYear())
	assert.Equal(t, 1, cal.MustNewDate(1, 1, -2).DayInYear())
	assert.Equal(t, 365, cal.MustNewDate(12, 31, -2).DayInYear())
	assert.Equal(t, 1, cal.MustNewDate(1, 1, -5).DayInYear())
	assert.Equal(t, 366, cal.MustNewDate(12, 31, -5).DayInYear())
}

func TestMonth(t *testing.T) {
	cal := calendar.Gregorian()
	assert.Equal(t, 1, cal.MustNewDate(1, 1, 1).Month())
	assert.Equal(t, 1, cal.MustNewDate(1, 31, 1).Month())
	assert.Equal(t, 2, cal.MustNewDate(2, 1, 1).Month())
	assert.Equal(t, 2, cal.MustNewDate(2, 28, 1).Month())
	assert.Equal(t, 3, cal.MustNewDate(3, 1, 1).Month())
	assert.Equal(t, 12, cal.MustNewDate(12, 31, 1).Month())
	assert.Equal(t, 1, cal.MustNewDate(1, 1, 2).Month())
	assert.Equal(t, 2, cal.MustNewDate(2, 28, 4).Month())
	assert.Equal(t, 2, cal.MustNewDate(2, 29, 4).Month())
	assert.Equal(t, 3, cal.MustNewDate(3, 1, 4).Month())

	assert.Equal(t, 2, cal.MustNewDate(2, 29, -1).Month())
	assert.Equal(t, 2, cal.MustNewDate(2, 28, -2).Month())
}

func TestDayInMonth(t *testing.T) {
	cal := calendar.Gregorian()
	assert.Equal(t, 1, cal.MustNewDate(1, 1, 1).DayInMonth())
	assert.Equal(t, 31, cal.MustNewDate(1, 31, 1).DayInMonth())
	assert.Equal(t, 1, cal.MustNewDate(2, 1, 1).DayInMonth())
	assert.Equal(t, 28, cal.MustNewDate(2, 28, 1).DayInMonth())
	assert.Equal(t, 1, cal.MustNewDate(3, 1, 1).DayInMonth())
	assert.Equal(t, 31, cal.MustNewDate(12, 31, 1).DayInMonth())
	assert.Equal(t, 1, cal.MustNewDate(1, 1, 2).DayInMonth())
	assert.Equal(t, 28, cal.MustNewDate(2, 28, 2).DayInMonth())
	assert.Equal(t, 1, cal.MustNewDate(3, 1, 2).DayInMonth())
	assert.Equal(t, 28, cal.MustNewDate(2, 28, 4).DayInMonth())
	assert.Equal(t, 29, cal.MustNewDate(2, 29, 4).DayInMonth())
	assert.Equal(t, 1, cal.MustNewDate(3, 1, 4).DayInMonth())

	assert.Equal(t, 29, cal.MustNewDate(2, 29, -1).DayInMonth())
	assert.Equal(t, 28, cal.MustNewDate(2, 28, -2).DayInMonth())
}

func TestDateToString(t *testing.T) {
	cal := calendar.Gregorian()
	assert.Equal(t, "1/1/1", cal.MustNewDate(1, 1, 1).String())
	assert.Equal(t, "12/31/1", cal.MustNewDate(12, 31, 1).String())
	assert.Equal(t, "1/1/2", cal.MustNewDate(1, 1, 2).String())
	assert.Equal(t, "1/1/2017", cal.MustNewDate(1, 1, 2017).String())
	assert.Equal(t, "9/22/2017", cal.MustNewDate(9, 22, 2017).String())

	assert.Equal(t, "1/1/1 BC", cal.MustNewDate(1, 1, -1).String())
	assert.Equal(t, "12/31/1 BC", cal.MustNewDate(12, 31, -1).String())
	assert.Equal(t, "1/1/2 BC", cal.MustNewDate(1, 1, -2).String())
	assert.Equal(t, "12/31/2 BC", cal.MustNewDate(12, 31, -2).String())
	assert.Equal(t, "12/31/3 BC", cal.MustNewDate(12, 31, -3).String())
}

func TestWeekDay(t *testing.T) {
	cal := calendar.Gregorian()
	assert.Equal(t, 1, cal.MustNewDate(1, 1, 1).WeekDay())
	assert.Equal(t, 4, cal.MustNewDate(1, 4, 1).WeekDay())
	assert.Equal(t, 1, cal.MustNewDate(1, 8, 1).WeekDay())
	assert.Equal(t, 0, cal.MustNewDate(12, 31, -1).WeekDay())
	assert.Equal(t, 6, cal.MustNewDate(12, 30, -1).WeekDay())
	assert.Equal(t, 0, cal.MustNewDate(12, 24, -1).WeekDay())
	assert.Equal(t, 6, cal.MustNewDate(1, 1, 2000).WeekDay())
	assert.Equal(t, 1, cal.MustNewDate(9, 3, 2018).WeekDay())
}

func TestFormat(t *testing.T) {
	cal := calendar.Gregorian()
	d := cal.MustNewDate(9, 22, 2017)
	assert.Equal(t, "9/22/2017", d.Format(calendar.ShortFormat))
	assert.Equal(t, "Sep 22, 2017", d.Format(calendar.MediumFormat))
	assert.Equal(t, "September 22, 2017", d.Format(calendar.LongFormat))
	assert.Equal(t, "Friday, September 22, 2017", d.Format(calendar.FullFormat))
	assert.Equal(t, "%Fri%", d.Format("%%%w%%"))
	assert.Equal(t, "Friday, September 22, 2017 AD", d.Format("%W, %M %D, %y"))

	d = cal.MustNewDate(9, 22, -1)
	assert.Equal(t, "9/22/1 BC", d.Format(calendar.ShortFormat))
	assert.Equal(t, "Sep 22, 1 BC", d.Format(calendar.MediumFormat))
	assert.Equal(t, "September 22, 1 BC", d.Format(calendar.LongFormat))
	assert.Equal(t, "Friday, September 22, 1 BC", d.Format(calendar.FullFormat))
	assert.Equal(t, "%Fri%", d.Format("%%%w%%"))
	assert.Equal(t, "Friday, September 22, 1 BC", d.Format("%W, %M %D, %y"))
}

func TestParseDate(t *testing.T) {
	cal := calendar.Gregorian()
	targetDate := cal.MustNewDate(9, 22, 2017)
	date, err := cal.ParseDate("A long, rambling prefix September 22, 2017 and a long suffix")
	assert.NoError(t, err)
	assert.Equal(t, targetDate, date)
	date, err = cal.ParseDate("Friday, September 22, 2017")
	assert.NoError(t, err)
	assert.Equal(t, targetDate, date)
	date, err = cal.ParseDate("September 22, 2017")
	assert.NoError(t, err)
	assert.Equal(t, targetDate, date)
	date, err = cal.ParseDate("9/22/2017")
	assert.NoError(t, err)
	assert.Equal(t, targetDate, date)
	date, err = cal.ParseDate("what 9/22/2017 how?")
	assert.NoError(t, err)
	assert.Equal(t, targetDate, date)
	_, err = cal.ParseDate("9/22")
	assert.Error(t, err)
	_, err = cal.ParseDate("9/666/2017")
	assert.Error(t, err)
	_, err = cal.ParseDate("13/22/2017")
	assert.Error(t, err)
	date, err = cal.ParseDate("September 22, 2017 AD")
	assert.NoError(t, err)
	assert.Equal(t, targetDate, date)
	date, err = cal.ParseDate("September 22, 1 BC")
	assert.NoError(t, err)
	assert.Equal(t, cal.MustNewDate(9, 22, -1), date)

	targetDate = cal.MustNewDate(9, 22, -2017)
	date, err = cal.ParseDate("9/22/-2017")
	assert.NoError(t, err)
	assert.Equal(t, targetDate, date)
	date, err = cal.ParseDate("September 22, -2017")
	assert.NoError(t, err)
	assert.Equal(t, targetDate, date)
}

func TestMarshaling(t *testing.T) {
	cal := calendar.Gregorian()
	date := cal.MustNewDate(9, 22, 2017)
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
	cal := calendar.Gregorian()
	target := cal.MustNewDate(9, 22, 2017)
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
