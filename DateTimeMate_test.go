package DateTimeMate

import (
	"testing"
)

func TestShrinkPeriod(t *testing.T) {
	var allLongPeriods = []string{"nanoseconds", "microseconds", "milliseconds", "seconds", "minutes", "hours", "days", "weeks", "months", "years"}
	var allShortPeriods = []string{"ns", "us", "ms", "s", "m", "h", "D", "W", "M", "Y"}

	for i, period := range allLongPeriods {
		shrunk := shrinkPeriod(period)
		if shrunk != allShortPeriods[i] {
			t.Errorf("[computed: %v] != [correct: %v]", shrunk, allShortPeriods[i])
		}
	}

	for i, period := range allLongPeriods {
		period = period[:len(period)-1]
		shrunk := shrinkPeriod(period)
		if shrunk != allShortPeriods[i] {
			t.Errorf("[computed: %v] != [correct: %v]", shrunk, allShortPeriods[i])
		}
	}
}
