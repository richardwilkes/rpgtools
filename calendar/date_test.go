// Copyright (c) 2017-2025 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/toolbox/v2/check"
	"gopkg.in/yaml.v3"
)

func TestNewDate(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	calendar.Default = cal
	d, err := cal.NewDate(1, 1, 1)
	c.NoError(err)
	c.Equal(cal.NewDateByDays(0), d)
	d, err = cal.NewDate(12, 31, 1)
	c.NoError(err)
	c.Equal(cal.NewDateByDays(364), d)
	d, err = cal.NewDate(1, 1, 2)
	c.NoError(err)
	c.Equal(cal.NewDateByDays(365), d)

	d, err = cal.NewDate(1, 1, -1)
	c.NoError(err)
	c.Equal(cal.NewDateByDays(-366), d)
	d, err = cal.NewDate(12, 31, -1)
	c.NoError(err)
	c.Equal(cal.NewDateByDays(-1), d)
	d, err = cal.NewDate(1, 1, -2)
	c.NoError(err)
	c.Equal(cal.NewDateByDays(-731), d)
	d, err = cal.NewDate(12, 31, -2)
	c.NoError(err)
	c.Equal(cal.NewDateByDays(-367), d)
	d, err = cal.NewDate(12, 31, -3)
	c.NoError(err)
	c.Equal(cal.NewDateByDays(-732), d)

	_, err = cal.NewDate(1, 1, 0)
	c.HasError(err)
	_, err = cal.NewDate(13, 22, 2017)
	c.HasError(err)
	_, err = cal.NewDate(9, 888, 2017)
	c.HasError(err)
}

func TestYear(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	calendar.Default = cal
	c.Equal(1, cal.NewDateByDays(0).Year(), "First day of year 1")
	c.Equal(1, cal.NewDateByDays(364).Year(), "Last day of year 1")
	c.Equal(2, cal.NewDateByDays(365).Year(), "First day of year 2")

	c.Equal(-1, cal.NewDateByDays(-366).Year(), "First day of year -1")
	c.Equal(-1, cal.NewDateByDays(-1).Year(), "Last day of year -1")
	c.Equal(-2, cal.NewDateByDays(-731).Year(), "First day of year -2")
	c.Equal(-2, cal.NewDateByDays(-367).Year(), "Last day of year -2")
	c.Equal(-3, cal.NewDateByDays(-732).Year(), "Last day of year -3")

	for year := 1; year < 5000; year++ {
		c.Equal(year, cal.MustNewDate(1, 1, year).Year())
		c.Equal(year, cal.MustNewDate(12, 31, year).Year())
	}

	for year := -1; year > -5000; year-- {
		c.Equal(year, cal.MustNewDate(1, 1, year).Year())
		c.Equal(year, cal.MustNewDate(12, 31, year).Year())
	}
}

func TestDayInYear(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	calendar.Default = cal
	c.Equal(1, cal.MustNewDate(1, 1, 1).DayInYear())
	c.Equal(365, cal.MustNewDate(12, 31, 1).DayInYear())
	c.Equal(1, cal.MustNewDate(1, 1, 2).DayInYear())
	c.Equal(1, cal.MustNewDate(1, 1, 4).DayInYear())
	c.Equal(366, cal.MustNewDate(12, 31, 4).DayInYear())

	c.Equal(1, cal.MustNewDate(1, 1, -1).DayInYear())
	c.Equal(366, cal.MustNewDate(12, 31, -1).DayInYear())
	c.Equal(1, cal.MustNewDate(1, 1, -2).DayInYear())
	c.Equal(365, cal.MustNewDate(12, 31, -2).DayInYear())
	c.Equal(1, cal.MustNewDate(1, 1, -5).DayInYear())
	c.Equal(366, cal.MustNewDate(12, 31, -5).DayInYear())
}

func TestMonth(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	calendar.Default = cal
	c.Equal(1, cal.MustNewDate(1, 1, 1).Month())
	c.Equal(1, cal.MustNewDate(1, 31, 1).Month())
	c.Equal(2, cal.MustNewDate(2, 1, 1).Month())
	c.Equal(2, cal.MustNewDate(2, 28, 1).Month())
	c.Equal(3, cal.MustNewDate(3, 1, 1).Month())
	c.Equal(12, cal.MustNewDate(12, 31, 1).Month())
	c.Equal(1, cal.MustNewDate(1, 1, 2).Month())
	c.Equal(2, cal.MustNewDate(2, 28, 4).Month())
	c.Equal(2, cal.MustNewDate(2, 29, 4).Month())
	c.Equal(3, cal.MustNewDate(3, 1, 4).Month())

	c.Equal(2, cal.MustNewDate(2, 29, -1).Month())
	c.Equal(2, cal.MustNewDate(2, 28, -2).Month())
}

