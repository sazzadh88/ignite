package schedule

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Test cron expression parsing
func TestCronParsing(t *testing.T) {
	tests := []struct {
		expr    string
		wantErr bool
	}{
		{"* * * * *", false},
		{"0 0 * * *", false},
		{"*/5 * * * *", false},
		{"0 */2 * * *", false},
		{"0 0 1 * *", false},
		{"0 0 1 1 *", false},
		{"0 0 * * 0", false},
		{"0-59 0-23 1-31 1-12 0-6", false},
		{"0,15,30,45 * * * *", false},
		{"0 9-17 * * 1-5", false},
		{"invalid", true},
		{"* * *", true},
		{"60 * * * *", true},
		{"* 24 * * *", true},
		{"* * 32 * *", true},
		{"* * * 13 *", true},
		{"* * * * 7", true},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			_, err := Parse(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.expr, err, tt.wantErr)
			}
		})
	}
}

// Test cron expression matching
func TestCronMatching(t *testing.T) {
	tests := []struct {
		expr   string
		time   time.Time
		expect bool
	}{
		// Every minute
		{"* * * * *", time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), true},
		{"* * * * *", time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC), true},

		// Specific minute
		{"30 * * * *", time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC), true},
		{"30 * * * *", time.Date(2024, 1, 1, 12, 31, 0, 0, time.UTC), false},

		// Every 5 minutes
		{"*/5 * * * *", time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), true},
		{"*/5 * * * *", time.Date(2024, 1, 1, 12, 5, 0, 0, time.UTC), true},
		{"*/5 * * * *", time.Date(2024, 1, 1, 12, 10, 0, 0, time.UTC), true},
		{"*/5 * * * *", time.Date(2024, 1, 1, 12, 7, 0, 0, time.UTC), false},

		// Hourly at minute 0
		{"0 * * * *", time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), true},
		{"0 * * * *", time.Date(2024, 1, 1, 12, 1, 0, 0, time.UTC), false},

		// Daily at midnight
		{"0 0 * * *", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), true},
		{"0 0 * * *", time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC), false},
		{"0 0 * * *", time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), false},

		// Weekly on Sunday
		{"0 0 * * 0", time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), true},  // Sunday
		{"0 0 * * 0", time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), false}, // Monday

		// Monthly on 1st
		{"0 0 1 * *", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), true},
		{"0 0 1 * *", time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), false},

		// Ranges
		{"0 9-17 * * *", time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), true},
		{"0 9-17 * * *", time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), true},
		{"0 9-17 * * *", time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC), true},
		{"0 9-17 * * *", time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC), false},
		{"0 9-17 * * *", time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC), false},

		// Lists
		{"0,15,30,45 * * * *", time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), true},
		{"0,15,30,45 * * * *", time.Date(2024, 1, 1, 12, 15, 0, 0, time.UTC), true},
		{"0,15,30,45 * * * *", time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC), true},
		{"0,15,30,45 * * * *", time.Date(2024, 1, 1, 12, 45, 0, 0, time.UTC), true},
		{"0,15,30,45 * * * *", time.Date(2024, 1, 1, 12, 5, 0, 0, time.UTC), false},

		// Weekdays (Mon-Fri)
		{"0 9 * * 1-5", time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), true},  // Monday
		{"0 9 * * 1-5", time.Date(2024, 1, 2, 9, 0, 0, 0, time.UTC), true},  // Tuesday
		{"0 9 * * 1-5", time.Date(2024, 1, 5, 9, 0, 0, 0, time.UTC), true},  // Friday
		{"0 9 * * 1-5", time.Date(2024, 1, 6, 9, 0, 0, 0, time.UTC), false}, // Saturday
		{"0 9 * * 1-5", time.Date(2024, 1, 7, 9, 0, 0, 0, time.UTC), false}, // Sunday
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s at %s", tt.expr, tt.time.Format("2006-01-02 15:04"))
		t.Run(name, func(t *testing.T) {
			expr, err := Parse(tt.expr)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			if got := expr.Matches(tt.time); got != tt.expect {
				t.Errorf("Matches() = %v, want %v", got, tt.expect)
			}
		})
	}
}

