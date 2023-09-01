// Copyright ©2017-2023 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/rpgtools/calendar/pathfinder"
	"github.com/richardwilkes/toolbox/check"
	"gopkg.in/yaml.v3"
)

func TestNewDate(t *testing.T) {
	cal := calendar.Gregorian()
	calendar.Default = cal
	d, err := cal.NewDate(1, 1, 1)
	check.NoError(t, err)
	check.Equal(t, cal.NewDateByDays(0), d)
	d, err = cal.NewDate(12, 31, 1)
	check.NoError(t, err)
	check.Equal(t, cal.NewDateByDays(364), d)
	d, err = cal.NewDate(1, 1, 2)
	check.NoError(t, err)
	check.Equal(t, cal.NewDateByDays(365), d)

	d, err = cal.NewDate(1, 1, -1)
	check.NoError(t, err)
	check.Equal(t, cal.NewDateByDays(-366), d)
	d, err = cal.NewDate(12, 31, -1)
	check.NoError(t, err)
	check.Equal(t, cal.NewDateByDays(-1), d)
	d, err = cal.NewDate(1, 1, -2)
	check.NoError(t, err)
	check.Equal(t, cal.NewDateByDays(-731), d)
	d, err = cal.NewDate(12, 31, -2)
	check.NoError(t, err)
	check.Equal(t, cal.NewDateByDays(-367), d)
	d, err = cal.NewDate(12, 31, -3)
	check.NoError(t, err)
	check.Equal(t, cal.NewDateByDays(-732), d)

	_, err = cal.NewDate(1, 1, 0)
	check.Error(t, err)
	_, err = cal.NewDate(13, 22, 2017)
	check.Error(t, err)
	_, err = cal.NewDate(9, 888, 2017)
	check.Error(t, err)
}

func TestYear(t *testing.T) {
	cal := calendar.Gregorian()
	calendar.Default = cal
	check.Equal(t, 1, cal.NewDateByDays(0).Year(), "First day of year 1")
	check.Equal(t, 1, cal.NewDateByDays(364).Year(), "Last day of year 1")
	check.Equal(t, 2, cal.NewDateByDays(365).Year(), "First day of year 2")

	check.Equal(t, -1, cal.NewDateByDays(-366).Year(), "First day of year -1")
	check.Equal(t, -1, cal.NewDateByDays(-1).Year(), "Last day of year -1")
	check.Equal(t, -2, cal.NewDateByDays(-731).Year(), "First day of year -2")
	check.Equal(t, -2, cal.NewDateByDays(-367).Year(), "Last day of year -2")
	check.Equal(t, -3, cal.NewDateByDays(-732).Year(), "Last day of year -3")

	for year := 1; year < 5000; year++ {
		check.Equal(t, year, cal.MustNewDate(1, 1, year).Year())
		check.Equal(t, year, cal.MustNewDate(12, 31, year).Year())
	}

	for year := -1; year > -5000; year-- {
		check.Equal(t, year, cal.MustNewDate(1, 1, year).Year())
		check.Equal(t, year, cal.MustNewDate(12, 31, year).Year())
	}
}

func TestDayInYear(t *testing.T) {
	cal := calendar.Gregorian()
	calendar.Default = cal
	check.Equal(t, 1, cal.MustNewDate(1, 1, 1).DayInYear())
	check.Equal(t, 365, cal.MustNewDate(12, 31, 1).DayInYear())
	check.Equal(t, 1, cal.MustNewDate(1, 1, 2).DayInYear())
	check.Equal(t, 1, cal.MustNewDate(1, 1, 4).DayInYear())
	check.Equal(t, 366, cal.MustNewDate(12, 31, 4).DayInYear())

	check.Equal(t, 1, cal.MustNewDate(1, 1, -1).DayInYear())
	check.Equal(t, 366, cal.MustNewDate(12, 31, -1).DayInYear())
	check.Equal(t, 1, cal.MustNewDate(1, 1, -2).DayInYear())
	check.Equal(t, 365, cal.MustNewDate(12, 31, -2).DayInYear())
	check.Equal(t, 1, cal.MustNewDate(1, 1, -5).DayInYear())
	check.Equal(t, 366, cal.MustNewDate(12, 31, -5).DayInYear())
}

func TestMonth(t *testing.T) {
	cal := calendar.Gregorian()
	calendar.Default = cal
	check.Equal(t, 1, cal.MustNewDate(1, 1, 1).Month())
	check.Equal(t, 1, cal.MustNewDate(1, 31, 1).Month())
	check.Equal(t, 2, cal.MustNewDate(2, 1, 1).Month())
	check.Equal(t, 2, cal.MustNewDate(2, 28, 1).Month())
	check.Equal(t, 3, cal.MustNewDate(3, 1, 1).Month())
	check.Equal(t, 12, cal.MustNewDate(12, 31, 1).Month())
	check.Equal(t, 1, cal.MustNewDate(1, 1, 2).Month())
	check.Equal(t, 2, cal.MustNewDate(2, 28, 4).Month())
	check.Equal(t, 2, cal.MustNewDate(2, 29, 4).Month())
	check.Equal(t, 3, cal.MustNewDate(3, 1, 4).Month())

	check.Equal(t, 2, cal.MustNewDate(2, 29, -1).Month())
	check.Equal(t, 2, cal.MustNewDate(2, 28, -2).Month())
}

