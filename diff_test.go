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

// testDiffAbsolute mirrors testDiffStartEnd but also exercises the Absolute
// option and asserts the sign of the returned time.Duration
func testDiffAbsolute(t *testing.T, start, end string, brief, absolute bool, correct string, wantNegative bool) {
	t.Helper()
	diff := NewDiff(
		DiffWithStart(start),
		DiffWithEnd(end),
		DiffWithBrief(brief),
		DiffWithAbsolute(absolute))
	result, duration, err := diff.CalculateDiff()
	if err != nil {
		t.Error(err)
	}
	if result != correct {
		t.Errorf("[computed: %v] != [correct: %v]", result, correct)
	}
	if wantNegative && duration >= 0 {
		t.Errorf("expected a negative duration, got: %v", duration)
	}
	if !wantNegative && duration < 0 {
		t.Errorf("expected a non-negative duration, got: %v", duration)
	}
}

func TestDiffSignedResult(t *testing.T) {
	t.Parallel()
	// the difference is signed when the start is later than the end
	testDiffAbsolute(t, "15:30:45", "12:00:00", false, false, "-3 hours 30 minutes 45 seconds", true)
	testDiffAbsolute(t, "15:30:45", "12:00:00", true, false, "-3h30m45s", true)
}

func TestDiffAbsoluteResult(t *testing.T) {
	t.Parallel()
	// Absolute makes both the formatted string and the returned duration positive
	testDiffAbsolute(t, "15:30:45", "12:00:00", false, true, "3 hours 30 minutes 45 seconds", false)
	testDiffAbsolute(t, "15:30:45", "12:00:00", true, true, "3h30m45s", false)
	// Absolute is a no-op when the result is already positive
	testDiffAbsolute(t, "12:00:00", "15:30:45", false, true, "3 hours 30 minutes 45 seconds", false)
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

func TestDiffUnixTimestamps(t *testing.T) {
	t.Parallel()
	// 10-digit timestamps are seconds; 13-digit timestamps are milliseconds
	testDiffStartEnd(t, "1700000000", "1700003600", false, "1 hour")
	testDiffStartEnd(t, "1700000000000", "1700000000500", false, "500 milliseconds")
}

func TestDiffAmbiguousTimestamp(t *testing.T) {
	t.Parallel()
	// 11 and 12 digit integers are neither seconds (10) nor milliseconds (13)
	// and previously fell through to the date/time parsers
	for _, start := range []string{"17000000001", "170000000012"} {
		diff := NewDiff(
			DiffWithStart(start),
			DiffWithEnd("1700003600"))
		if _, _, err := diff.CalculateDiff(); err == nil {
			t.Errorf("expected an error for ambiguous timestamp %q, got nil", start)
		}
	}
}

func TestDiffYearOverflow(t *testing.T) {
	t.Parallel()
	diff := NewDiff(
		DiffWithStart("0001-10-19"),
		DiffWithEnd("2000-10-10"))
	_, _, err := diff.CalculateDiff()
	if err == nil {
		t.Error("expected error for year difference exceeding 291 years")
	}
}

func TestDiffPre1970(t *testing.T) {
	t.Parallel()
	// both sides before 1970
	testDiffStartEnd(t, "1950-01-01 12:00:00", "1950-01-02 13:30:00", false, "1 day 1 hour 30 minutes")
	testDiffStartEnd(t, "1950-01-01 12:00:00", "1950-01-02 13:30:00", true, "1D1h30m")
	// pre-1970 with explicit zones
	testDiffStartEnd(t, "1950-06-07T08:00:00Z", "1950-06-07T09:02:03Z", false, "1 hour 2 minutes 3 seconds")
	// crossing the unix epoch
	testDiffStartEnd(t, "1969-12-31 23:00:00", "1970-01-01 01:00:00", false, "2 hours")
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
