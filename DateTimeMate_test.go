package DateTimeMate

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestReformatWinterUnixSeconds(t *testing.T) {
	// a zone-less winter date must convert to the unix timestamp using the
	// local UTC offset in effect on that date, not the offset in effect today
	correct, err := time.ParseInLocation("2006-01-02 15:04:05", "2024-01-01 00:00:00", time.Local)
	if err != nil {
		t.Fatal(err)
	}
	computed, err := Reformat("2024-01-01 00:00:00", "%s")
	if err != nil {
		t.Fatal(err)
	}
	if computed != strconv.FormatInt(correct.Unix(), 10) {
		t.Errorf("[computed: %v] != [correct: %v]", computed, correct.Unix())
	}
}

func TestReformatAmbiguousTimestampLength(t *testing.T) {
	// 11, 12, and 14+ digit integers are neither seconds nor milliseconds
	for _, source := range []string{"12345678901", "123456789012", "12345678901234"} {
		if _, err := Reformat(source, "%F"); err == nil {
			t.Errorf("expected an error for timestamp %q, got nil", source)
		}
	}
}

func TestShrinkPeriod(t *testing.T) {
	// explicitly define these here as a sanity check vs just iterating through abbrevMap
	var allLongPeriods = []string{"nanoseconds", "microseconds", "milliseconds", "seconds", "minutes", "hours", "days", "weeks", "years"}
	var allShortPeriods = []string{"ns", "us", "ms", "s", "m", "h", "D", "W", "Y"}

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

	source = "2024-11-16T14:01:02-05:00"
	fmt = "%s"
	correct = "1731783662"
	testFormat(t, source, fmt, correct)

	source = "2024-06-16T14:01:02-04:00"
	fmt = "%s"
	correct = "1718560862"
	testFormat(t, source, fmt, correct)

	// unix timestamps render in local time, so compute the expected
	// value dynamically to keep this test timezone-independent
	source = "1704085262"
	fmt = "%F %T"
	correct = time.Unix(1704085262, 0).Format("2006-01-02 15:04:05")
	testFormat(t, source, fmt, correct)

	source = "1704085262999"
	fmt = "%F %T"
	testFormat(t, source, fmt, correct)
}

func TestParseDateTimePre1970(t *testing.T) {
	// parsetime silently corrupts pre-1970 date/times (e.g. 1950-01-01
	// became 2026-01-09); these must parse correctly via standard layouts
	testFormat(t, "1950-01-01 12:00:00", "%Y-%m-%d %H:%M:%S", "1950-01-01 12:00:00")
	testFormat(t, "1900-02-28 23:59:59", "%Y-%m-%d %H:%M:%S", "1900-02-28 23:59:59")
	testFormat(t, "1969-12-31T16:00:00Z", "%Y-%m-%d %H:%M:%S %Z", "1969-12-31 16:00:00 UTC")

	dur := NewDur(DurWithFrom("1969-12-31 16:00:00"), DurWithDur("1h"))
	future, err := dur.Add()
	if err != nil {
		t.Fatal(err)
	}
	if len(future) != 1 || !strings.Contains(future[0], "1969-12-31 17:00:00") {
		t.Errorf("[computed: %v] does not contain: [correct: 1969-12-31 17:00:00]", future)
	}
}

func TestParsedYearMismatch(t *testing.T) {
	jan2026 := time.Date(2026, 1, 9, 0, 0, 0, 0, time.UTC)
	// a corrupted parsetime result names a year absent from the input
	if !parsedYearMismatch("Mon Feb 28 23:59:59 EST 1900", jan2026) {
		t.Error("expected a mismatch for a 1900 input parsed as 2026")
	}
	// the year matching anywhere in the input is accepted
	if parsedYearMismatch("2026-01-09 00:00:00", jan2026) {
		t.Error("expected no mismatch when the input contains the parsed year")
	}
	// inputs without a year (time-only, fractional seconds) are never a mismatch
	if parsedYearMismatch("12:34:56", jan2026) {
		t.Error("expected no mismatch for a time-only input")
	}
	if parsedYearMismatch("12:34:56.1234", jan2026) {
		t.Error("expected no mismatch for fractional seconds")
	}
}