func TestDayInMonth(t *testing.T) {
	cal := calendar.Gregorian()
	calendar.Default = cal
	check.Equal(t, 1, cal.MustNewDate(1, 1, 1).DayInMonth())
	check.Equal(t, 31, cal.MustNewDate(1, 31, 1).DayInMonth())
	check.Equal(t, 1, cal.MustNewDate(2, 1, 1).DayInMonth())
	check.Equal(t, 28, cal.MustNewDate(2, 28, 1).DayInMonth())
	check.Equal(t, 1, cal.MustNewDate(3, 1, 1).DayInMonth())
	check.Equal(t, 31, cal.MustNewDate(12, 31, 1).DayInMonth())
	check.Equal(t, 1, cal.MustNewDate(1, 1, 2).DayInMonth())
	check.Equal(t, 28, cal.MustNewDate(2, 28, 2).DayInMonth())
	check.Equal(t, 1, cal.MustNewDate(3, 1, 2).DayInMonth())
	check.Equal(t, 28, cal.MustNewDate(2, 28, 4).DayInMonth())
	check.Equal(t, 29, cal.MustNewDate(2, 29, 4).DayInMonth())
	check.Equal(t, 1, cal.MustNewDate(3, 1, 4).DayInMonth())

	check.Equal(t, 29, cal.MustNewDate(2, 29, -1).DayInMonth())
	check.Equal(t, 28, cal.MustNewDate(2, 28, -2).DayInMonth())
}

func TestDateToString(t *testing.T) {
	cal := calendar.Gregorian()
	calendar.Default = cal
	check.Equal(t, "1/1/1", cal.MustNewDate(1, 1, 1).String())
	check.Equal(t, "12/31/1", cal.MustNewDate(12, 31, 1).String())
	check.Equal(t, "1/1/2", cal.MustNewDate(1, 1, 2).String())
	check.Equal(t, "1/1/2017", cal.MustNewDate(1, 1, 2017).String())
	check.Equal(t, "9/22/2017", cal.MustNewDate(9, 22, 2017).String())

	check.Equal(t, "1/1/1 BC", cal.MustNewDate(1, 1, -1).String())
	check.Equal(t, "12/31/1 BC", cal.MustNewDate(12, 31, -1).String())
	check.Equal(t, "1/1/2 BC", cal.MustNewDate(1, 1, -2).String())
	check.Equal(t, "12/31/2 BC", cal.MustNewDate(12, 31, -2).String())
	check.Equal(t, "12/31/3 BC", cal.MustNewDate(12, 31, -3).String())
}

func TestWeekDay(t *testing.T) {
	cal := calendar.Gregorian()
	calendar.Default = cal
	check.Equal(t, 1, cal.MustNewDate(1, 1, 1).WeekDay())
	check.Equal(t, 4, cal.MustNewDate(1, 4, 1).WeekDay())
	check.Equal(t, 1, cal.MustNewDate(1, 8, 1).WeekDay())
	check.Equal(t, 0, cal.MustNewDate(12, 31, -1).WeekDay())
	check.Equal(t, 6, cal.MustNewDate(12, 30, -1).WeekDay())
	check.Equal(t, 0, cal.MustNewDate(12, 24, -1).WeekDay())
	check.Equal(t, 6, cal.MustNewDate(1, 1, 2000).WeekDay())
	check.Equal(t, 1, cal.MustNewDate(9, 3, 2018).WeekDay())
}

func TestFormat(t *testing.T) {
	cal := calendar.Gregorian()
	calendar.Default = cal
	d := cal.MustNewDate(9, 22, 2017)
	check.Equal(t, "9/22/2017", d.Format(calendar.ShortFormat))
	check.Equal(t, "Sep 22, 2017", d.Format(calendar.MediumFormat))
	check.Equal(t, "September 22, 2017", d.Format(calendar.LongFormat))
	check.Equal(t, "Friday, September 22, 2017", d.Format(calendar.FullFormat))
	check.Equal(t, "%Fri%", d.Format("%%%w%%"))
	check.Equal(t, "Friday, September 22, 2017 AD", d.Format("%W, %M %D, %y"))

	d = cal.MustNewDate(9, 22, -1)
	check.Equal(t, "9/22/1 BC", d.Format(calendar.ShortFormat))
	check.Equal(t, "Sep 22, 1 BC", d.Format(calendar.MediumFormat))
	check.Equal(t, "September 22, 1 BC", d.Format(calendar.LongFormat))
	check.Equal(t, "Friday, September 22, 1 BC", d.Format(calendar.FullFormat))
	check.Equal(t, "%Fri%", d.Format("%%%w%%"))
	check.Equal(t, "Friday, September 22, 1 BC", d.Format("%W, %M %D, %y"))
}