func TestDayInMonth(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	calendar.Default = cal
	c.Equal(1, cal.MustNewDate(1, 1, 1).DayInMonth())
	c.Equal(31, cal.MustNewDate(1, 31, 1).DayInMonth())
	c.Equal(1, cal.MustNewDate(2, 1, 1).DayInMonth())
	c.Equal(28, cal.MustNewDate(2, 28, 1).DayInMonth())
	c.Equal(1, cal.MustNewDate(3, 1, 1).DayInMonth())
	c.Equal(31, cal.MustNewDate(12, 31, 1).DayInMonth())
	c.Equal(1, cal.MustNewDate(1, 1, 2).DayInMonth())
	c.Equal(28, cal.MustNewDate(2, 28, 2).DayInMonth())
	c.Equal(1, cal.MustNewDate(3, 1, 2).DayInMonth())
	c.Equal(28, cal.MustNewDate(2, 28, 4).DayInMonth())
	c.Equal(29, cal.MustNewDate(2, 29, 4).DayInMonth())
	c.Equal(1, cal.MustNewDate(3, 1, 4).DayInMonth())

	c.Equal(29, cal.MustNewDate(2, 29, -1).DayInMonth())
	c.Equal(28, cal.MustNewDate(2, 28, -2).DayInMonth())
}

func TestDateToString(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	calendar.Default = cal
	c.Equal("1/1/1", cal.MustNewDate(1, 1, 1).String())
	c.Equal("12/31/1", cal.MustNewDate(12, 31, 1).String())
	c.Equal("1/1/2", cal.MustNewDate(1, 1, 2).String())
	c.Equal("1/1/2017", cal.MustNewDate(1, 1, 2017).String())
	c.Equal("9/22/2017", cal.MustNewDate(9, 22, 2017).String())

	c.Equal("1/1/1 BC", cal.MustNewDate(1, 1, -1).String())
	c.Equal("12/31/1 BC", cal.MustNewDate(12, 31, -1).String())
	c.Equal("1/1/2 BC", cal.MustNewDate(1, 1, -2).String())
	c.Equal("12/31/2 BC", cal.MustNewDate(12, 31, -2).String())
	c.Equal("12/31/3 BC", cal.MustNewDate(12, 31, -3).String())
}

func TestWeekDay(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	calendar.Default = cal
	c.Equal(1, cal.MustNewDate(1, 1, 1).WeekDay())
	c.Equal(4, cal.MustNewDate(1, 4, 1).WeekDay())
	c.Equal(1, cal.MustNewDate(1, 8, 1).WeekDay())
	c.Equal(0, cal.MustNewDate(12, 31, -1).WeekDay())
	c.Equal(6, cal.MustNewDate(12, 30, -1).WeekDay())
	c.Equal(0, cal.MustNewDate(12, 24, -1).WeekDay())
	c.Equal(6, cal.MustNewDate(1, 1, 2000).WeekDay())
	c.Equal(1, cal.MustNewDate(9, 3, 2018).WeekDay())
}

func TestFormat(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	calendar.Default = cal
	d := cal.MustNewDate(9, 22, 2017)
	c.Equal("9/22/2017", d.Format(calendar.ShortFormat))
	c.Equal("Sep 22, 2017", d.Format(calendar.MediumFormat))
	c.Equal("September 22, 2017", d.Format(calendar.LongFormat))
	c.Equal("Friday, September 22, 2017", d.Format(calendar.FullFormat))
	c.Equal("%Fri%", d.Format("%%%w%%"))
	c.Equal("Friday, September 22, 2017 AD", d.Format("%W, %M %D, %y"))

	d = cal.MustNewDate(9, 22, -1)
	c.Equal("9/22/1 BC", d.Format(calendar.ShortFormat))
	c.Equal("Sep 22, 1 BC", d.Format(calendar.MediumFormat))
	c.Equal("September 22, 1 BC", d.Format(calendar.LongFormat))
	c.Equal("Friday, September 22, 1 BC", d.Format(calendar.FullFormat))
	c.Equal("%Fri%", d.Format("%%%w%%"))
	c.Equal("Friday, September 22, 1 BC", d.Format("%W, %M %D, %y"))
}

