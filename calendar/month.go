package calendar

import "github.com/richardwilkes/toolbox/errs"

// Month holds information about a month within the calendar.
type Month struct {
	Name string `json:"name" yaml:"name"`
	Days int    `json:"days" yaml:"days"`
}

// Valid returns nil if the month data is usable.
func (month *Month) Valid() error {
	if month.Name == "" {
		return errs.New("Calendar month names must not be empty")
	}
	if month.Days < 1 {
		return errs.New("Calendar months must have at least 1 day")
	}
	return nil
}