func TestParseDate(t *testing.T) {
	cal := calendar.Gregorian()
	calendar.Default = cal
	targetDate := cal.MustNewDate(9, 22, 2017)
	date, err := cal.ParseDate("A long, rambling prefix September 22, 2017 and a long suffix")
	check.NoError(t, err)
	check.Equal(t, targetDate, date)
	date, err = cal.ParseDate("Friday, September 22, 2017")
	check.NoError(t, err)
	check.Equal(t, targetDate, date)
	date, err = cal.ParseDate("September 22, 2017")
	check.NoError(t, err)
	check.Equal(t, targetDate, date)
	date, err = cal.ParseDate("9/22/2017")
	check.NoError(t, err)
	check.Equal(t, targetDate, date)
	date, err = cal.ParseDate("what 9/22/2017 how?")
	check.NoError(t, err)
	check.Equal(t, targetDate, date)
	_, err = cal.ParseDate("9/22")
	check.Error(t, err)
	_, err = cal.ParseDate("9/666/2017")
	check.Error(t, err)
	_, err = cal.ParseDate("13/22/2017")
	check.Error(t, err)
	date, err = cal.ParseDate("September 22, 2017 AD")
	check.NoError(t, err)
	check.Equal(t, targetDate, date)
	date, err = cal.ParseDate("September 22, 1 BC")
	check.NoError(t, err)
	check.Equal(t, cal.MustNewDate(9, 22, -1), date)

	targetDate = cal.MustNewDate(9, 22, -2017)
	date, err = cal.ParseDate("9/22/-2017")
	check.NoError(t, err)
	check.Equal(t, targetDate, date)
	date, err = cal.ParseDate("September 22, -2017")
	check.NoError(t, err)
	check.Equal(t, targetDate, date)
}

func TestMarshaling(t *testing.T) {
	cal := calendar.Gregorian()
	calendar.Default = cal
	date := cal.MustNewDate(9, 22, 2017)
	text, err := date.MarshalText()
	check.NoError(t, err)
	check.Equal(t, "9/22/2017", string(text))

	type embedded struct {
		Date calendar.Date
	}
	embeddedDate := embedded{Date: date}
	text, err = json.Marshal(&embeddedDate)
	check.NoError(t, err)
	check.Equal(t, `{"Date":"9/22/2017"}`, string(text))

	text, err = yaml.Marshal(&embeddedDate)
	check.NoError(t, err)
	check.Equal(t, "date: 9/22/2017\n", string(text))

	type embeddedPtr struct {
		Date *calendar.Date
	}
	embeddedPtrDate := embeddedPtr{Date: &date}
	text, err = json.Marshal(&embeddedPtrDate)
	check.NoError(t, err)
	check.Equal(t, `{"Date":"9/22/2017"}`, string(text))

	text, err = yaml.Marshal(&embeddedPtrDate)
	check.NoError(t, err)
	check.Equal(t, "date: 9/22/2017\n", string(text))
}

func TestUnmarshaling(t *testing.T) {
	cal := calendar.Gregorian()
	calendar.Default = cal
	target := cal.MustNewDate(9, 22, 2017)
	var date calendar.Date
	check.NoError(t, date.UnmarshalText([]byte("9/22/2017")))
	check.Equal(t, target, date)

	type embedded struct {
		Date calendar.Date
	}
	var embeddedDate embedded
	check.NoError(t, json.Unmarshal([]byte(`{"Date":"9/22/2017"}`), &embeddedDate))
	check.Equal(t, target, embeddedDate.Date)

	check.NoError(t, yaml.Unmarshal([]byte(`date: 9/22/2017`), &embeddedDate))
	check.Equal(t, target, embeddedDate.Date)

	type embeddedPtr struct {
		Date *calendar.Date
	}
	var embeddedPtrDate embeddedPtr
	check.NoError(t, json.Unmarshal([]byte(`{"Date":"9/22/2017"}`), &embeddedPtrDate))
	check.Equal(t, target, *embeddedPtrDate.Date)

	check.NoError(t, yaml.Unmarshal([]byte(`date: 9/22/2017`), &embeddedPtrDate))
	check.Equal(t, target, *embeddedPtrDate.Date)

	cal = pathfinder.AbsalomReckoning()
	calendar.Default = cal
	date = calendar.Date{}
	target = cal.MustNewDate(9, 22, 2017)
	check.NoError(t, date.UnmarshalText([]byte("9/22/2017 AR")))
	check.Equal(t, target, date)
}
