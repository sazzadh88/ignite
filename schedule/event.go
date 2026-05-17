package schedule

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Event represents a scheduled task with its frequency, constraints, and hooks.
type Event struct {
	// Core task
	callback func()
	command  string
	job      any

	// Cron scheduling
	expression string

	// Constraints
	filters      []func() bool
	skipFilters  []func() bool
	timeWindows  []timeWindow
	skipWindows  []timeWindow
	environments []string
	dayOfWeek    []time.Weekday

	// Concurrency control
	preventOverlap bool
	runInBg        bool
	onOneServer    bool
	mutex          Mutex
	mutexName      string

	// Hooks
	before    []func()
	after     []func()
	onSuccess []func()
	onFailure []func(error)

	// State
	running sync.Mutex
}

type timeWindow struct {
	start string
	end   string
}

// Cron sets a custom cron expression for the event.
func (e *Event) Cron(expression string) *Event {
	e.expression = expression
	return e
}

// EveryMinute schedules the event to run every minute.
func (e *Event) EveryMinute() *Event {
	return e.Cron("* * * * *")
}

// EveryTwoMinutes schedules the event to run every two minutes.
func (e *Event) EveryTwoMinutes() *Event {
	return e.Cron("*/2 * * * *")
}

// EveryFiveMinutes schedules the event to run every five minutes.
func (e *Event) EveryFiveMinutes() *Event {
	return e.Cron("*/5 * * * *")
}

// EveryTenMinutes schedules the event to run every ten minutes.
func (e *Event) EveryTenMinutes() *Event {
	return e.Cron("*/10 * * * *")
}

// EveryFifteenMinutes schedules the event to run every fifteen minutes.
func (e *Event) EveryFifteenMinutes() *Event {
	return e.Cron("*/15 * * * *")
}

// EveryThirtyMinutes schedules the event to run every thirty minutes.
func (e *Event) EveryThirtyMinutes() *Event {
	return e.Cron("*/30 * * * *")
}

// Hourly schedules the event to run every hour at minute 0.
func (e *Event) Hourly() *Event {
	return e.Cron("0 * * * *")
}

// HourlyAt schedules the event to run every hour at the specified minute.
func (e *Event) HourlyAt(minute int) *Event {
	return e.Cron(fmt.Sprintf("%d * * * *", minute))
}

// EveryTwoHours schedules the event to run every two hours.
func (e *Event) EveryTwoHours() *Event {
	return e.Cron("0 */2 * * *")
}

// EveryThreeHours schedules the event to run every three hours.
func (e *Event) EveryThreeHours() *Event {
	return e.Cron("0 */3 * * *")
}

// EveryFourHours schedules the event to run every four hours.
func (e *Event) EveryFourHours() *Event {
	return e.Cron("0 */4 * * *")
}

// EverySixHours schedules the event to run every six hours.
func (e *Event) EverySixHours() *Event {
	return e.Cron("0 */6 * * *")
}

// Daily schedules the event to run daily at midnight.
func (e *Event) Daily() *Event {
	return e.Cron("0 0 * * *")
}

// DailyAt schedules the event to run daily at the specified time.
// Time format: "HH:MM" (e.g., "13:30", "09:00")
func (e *Event) DailyAt(timeStr string) *Event {
	hour, minute := parseTime(timeStr)
	return e.Cron(fmt.Sprintf("%d %d * * *", minute, hour))
}

// TwiceDaily schedules the event to run twice daily at the specified hours.
func (e *Event) TwiceDaily(hour1, hour2 int) *Event {
	return e.Cron(fmt.Sprintf("0 %d,%d * * *", hour1, hour2))
}

// Weekly schedules the event to run weekly on Sunday at midnight.
func (e *Event) Weekly() *Event {
	return e.Cron("0 0 * * 0")
}

// WeeklyOn schedules the event to run weekly on the specified day and time.
// day: 0 (Sunday) to 6 (Saturday)
// time: "HH:MM" format
func (e *Event) WeeklyOn(day int, timeStr string) *Event {
	hour, minute := parseTime(timeStr)
	return e.Cron(fmt.Sprintf("%d %d * * %d", minute, hour, day))
}