func TestParseDate(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	calendar.Default = cal
	targetDate := cal.MustNewDate(9, 22, 2017)
	date, err := cal.ParseDate("A long, rambling prefix September 22, 2017 and a long suffix")
	c.NoError(err)
	c.Equal(targetDate, date)
	date, err = cal.ParseDate("Friday, September 22, 2017")
	c.NoError(err)
	c.Equal(targetDate, date)
	date, err = cal.ParseDate("September 22, 2017")
	c.NoError(err)
	c.Equal(targetDate, date)
	date, err = cal.ParseDate("9/22/2017")
	c.NoError(err)
	c.Equal(targetDate, date)
	date, err = cal.ParseDate("what 9/22/2017 how?")
	c.NoError(err)
	c.Equal(targetDate, date)
	_, err = cal.ParseDate("9/22")
	c.HasError(err)
	_, err = cal.ParseDate("9/666/2017")
	c.HasError(err)
	_, err = cal.ParseDate("13/22/2017")
	c.HasError(err)
	date, err = cal.ParseDate("September 22, 2017 AD")
	c.NoError(err)
	c.Equal(targetDate, date)
	date, err = cal.ParseDate("September 22, 1 BC")
	c.NoError(err)
	c.Equal(cal.MustNewDate(9, 22, -1), date)

	targetDate = cal.MustNewDate(9, 22, -2017)
	date, err = cal.ParseDate("9/22/-2017")
	c.NoError(err)
	c.Equal(targetDate, date)
	date, err = cal.ParseDate("September 22, -2017")
	c.NoError(err)
	c.Equal(targetDate, date)
}

func TestMarshaling(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	calendar.Default = cal
	date := cal.MustNewDate(9, 22, 2017)
	text, err := date.MarshalText()
	c.NoError(err)
	c.Equal("9/22/2017", string(text))

	type embedded struct {
		Date calendar.Date
	}
	embeddedDate := embedded{Date: date}
	text, err = json.Marshal(&embeddedDate)
	c.NoError(err)
	c.Equal(`{"Date":"9/22/2017"}`, string(text))

	text, err = yaml.Marshal(&embeddedDate)
	c.NoError(err)
	c.Equal("date: 9/22/2017\n", string(text))

	type embeddedPtr struct {
		Date *calendar.Date
	}
	embeddedPtrDate := embeddedPtr{Date: &date}
	text, err = json.Marshal(&embeddedPtrDate)
	c.NoError(err)
	c.Equal(`{"Date":"9/22/2017"}`, string(text))

	text, err = yaml.Marshal(&embeddedPtrDate)
	c.NoError(err)
	c.Equal("date: 9/22/2017\n", string(text))
}

func TestUnmarshaling(t *testing.T) {
	c := check.New(t)
	cal := calendar.Gregorian()
	calendar.Default = cal
	target := cal.MustNewDate(9, 22, 2017)
	var date calendar.Date
	c.NoError(date.UnmarshalText([]byte("9/22/2017")))
	c.Equal(target, date)

	type embedded struct {
		Date calendar.Date
	}
	var embeddedDate embedded
	c.NoError(json.Unmarshal([]byte(`{"Date":"9/22/2017"}`), &embeddedDate))
	c.Equal(target, embeddedDate.Date)

	c.NoError(yaml.Unmarshal([]byte(`date: 9/22/2017`), &embeddedDate))
	c.Equal(target, embeddedDate.Date)

	type embeddedPtr struct {
		Date *calendar.Date
	}
	var embeddedPtrDate embeddedPtr
	c.NoError(json.Unmarshal([]byte(`{"Date":"9/22/2017"}`), &embeddedPtrDate))
	c.Equal(target, *embeddedPtrDate.Date)

	c.NoError(yaml.Unmarshal([]byte(`date: 9/22/2017`), &embeddedPtrDate))
	c.Equal(target, *embeddedPtrDate.Date)

	cal = pathfinder.AbsalomReckoning()
	calendar.Default = cal
	date = calendar.Date{}
	target = cal.MustNewDate(9, 22, 2017)
	c.NoError(date.UnmarshalText([]byte("9/22/2017 AR")))
	c.Equal(target, date)
}
