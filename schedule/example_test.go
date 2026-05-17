package schedule_test

import (
	"fmt"
	"time"

	"github.com/sazzad/ignite/schedule"
)

func ExampleSchedule_Call() {
	s := schedule.NewSchedule()

	// Schedule a callback to run every minute
	s.Call(func() {
		fmt.Println("Running every minute")
	}).EveryMinute()

	// Schedule with constraints
	s.Call(func() {
		fmt.Println("Running on weekdays during business hours")
	}).
		EveryFiveMinutes().
		Weekdays().
		Between("09:00", "17:00")
}

func ExampleSchedule_Command() {
	s := schedule.NewSchedule()

	// Schedule a shell command to run daily
	s.Command("php artisan inspire").Daily()

	// Run at a specific time
	s.Command("backup:run").DailyAt("02:00")
}

func ExampleEvent_Cron() {
	s := schedule.NewSchedule()

	// Use a custom cron expression
	s.Call(func() {
		fmt.Println("Custom cron schedule")
	}).Cron("*/15 9-17 * * 1-5") // Every 15 minutes, 9am-5pm, Mon-Fri
}

func ExampleEvent_frequency() {
	s := schedule.NewSchedule()

	// Various frequency options
	s.Call(func() {}).EveryMinute()
	s.Call(func() {}).EveryFiveMinutes()
	s.Call(func() {}).EveryFifteenMinutes()
	s.Call(func() {}).Hourly()
	s.Call(func() {}).HourlyAt(30) // At 30 minutes past the hour
	s.Call(func() {}).Daily()
	s.Call(func() {}).DailyAt("13:30")
	s.Call(func() {}).TwiceDaily(9, 17) // 9am and 5pm
	s.Call(func() {}).Weekly()
	s.Call(func() {}).WeeklyOn(1, "09:00") // 1 = Monday
	s.Call(func() {}).Monthly()
	s.Call(func() {}).MonthlyOn(15, "14:00")
	s.Call(func() {}).Quarterly()
	s.Call(func() {}).Yearly()
}

func ExampleEvent_constraints() {
	s := schedule.NewSchedule()

	// Time window constraints
	s.Call(func() {}).
		Hourly().
		Between("09:00", "17:00"). // Only during business hours
		UnlessBetween("12:00", "13:00") // Except lunch time

	// Day of week constraints
	s.Call(func() {}).Daily().Weekdays() // Monday-Friday
	s.Call(func() {}).Daily().Weekends() // Saturday-Sunday
	s.Call(func() {}).Daily().Mondays()  // Only Monday
	s.Call(func() {}).Daily().Days(time.Monday, time.Wednesday, time.Friday)

	// Conditional execution
	s.Call(func() {}).
		Daily().
		When(func() bool {
			// Only run if condition is true
			return true
		})

	s.Call(func() {}).
		Daily().
		Skip(func() bool {
			// Skip if condition is true
			return false
		})

	// Prevent overlapping runs
	s.Call(func() {
		// Long-running task
		time.Sleep(5 * time.Minute)
	}).EveryMinute().WithoutOverlapping()

	// Run in background
	s.Call(func() {
		// Task that can run concurrently
	}).Hourly().RunInBackground()
}

func ExampleEvent_hooks() {
	s := schedule.NewSchedule()

	s.Call(func() {
		fmt.Println("Main task")
	}).
		Daily().
		Before(func() {
			fmt.Println("Before task")
		}).
		After(func() {
			fmt.Println("After task")
		}).
		OnSuccess(func() {
			fmt.Println("Task succeeded")
		}).
		OnFailure(func(err error) {
			fmt.Printf("Task failed: %v\n", err)
		})
}

func ExampleSchedule_Run() {
	s := schedule.NewSchedule()

	// Register some tasks
	s.Call(func() {
		fmt.Println("Task 1")
	}).EveryMinute()

	s.Call(func() {
		fmt.Println("Task 2")
	}).Hourly()

	// Run all due tasks (typically called by scheduler daemon every minute)
	s.Run()
}

func ExampleSchedule_DueEvents() {
	s := schedule.NewSchedule()

	s.Call(func() {}).Hourly()
	s.Call(func() {}).Daily()

	// Get events that are due at a specific time
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	due := s.DueEvents(now)

	fmt.Printf("Due events at %v: %d\n", now, len(due))
	// Output: Due events at 2024-01-01 12:00:00 +0000 UTC: 1
}

func ExampleCronExpression() {
	// Parse a cron expression
	expr, err := schedule.Parse("*/5 * * * *")
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	// Check if a time matches
	t := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if expr.Matches(t) {
		fmt.Println("Time matches cron expression")
	}
	// Output: Time matches cron expression
}

func ExampleFileMutex() {
	// Create a file-based mutex
	mutex := schedule.NewFileMutex("/tmp/schedule-locks")

	// Try to acquire lock
	if mutex.Create("my-task") {
		fmt.Println("Lock acquired")

		// Do work...

		// Release lock
		mutex.Forget("my-task")
	} else {
		fmt.Println("Lock already held")
	}
}

func ExampleMemoryMutex() {
	// Create an in-memory mutex
	mutex := schedule.NewMemoryMutex()

	// Try to acquire lock
	if mutex.Create("my-task") {
		fmt.Println("Lock acquired")

		// Do work...

		// Release lock
		mutex.Forget("my-task")
	} else {
		fmt.Println("Lock already held")
	}
	// Output: Lock acquired
}
