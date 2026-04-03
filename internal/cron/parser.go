package cron

import (
	"fmt"
	"time"
)

type Schedule struct {
	Minute     int
	Hour       int
	DayOfMonth int
	Month      int
	DayOfWeek  int
}

func parseCronExpression(expr string) (*Schedule, error) {
	var schedule Schedule
	var err error

	_, err = fmt.Sscanf(expr, "%d %d %d %d %d",
		&schedule.Minute,
		&schedule.Hour,
		&schedule.DayOfMonth,
		&schedule.Month,
		&schedule.DayOfWeek,
	)

	if err != nil {
		return nil, err
	}

	return &schedule, nil
}

func (s *Schedule) Next(t time.Time) time.Time {
	next := t.Add(1 * time.Minute)

	for i := 0; i < 60*24*366; i++ {
		if s.matches(next) {
			return next
		}
		next = next.Add(1 * time.Minute)
	}

	return time.Time{}
}

func (s *Schedule) matches(t time.Time) bool {
	matches := true

	if s.Minute != -1 && s.Minute != t.Minute() {
		matches = false
	}
	if s.Hour != -1 && s.Hour != t.Hour() {
		matches = false
	}
	if s.DayOfMonth != -1 && s.DayOfMonth != t.Day() {
		matches = false
	}
	if s.Month != -1 && s.Month != int(t.Month()) {
		matches = false
	}
	if s.DayOfWeek != -1 && s.DayOfWeek != int(t.Weekday()) {
		matches = false
	}

	return matches
}

func ParseStandard(expr string) (*Schedule, error) {
	schedule := &Schedule{
		Minute:     -1,
		Hour:       -1,
		DayOfMonth: -1,
		Month:      -1,
		DayOfWeek:  -1,
	}

	var minute, hour, dayOfMonth, month, dayOfWeek string
	_, err := fmt.Sscanf(expr, "%s %s %s %s %s",
		&minute, &hour, &dayOfMonth, &month, &dayOfWeek,
	)

	if err != nil {
		return nil, err
	}

	schedule.Minute = parseField(minute, 0, 59)
	schedule.Hour = parseField(hour, 0, 23)
	schedule.DayOfMonth = parseField(dayOfMonth, 1, 31)
	schedule.Month = parseField(month, 1, 12)
	schedule.DayOfWeek = parseField(dayOfWeek, 0, 6)

	return schedule, nil
}

func parseField(field string, min, max int) int {
	if field == "*" {
		return -1
	}

	var value int
	_, err := fmt.Sscanf(field, "%d", &value)
	if err != nil {
		return -1
	}

	if value < min || value > max {
		return -1
	}

	return value
}