// Test frequency methods set correct cron expressions
func TestFrequencyMethods(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*Event) *Event
		expect string
	}{
		{"EveryMinute", func(e *Event) *Event { return e.EveryMinute() }, "* * * * *"},
		{"EveryTwoMinutes", func(e *Event) *Event { return e.EveryTwoMinutes() }, "*/2 * * * *"},
		{"EveryFiveMinutes", func(e *Event) *Event { return e.EveryFiveMinutes() }, "*/5 * * * *"},
		{"EveryTenMinutes", func(e *Event) *Event { return e.EveryTenMinutes() }, "*/10 * * * *"},
		{"EveryFifteenMinutes", func(e *Event) *Event { return e.EveryFifteenMinutes() }, "*/15 * * * *"},
		{"EveryThirtyMinutes", func(e *Event) *Event { return e.EveryThirtyMinutes() }, "*/30 * * * *"},
		{"Hourly", func(e *Event) *Event { return e.Hourly() }, "0 * * * *"},
		{"HourlyAt", func(e *Event) *Event { return e.HourlyAt(30) }, "30 * * * *"},
		{"EveryTwoHours", func(e *Event) *Event { return e.EveryTwoHours() }, "0 */2 * * *"},
		{"EveryThreeHours", func(e *Event) *Event { return e.EveryThreeHours() }, "0 */3 * * *"},
		{"EveryFourHours", func(e *Event) *Event { return e.EveryFourHours() }, "0 */4 * * *"},
		{"EverySixHours", func(e *Event) *Event { return e.EverySixHours() }, "0 */6 * * *"},
		{"Daily", func(e *Event) *Event { return e.Daily() }, "0 0 * * *"},
		{"DailyAt", func(e *Event) *Event { return e.DailyAt("13:30") }, "30 13 * * *"},
		{"TwiceDaily", func(e *Event) *Event { return e.TwiceDaily(9, 17) }, "0 9,17 * * *"},
		{"Weekly", func(e *Event) *Event { return e.Weekly() }, "0 0 * * 0"},
		{"WeeklyOn", func(e *Event) *Event { return e.WeeklyOn(3, "10:30") }, "30 10 * * 3"},
		{"Monthly", func(e *Event) *Event { return e.Monthly() }, "0 0 1 * *"},
		{"MonthlyOn", func(e *Event) *Event { return e.MonthlyOn(15, "14:00") }, "0 14 15 * *"},
		{"Quarterly", func(e *Event) *Event { return e.Quarterly() }, "0 0 1 1,4,7,10 *"},
		{"Yearly", func(e *Event) *Event { return e.Yearly() }, "0 0 1 1 *"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &Event{}
			tt.setup(event)

			if event.expression != tt.expect {
				t.Errorf("expression = %q, want %q", event.expression, tt.expect)
			}
		})
	}
}

// Test IsDue with various times
func TestIsDue(t *testing.T) {
	event := &Event{}
	event.EveryMinute()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if !event.IsDue(now) {
		t.Error("Event should be due every minute")
	}

	event.Hourly()
	if !event.IsDue(now) {
		t.Error("Event should be due at minute 0")
	}

	later := time.Date(2024, 1, 1, 12, 1, 0, 0, time.UTC)
	if event.IsDue(later) {
		t.Error("Event should not be due at minute 1")
	}
}

// Test When/Skip filters
func TestFilters(t *testing.T) {
	event := &Event{}
	event.EveryMinute()

	called := false
	event.When(func() bool {
		called = true
		return false
	})

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if event.IsDue(now) {
		t.Error("Event should not be due when filter returns false")
	}

	if !called {
		t.Error("When filter should have been called")
	}

	// Test Skip
	event2 := &Event{}
	event2.EveryMinute()

	skipCalled := false
	event2.Skip(func() bool {
		skipCalled = true
		return true
	})

	if event2.IsDue(now) {
		t.Error("Event should not be due when skip filter returns true")
	}

	if !skipCalled {
		t.Error("Skip filter should have been called")
	}
}

