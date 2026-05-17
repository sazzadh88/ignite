// Package schedule provides a task scheduling system for Ignite.
// It supports cron-based scheduling with Laravel-inspired fluent API.
//
// Basic usage:
//
//	s := schedule.NewSchedule()
//
//	// Schedule a callback
//	s.Call(func() {
//	    fmt.Println("Hello")
//	}).EveryMinute()
//
//	// Schedule a command
//	s.Command("backup:run").Daily()
//
//	// Add constraints
//	s.Call(func() {
//	    // Task
//	}).EveryFiveMinutes().
//	    Weekdays().
//	    Between("09:00", "17:00")
//
//	// Run due tasks (typically called every minute by scheduler daemon)
//	s.Run()
//
// The package includes:
//   - Cron expression parser supporting standard 5-field format
//   - Fluent API for common frequencies (EveryMinute, Hourly, Daily, etc.)
//   - Time window constraints (Between, UnlessBetween)
//   - Day-of-week filters (Weekdays, Weekends, Mondays, etc.)
//   - Conditional execution (When, Skip)
//   - Overlap prevention (WithoutOverlapping)
//   - Lifecycle hooks (Before, After, OnSuccess, OnFailure)
//   - Background execution support
//
// Zero external dependencies - uses only Go standard library.
package schedule

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CronExpression represents a parsed cron expression with five fields:
// minute hour day month weekday
type CronExpression struct {
	minute  fieldSet
	hour    fieldSet
	day     fieldSet
	month   fieldSet
	weekday fieldSet
}

// fieldSet represents allowed values for a cron field.
type fieldSet struct {
	values map[int]bool
	any    bool
}

// Parse parses a standard 5-field cron expression.
// Format: minute hour day month weekday
// Supports: *, specific values, ranges (1-5), steps (*/5), lists (1,3,5)
func Parse(expr string) (*CronExpression, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return nil, fmt.Errorf("invalid cron expression: expected 5 fields, got %d", len(fields))
	}

	ce := &CronExpression{}
	var err error

	ce.minute, err = parseField(fields[0], 0, 59)
	if err != nil {
		return nil, fmt.Errorf("invalid minute field: %w", err)
	}

	ce.hour, err = parseField(fields[1], 0, 23)
	if err != nil {
		return nil, fmt.Errorf("invalid hour field: %w", err)
	}

	ce.day, err = parseField(fields[2], 1, 31)
	if err != nil {
		return nil, fmt.Errorf("invalid day field: %w", err)
	}

	ce.month, err = parseField(fields[3], 1, 12)
	if err != nil {
		return nil, fmt.Errorf("invalid month field: %w", err)
	}

	ce.weekday, err = parseField(fields[4], 0, 6)
	if err != nil {
		return nil, fmt.Errorf("invalid weekday field: %w", err)
	}

	return ce, nil
}

// Matches checks if the given time matches this cron expression.
func (ce *CronExpression) Matches(t time.Time) bool {
	return ce.minute.matches(t.Minute()) &&
		ce.hour.matches(t.Hour()) &&
		ce.day.matches(t.Day()) &&
		ce.month.matches(int(t.Month())) &&
		ce.weekday.matches(int(t.Weekday()))
}

// parseField parses a single cron field with support for *, ranges, steps, and lists.
func parseField(field string, min, max int) (fieldSet, error) {
	fs := fieldSet{values: make(map[int]bool)}

	// Handle wildcard
	if field == "*" {
		fs.any = true
		return fs, nil
	}

	// Handle step values like */5 or 10-20/2
	if strings.Contains(field, "/") {
		return parseStep(field, min, max)
	}

	// Handle comma-separated lists
	if strings.Contains(field, ",") {
		parts := strings.Split(field, ",")
		for _, part := range parts {
			if strings.Contains(part, "-") {
				start, end, err := parseRange(part)
				if err != nil {
					return fs, err
				}
				for i := start; i <= end; i++ {
					if i < min || i > max {
						return fs, fmt.Errorf("value %d out of range [%d-%d]", i, min, max)
					}
					fs.values[i] = true
				}
			} else {
				val, err := strconv.Atoi(strings.TrimSpace(part))
				if err != nil {
					return fs, fmt.Errorf("invalid value: %s", part)
				}
				if val < min || val > max {
					return fs, fmt.Errorf("value %d out of range [%d-%d]", val, min, max)
				}
				fs.values[val] = true
			}
		}
		return fs, nil
	}

	// Handle ranges like 1-5
	if strings.Contains(field, "-") {
		start, end, err := parseRange(field)
		if err != nil {
			return fs, err
		}
		if start < min || end > max {
			return fs, fmt.Errorf("range %d-%d out of bounds [%d-%d]", start, end, min, max)
		}
		for i := start; i <= end; i++ {
			fs.values[i] = true
		}
		return fs, nil
	}

	// Handle single value
	val, err := strconv.Atoi(strings.TrimSpace(field))
	if err != nil {
		return fs, fmt.Errorf("invalid value: %s", field)
	}
	if val < min || val > max {
		return fs, fmt.Errorf("value %d out of range [%d-%d]", val, min, max)
	}
	fs.values[val] = true

	return fs, nil
}

// parseStep parses step expressions like */5 or 10-20/2.
func parseStep(field string, min, max int) (fieldSet, error) {
	fs := fieldSet{values: make(map[int]bool)}
	parts := strings.Split(field, "/")
	if len(parts) != 2 {
		return fs, fmt.Errorf("invalid step expression: %s", field)
	}

	step, err := strconv.Atoi(parts[1])
	if err != nil || step <= 0 {
		return fs, fmt.Errorf("invalid step value: %s", parts[1])
	}

	start, end := min, max

	// Handle range before step (e.g., 10-20/2)
	if parts[0] != "*" {
		if strings.Contains(parts[0], "-") {
			start, end, err = parseRange(parts[0])
			if err != nil {
				return fs, err
			}
		} else {
			start, err = strconv.Atoi(parts[0])
			if err != nil {
				return fs, fmt.Errorf("invalid start value: %s", parts[0])
			}
		}
	}

	if start < min || end > max {
		return fs, fmt.Errorf("range %d-%d out of bounds [%d-%d]", start, end, min, max)
	}

	for i := start; i <= end; i += step {
		fs.values[i] = true
	}

	return fs, nil
}

// parseRange parses a range expression like 1-5.
func parseRange(expr string) (int, int, error) {
	parts := strings.Split(expr, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range expression: %s", expr)
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid range start: %s", parts[0])
	}

	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid range end: %s", parts[1])
	}

	if start > end {
		return 0, 0, fmt.Errorf("invalid range: start %d > end %d", start, end)
	}

	return start, end, nil
}

// matches checks if a value matches this field set.
func (fs *fieldSet) matches(value int) bool {
	if fs.any {
		return true
	}
	return fs.values[value]
}
