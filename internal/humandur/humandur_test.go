package humandur

import (
	"math"
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"zero", 0, "0 seconds"},
		{"one second", time.Second, "1 second"},
		{"plural seconds", 2 * time.Second, "2 seconds"},
		{"one of each singular", 365*24*time.Hour + 7*24*time.Hour + 24*time.Hour + time.Hour + time.Minute + time.Second + time.Millisecond + time.Microsecond + time.Nanosecond,
			"1 year 1 week 1 day 1 hour 1 minute 1 second 1 millisecond 1 microsecond 1 nanosecond"},
		{"all units plural", 2*365*24*time.Hour + 2*7*24*time.Hour + 2*24*time.Hour + 2*time.Hour + 2*time.Minute + 2*time.Second + 2*time.Millisecond + 2*time.Microsecond + 2*time.Nanosecond,
			"2 years 2 weeks 2 days 2 hours 2 minutes 2 seconds 2 milliseconds 2 microseconds 2 nanoseconds"},
		{"zero components omitted", 31*24*time.Hour + 62*time.Minute + 3*time.Second, "4 weeks 3 days 1 hour 2 minutes 3 seconds"},
		{"negative", -(3*time.Hour + 30*time.Minute + 45*time.Second), "-3 hours 30 minutes 45 seconds"},
		{"sub-microsecond only", 500 * time.Nanosecond, "500 nanoseconds"},
		{"single nanosecond", time.Nanosecond, "1 nanosecond"},
		{"microsecond plus nanoseconds", 1500 * time.Nanosecond, "1 microsecond 500 nanoseconds"},
		{"negative sub-microsecond", -500 * time.Nanosecond, "-500 nanoseconds"},
		{"exactly one year", 365 * 24 * time.Hour, "1 year"},
		{"one day short of a year", 364 * 24 * time.Hour, "52 weeks"},
		{"366 days is a year and a day", 366 * 24 * time.Hour, "1 year 1 day"},
		{"exactly one week", 7 * 24 * time.Hour, "1 week"},
		{"boundary below a minute", 59 * time.Second, "59 seconds"},
		{"boundary at a minute", 60 * time.Second, "1 minute"},
		{"minimum duration does not overflow negation", time.Duration(math.MinInt64),
			"-292 years 24 weeks 3 days 23 hours 47 minutes 16 seconds 854 milliseconds 775 microseconds 808 nanoseconds"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Format(tt.d); got != tt.want {
				t.Errorf("Format(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}