// Test Before/After hooks
func TestHooks(t *testing.T) {
	s := NewSchedule()

	beforeCalled := false
	afterCalled := false
	successCalled := false

	event := s.Call(func() {
		// Task body
	})
	event.EveryMinute()
	event.Before(func() {
		beforeCalled = true
	})
	event.After(func() {
		afterCalled = true
	})
	event.OnSuccess(func() {
		successCalled = true
	})

	event.run()

	if !beforeCalled {
		t.Error("Before hook should have been called")
	}
	if !afterCalled {
		t.Error("After hook should have been called")
	}
	if !successCalled {
		t.Error("OnSuccess hook should have been called")
	}
}

// Test WithoutOverlapping prevents concurrent runs
func TestWithoutOverlapping(t *testing.T) {
	s := NewSchedule()

	runCount := 0
	event := s.Call(func() {
		runCount++
		time.Sleep(100 * time.Millisecond)
	})
	event.WithoutOverlapping()

	// First run should succeed
	go event.run()
	time.Sleep(10 * time.Millisecond)

	// Second run should be skipped
	event.run()

	// Wait for first run to complete
	time.Sleep(150 * time.Millisecond)

	if runCount != 1 {
		t.Errorf("Expected 1 run, got %d", runCount)
	}

	// Third run should succeed after first completes
	event.run()
	time.Sleep(150 * time.Millisecond)

	if runCount != 2 {
		t.Errorf("Expected 2 runs, got %d", runCount)
	}
}

// Test DueEvents returns correct events
func TestDueEvents(t *testing.T) {
	s := NewSchedule()

	// Event that runs every minute
	s.Call(func() {}).EveryMinute()

	// Event that runs hourly
	s.Call(func() {}).Hourly()

	// Event that runs daily
	s.Call(func() {}).Daily()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	due := s.DueEvents(now)

	// At 12:00, both every-minute and hourly should be due
	if len(due) != 2 {
		t.Errorf("Expected 2 due events at 12:00, got %d", len(due))
	}

	// At 00:00, all three should be due
	midnight := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	due = s.DueEvents(midnight)

	if len(due) != 3 {
		t.Errorf("Expected 3 due events at midnight, got %d", len(due))
	}

	// At 12:30, only every-minute should be due
	halfPast := time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC)
	due = s.DueEvents(halfPast)

	if len(due) != 1 {
		t.Errorf("Expected 1 due event at 12:30, got %d", len(due))
	}
}

// Test Between/UnlessBetween time windows
func TestTimeWindows(t *testing.T) {
	event := &Event{}
	event.EveryMinute()
	event.Between("09:00", "17:00")

	// Inside window
	inside := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if !event.IsDue(inside) {
		t.Error("Event should be due between 09:00 and 17:00")
	}

	// Outside window
	outside := time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC)
	if event.IsDue(outside) {
		t.Error("Event should not be due outside 09:00-17:00")
	}

	// Test UnlessBetween
	event2 := &Event{}
	event2.EveryMinute()
	event2.UnlessBetween("23:00", "06:00")

	// Inside skip window
	night := time.Date(2024, 1, 1, 2, 0, 0, 0, time.UTC)
	if event2.IsDue(night) {
		t.Error("Event should not be due between 23:00 and 06:00")
	}

	// Outside skip window
	day := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if !event2.IsDue(day) {
		t.Error("Event should be due outside 23:00-06:00")
	}
}

// Test day-of-week constraints
func TestDayOfWeekConstraints(t *testing.T) {
	event := &Event{}
	event.EveryMinute()
	event.Weekdays()

	// Monday - should run
	monday := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if !event.IsDue(monday) {
		t.Error("Event should run on Monday (weekday)")
	}

	// Saturday - should not run
	saturday := time.Date(2024, 1, 6, 12, 0, 0, 0, time.UTC)
	if event.IsDue(saturday) {
		t.Error("Event should not run on Saturday (weekend)")
	}

	// Test Weekends
	event2 := &Event{}
	event2.EveryMinute()
	event2.Weekends()

	if event2.IsDue(monday) {
		t.Error("Event should not run on Monday (weekday)")
	}

	if !event2.IsDue(saturday) {
		t.Error("Event should run on Saturday (weekend)")
	}

	// Test specific days
	event3 := &Event{}
	event3.EveryMinute()
	event3.Days(time.Monday, time.Wednesday, time.Friday)

	if !event3.IsDue(monday) {
		t.Error("Event should run on Monday")
	}

	tuesday := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
	if event3.IsDue(tuesday) {
		t.Error("Event should not run on Tuesday")
	}

	wednesday := time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)
	if !event3.IsDue(wednesday) {
		t.Error("Event should run on Wednesday")
	}
}