// Monthly schedules the event to run monthly on the 1st at midnight.
func (e *Event) Monthly() *Event {
	return e.Cron("0 0 1 * *")
}

// MonthlyOn schedules the event to run monthly on the specified day and time.
// day: day of month (1-31)
// time: "HH:MM" format
func (e *Event) MonthlyOn(day int, timeStr string) *Event {
	hour, minute := parseTime(timeStr)
	return e.Cron(fmt.Sprintf("%d %d %d * *", minute, hour, day))
}

// Quarterly schedules the event to run quarterly (Jan 1, Apr 1, Jul 1, Oct 1).
func (e *Event) Quarterly() *Event {
	return e.Cron("0 0 1 1,4,7,10 *")
}

// Yearly schedules the event to run yearly on January 1st at midnight.
func (e *Event) Yearly() *Event {
	return e.Cron("0 0 1 1 *")
}

// Weekdays schedules the event to run on weekdays (Monday-Friday).
func (e *Event) Weekdays() *Event {
	return e.Days(time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday)
}

// Weekends schedules the event to run on weekends (Saturday-Sunday).
func (e *Event) Weekends() *Event {
	return e.Days(time.Saturday, time.Sunday)
}

// Sundays schedules the event to run on Sundays.
func (e *Event) Sundays() *Event {
	return e.Days(time.Sunday)
}

// Mondays schedules the event to run on Mondays.
func (e *Event) Mondays() *Event {
	return e.Days(time.Monday)
}

// Tuesdays schedules the event to run on Tuesdays.
func (e *Event) Tuesdays() *Event {
	return e.Days(time.Tuesday)
}

// Wednesdays schedules the event to run on Wednesdays.
func (e *Event) Wednesdays() *Event {
	return e.Days(time.Wednesday)
}

// Thursdays schedules the event to run on Thursdays.
func (e *Event) Thursdays() *Event {
	return e.Days(time.Thursday)
}

// Fridays schedules the event to run on Fridays.
func (e *Event) Fridays() *Event {
	return e.Days(time.Friday)
}

// Saturdays schedules the event to run on Saturdays.
func (e *Event) Saturdays() *Event {
	return e.Days(time.Saturday)
}

// Days constrains the event to run only on the specified days of the week.
func (e *Event) Days(days ...time.Weekday) *Event {
	e.dayOfWeek = days
	return e
}

// Between constrains the event to run only between the specified times.
// Format: "HH:MM" (e.g., "09:00", "17:30")
func (e *Event) Between(start, end string) *Event {
	e.timeWindows = append(e.timeWindows, timeWindow{start: start, end: end})
	return e
}

// UnlessBetween constrains the event to skip running between the specified times.
func (e *Event) UnlessBetween(start, end string) *Event {
	e.skipWindows = append(e.skipWindows, timeWindow{start: start, end: end})
	return e
}

// When adds a callback constraint. The event will only run if the callback returns true.
func (e *Event) When(fn func() bool) *Event {
	e.filters = append(e.filters, fn)
	return e
}

// Skip adds a skip constraint. The event will not run if the callback returns true.
func (e *Event) Skip(fn func() bool) *Event {
	e.skipFilters = append(e.skipFilters, fn)
	return e
}

// WithoutOverlapping prevents the event from running if a previous instance is still running.
func (e *Event) WithoutOverlapping() *Event {
	e.preventOverlap = true
	e.mutexName = fmt.Sprintf("event-%p", e)
	return e
}

// RunInBackground runs the event in a goroutine.
func (e *Event) RunInBackground() *Event {
	e.runInBg = true
	return e
}

// OnOneServer is a placeholder for distributed lock support.
// This would require a distributed locking mechanism (Redis, database, etc).
func (e *Event) OnOneServer() *Event {
	e.onOneServer = true
	return e
}

// Environments constrains the event to run only in the specified environments.
func (e *Event) Environments(envs []string) *Event {
	e.environments = envs
	return e
}

