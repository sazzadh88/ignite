package ratelimit

// Limit represents rate limit configuration.
type Limit struct {
	// MaxAttempts is the maximum number of attempts allowed.
	MaxAttempts int

	// DecayMinutes is the number of minutes until the limit resets.
	DecayMinutes int

	// Key is the key used to identify the requester (e.g., IP address, user ID).
	Key string
}

// PerMinute creates a rate limit of maxAttempts per minute.
func PerMinute(max int) *Limit {
	return &Limit{
		MaxAttempts:  max,
		DecayMinutes: 1,
	}
}

// PerHour creates a rate limit of maxAttempts per hour.
func PerHour(max int) *Limit {
	return &Limit{
		MaxAttempts:  max,
		DecayMinutes: 60,
	}
}

// PerDay creates a rate limit of maxAttempts per day.
func PerDay(max int) *Limit {
	return &Limit{
		MaxAttempts:  max,
		DecayMinutes: 1440, // 24 * 60
	}
}

// None creates an unlimited rate limit (no restriction).
func None() *Limit {
	return &Limit{
		MaxAttempts:  -1, // Negative value indicates unlimited
		DecayMinutes: 0,
	}
}

// By sets the key used to identify the requester.
// This allows chaining: PerMinute(60).By("api:" + apiKey)
func (l *Limit) By(key string) *Limit {
	l.Key = key
	return l
}
