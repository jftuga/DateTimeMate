package DateTimeMate

import "testing"

func TestShrinkPeriod(t *testing.T) {
	// explicitly define these here as a sanity check vs just iterating through abbrevMap
	var allLongPeriods = []string{"nanoseconds", "microseconds", "milliseconds", "seconds", "minutes", "hours", "days", "weeks", "months", "years"}
	var allShortPeriods = []string{"ns", "us", "ms", "s", "m", "h", "D", "W", "M", "Y"}

	for i, period := range allLongPeriods {
		shrunk := shrinkPeriod(period)
		if shrunk != allShortPeriods[i] {
			t.Errorf("[computed: %v] != [correct: %v]", shrunk, allShortPeriods[i])
		}
	}

	for i, period := range allLongPeriods {
		period = removeTrailingS(period)
		shrunk := shrinkPeriod(period)
		if shrunk != allShortPeriods[i] {
			t.Errorf("[computed: %v] != [correct: %v]", shrunk, allShortPeriods[i])
		}
	}
}

func testFormat(t *testing.T, source, outputFormat, correct string) {
	t.Helper()
	computed, err := Reformat(source, outputFormat)
	if err != nil {
		t.Error(err)
	}
	if computed != correct {
		t.Errorf("[computed: %v] != [correct: %v]", computed, correct)
	}
}

func TestFormatCommand(t *testing.T) {
	source := "2024-07-22 08:21:44"
	fmt := "%T %D"
	correct := "08:21:44 07/22/24"
	testFormat(t, source, fmt, correct)

	fmt = "%v %r"
	correct = "22-Jul-2024 08:21:44 AM"
	testFormat(t, source, fmt, correct)

	fmt = "%Y%m%d.%H%M%S"
	correct = "20240722.082144"
	testFormat(t, source, fmt, correct)

	source = "2024-02-29T23:59:59Z"
	fmt = "%Y%m%d.%H%M%S"
	correct = "20240229.235959"
	testFormat(t, source, fmt, correct)

	fmt = "%Z"
	correct = "UTC"
	testFormat(t, source, fmt, correct)

	source = "Mon Jul 22 08:40:33 EDT 2024"
	fmt = "%Z %z"
	correct = "EDT -0400"
	testFormat(t, source, fmt, correct)
}
