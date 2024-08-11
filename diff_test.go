package DateTimeMate

import "testing"

func testDiffStartEnd(t *testing.T, start, end string, brief bool, correct string) {
	t.Helper()
	diff := NewDiff(
		DiffWithStart(start),
		DiffWithEnd(end),
		DiffWithBrief(brief))
	result, _, err := diff.CalculateDiff()
	if err != nil {
		t.Error(err)
	}
	if result != correct {
		t.Errorf("[computed: %v] != [correct: %v]", result, correct)
	}
}

func TestDiffTwoTimesSameDay(t *testing.T) {
	t.Parallel()
	start := "12:00:00"
	end := "15:30:45"
	correct := "3 hours 30 minutes 45 seconds"
	correctBrief := "3h30m45s"
	testDiffStartEnd(t, start, end, false, correct)
	testDiffStartEnd(t, start, end, true, correctBrief)
}

func TestDiffAmPm(t *testing.T) {
	t.Parallel()
	start := "11:00AM"
	end := "11:00PM"
	correct := "12 hours"
	correctBrief := "12h"
	testDiffStartEnd(t, start, end, false, correct)
	testDiffStartEnd(t, start, end, true, correctBrief)
}

func TestDiffIso8601(t *testing.T) {
	t.Parallel()
	start := "2024-06-07T08:00:00Z"
	end := "2024-07-08T09:02:03Z"
	correct := "4 weeks 3 days 1 hour 2 minutes 3 seconds"
	correctBrief := "4W3D1h2m3s"
	testDiffStartEnd(t, start, end, false, correct)
	testDiffStartEnd(t, start, end, true, correctBrief)
}

func TestDiffTimeZoneOffset(t *testing.T) {
	t.Parallel()
	start := "2024-06-07T08:00:00Z"
	end := "2024-06-07T08:05:05-05:00"
	correct := "5 hours 5 minutes 5 seconds"
	correctBrief := "5h5m5s"
	testDiffStartEnd(t, start, end, false, correct)
	testDiffStartEnd(t, start, end, true, correctBrief)
}

func TestDiffIncludeSpaces(t *testing.T) {
	t.Parallel()
	start := "2024-06-07 08:01:02"
	end := "2024-06-07 08:02"
	correct := "58 seconds"
	correctBrief := "58s"
	testDiffStartEnd(t, start, end, false, correct)
	testDiffStartEnd(t, start, end, true, correctBrief)
}

func TestDiffMicroSeconds(t *testing.T) {
	t.Parallel()
	start := "2024-06-07T08:00:00Z"
	end := "2024-06-07T08:00:00.000123Z"
	correct := "123 microseconds"
	correctBrief := "123us"
	testDiffStartEnd(t, start, end, false, correct)
	testDiffStartEnd(t, start, end, true, correctBrief)
}

func TestDiffMilliSeconds(t *testing.T) {
	t.Parallel()
	start := "2024-06-07T08:00:00Z"
	end := "2024-06-07T08:01:02.345Z"
	correct := "1 minute 2 seconds 345 milliseconds"
	correctBrief := "1m2s345ms"
	testDiffStartEnd(t, start, end, false, correct)
	testDiffStartEnd(t, start, end, true, correctBrief)
}

func TestDiffRelativeStartEnd(t *testing.T) {
	t.Parallel()
	start := "yesterday"
	end := "Today"
	correct := "1 day"
	testDiffStartEnd(t, start, end, false, correct)

	start = "Yesterday"
	end = "tomorrow"
	correct = "2 days"
	testDiffStartEnd(t, start, end, false, correct)

	start = "now"
	end = "today"
	correct = "0 seconds"
	testDiffStartEnd(t, start, end, false, correct)

	start = "today"
	end = "tomorrow"
	correct = "1 day"
	testDiffStartEnd(t, start, end, false, correct)
}