// Before registers a callback to run before the event.
func (e *Event) Before(fn func()) *Event {
	e.before = append(e.before, fn)
	return e
}

// After registers a callback to run after the event.
func (e *Event) After(fn func()) *Event {
	e.after = append(e.after, fn)
	return e
}

// OnSuccess registers a callback to run when the event completes successfully.
func (e *Event) OnSuccess(fn func()) *Event {
	e.onSuccess = append(e.onSuccess, fn)
	return e
}

// OnFailure registers a callback to run when the event fails.
func (e *Event) OnFailure(fn func(error)) *Event {
	e.onFailure = append(e.onFailure, fn)
	return e
}

// IsDue checks if the event should run at the given time.
func (e *Event) IsDue(now time.Time) bool {
	// Parse cron expression
	expr, err := Parse(e.expression)
	if err != nil || !expr.Matches(now) {
		return false
	}

	// Check day of week constraints
	if len(e.dayOfWeek) > 0 {
		found := false
		for _, day := range e.dayOfWeek {
			if now.Weekday() == day {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check time windows
	if len(e.timeWindows) > 0 {
		inWindow := false
		for _, window := range e.timeWindows {
			if e.isInTimeWindow(now, window) {
				inWindow = true
				break
			}
		}
		if !inWindow {
			return false
		}
	}

	// Check skip windows
	for _, window := range e.skipWindows {
		if e.isInTimeWindow(now, window) {
			return false
		}
	}

	// Check when filters
	for _, filter := range e.filters {
		if !filter() {
			return false
		}
	}

	// Check skip filters
	for _, filter := range e.skipFilters {
		if filter() {
			return false
		}
	}

	return true
}

// isInTimeWindow checks if the given time falls within a time window.
// Handles windows that cross midnight (e.g., "23:00" to "06:00").
func (e *Event) isInTimeWindow(now time.Time, window timeWindow) bool {
	currentTime := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())

	// Normal window (start < end, e.g., "09:00" to "17:00")
	if window.start <= window.end {
		return currentTime >= window.start && currentTime <= window.end
	}

	// Window crosses midnight (start > end, e.g., "23:00" to "06:00")
	// Time is in window if it's >= start OR <= end
	return currentTime >= window.start || currentTime <= window.end
}

// run executes the event.
func (e *Event) run() error {
	// Check for overlap
	if e.preventOverlap {
		if e.mutex != nil && e.mutex.Exists(e.mutexName) {
			return nil // Skip if already running
		}

		if e.mutex != nil && !e.mutex.Create(e.mutexName) {
			return nil // Failed to acquire lock
		}

		defer func() {
			if e.mutex != nil {
				e.mutex.Forget(e.mutexName)
			}
		}()
	}

	// Run before hooks
	for _, hook := range e.before {
		hook()
	}

	// Execute the task
	var err error
	if e.callback != nil {
		err = e.runCallback()
	} else if e.command != "" {
		err = e.runCommand()
	} else if e.job != nil {
		err = e.runJob()
	}

	// Run after hooks
	for _, hook := range e.after {
		hook()
	}

	// Run success/failure hooks
	if err != nil {
		for _, hook := range e.onFailure {
			hook(err)
		}
	} else {
		for _, hook := range e.onSuccess {
			hook()
		}
	}

	return err
}

// runCallback executes the callback function.
func (e *Event) runCallback() error {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()

	e.callback()
	return nil
}

// runCommand executes the shell command.
func (e *Event) runCommand() error {
	cmd := exec.Command("sh", "-c", e.command)
	return cmd.Run()
}

// runJob dispatches the queue job.
// This is a placeholder - actual implementation would depend on queue system.
func (e *Event) runJob() error {
	// TODO: Integrate with queue system
	return nil
}

// parseTime parses a time string in "HH:MM" format and returns hour and minute as integers.
func parseTime(timeStr string) (hour, minute int) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, 0
	}
	fmt.Sscanf(parts[0], "%d", &hour)
	fmt.Sscanf(parts[1], "%d", &minute)
	return hour, minute
}