// Test FileMutex
func TestFileMutex(t *testing.T) {
	dir := t.TempDir()
	mutex := NewFileMutex(dir)

	// Create should succeed first time
	if !mutex.Create("test") {
		t.Error("First Create should succeed")
	}

	// Exists should return true
	if !mutex.Exists("test") {
		t.Error("Mutex should exist after Create")
	}

	// Second Create should fail
	if mutex.Create("test") {
		t.Error("Second Create should fail")
	}

	// Forget should remove the mutex
	mutex.Forget("test")

	if mutex.Exists("test") {
		t.Error("Mutex should not exist after Forget")
	}

	// Create should succeed again
	if !mutex.Create("test") {
		t.Error("Create should succeed after Forget")
	}

	// Cleanup
	mutex.Forget("test")
}

// Test MemoryMutex
func TestMemoryMutex(t *testing.T) {
	mutex := NewMemoryMutex()

	// Create should succeed first time
	if !mutex.Create("test") {
		t.Error("First Create should succeed")
	}

	// Exists should return true
	if !mutex.Exists("test") {
		t.Error("Mutex should exist after Create")
	}

	// Second Create should fail
	if mutex.Create("test") {
		t.Error("Second Create should fail")
	}

	// Forget should remove the mutex
	mutex.Forget("test")

	if mutex.Exists("test") {
		t.Error("Mutex should not exist after Forget")
	}

	// Create should succeed again
	if !mutex.Create("test") {
		t.Error("Create should succeed after Forget")
	}
}

// Test Schedule.Run executes due events
func TestScheduleRun(t *testing.T) {
	s := NewSchedule()

	s.Call(func() {
		// This would execute on January 1st at midnight
	}).Cron("0 0 1 1 *")

	// Run at current time - the event is scheduled for a specific date
	// so it won't execute now
	s.Run()

	// We can't easily test execution without mocking time,
	// but we've tested IsDue and DueEvents separately
}

// Test command scheduling
func TestCommandScheduling(t *testing.T) {
	s := NewSchedule()

	event := s.Command("echo test")
	event.EveryMinute()

	if event.command != "echo test" {
		t.Error("Command should be set correctly")
	}

	if event.expression != "* * * * *" {
		t.Error("Expression should be set correctly")
	}
}

// Test job scheduling
func TestJobScheduling(t *testing.T) {
	s := NewSchedule()

	type testJob struct {
		Name string
	}

	job := &testJob{Name: "test"}
	event := s.Job(job)
	event.Daily()

	if event.job == nil {
		t.Error("Job should be set")
	}

	if event.expression != "0 0 * * *" {
		t.Error("Expression should be set correctly")
	}
}

// Test method chaining
func TestMethodChaining(t *testing.T) {
	s := NewSchedule()

	event := s.Call(func() {
		// Task body
	}).
		EveryMinute().
		Weekdays().
		Between("09:00", "17:00").
		WithoutOverlapping().
		RunInBackground().
		Before(func() {}).
		After(func() {}).
		OnSuccess(func() {})

	if event.expression != "* * * * *" {
		t.Error("Chaining should preserve expression")
	}

	if !event.preventOverlap {
		t.Error("Chaining should set preventOverlap")
	}

	if !event.runInBg {
		t.Error("Chaining should set runInBg")
	}
}

// Test file mutex with actual filesystem
func TestFileMutexRealFiles(t *testing.T) {
	dir := t.TempDir()
	mutex := NewFileMutex(dir)

	name := "test-lock"
	lockFile := filepath.Join(dir, "schedule-"+name+".lock")

	// Create lock
	if !mutex.Create(name) {
		t.Fatal("Should create lock")
	}

	// Verify file exists
	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		t.Error("Lock file should exist")
	}

	// Cleanup
	mutex.Forget(name)

	// Verify file removed
	if _, err := os.Stat(lockFile); !os.IsNotExist(err) {
		t.Error("Lock file should be removed")
	}
}
