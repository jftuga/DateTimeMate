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
	// 11 and 12 digit integers are neither seconds nor milliseconds; a
	// 14-digit integer is a compact date/time, so one that is not a valid
	// date/time errors instead of falling through to a lenient parser
	for _, source := range []string{"12345678901", "123456789012", "12345678901234", "123456"} {
		if _, err := Reformat(source, "%F"); err == nil {
			t.Errorf("expected an error for timestamp %q, got nil", source)
		}
	}
}

func TestReformatIntegerDate(t *testing.T) {
	// pure integers that are not unix timestamps parse as compact
	// date/times instead of being misread as epoch seconds
	testFormat(t, "2024", "%Y", "2024")
	testFormat(t, "20240101", "%F", "2024-01-01")
	testFormat(t, "20240101080102", "%F %T", "2024-01-01 08:01:02")
}

func TestNegativeTimestampRejected(t *testing.T) {
	// negative integers are rejected everywhere as negative timestamps
	for _, source := range []string{"-170000000", "-1700265600", "-17000000000"} {
		if _, err := Reformat(source, "%s"); err == nil {
			t.Errorf("Reformat: expected an error for %q, got nil", source)
		}
		dur := NewDur(DurWithFrom(source), DurWithDur("1 second"))
		if _, err := dur.Add(); err == nil {
			t.Errorf("Dur: expected an error for %q, got nil", source)
		}
		diff := NewDiff(DiffWithStart(source), DiffWithEnd("2024-01-01"))
		if _, _, err := diff.CalculateDiff(); err == nil {
			t.Errorf("Diff: expected an error for %q, got nil", source)
		}
	}

	// the ambiguous-length error counts digits, not characters: an 11-digit
	// negative value must be rejected as negative, never reported as length 12
	_, err := Reformat("-17000000000", "%s")
	if err == nil || !strings.Contains(err.Error(), "negative") {
		t.Errorf("expected a negative-timestamp error, got: %v", err)
	}
}

func TestTimestampLeadingWhitespace(t *testing.T) {
	// leading whitespace must not bypass the unix-timestamp gate and let a
	// lenient parser misread the digits as a time of day
	testFormat(t, " 1700265600", "%s", "1700265600")

	dur := NewDur(DurWithFrom(" 1700265600"), DurWithDur("0 seconds"), DurWithOutputFormat("%s"))
	future, err := dur.Add()
	if err != nil {
		t.Fatal(err)
	}
	if len(future) != 1 || future[0] != "1700265600" {
		t.Errorf("[computed: %v] != [correct: 1700265600]", future)
	}

	diff := NewDiff(DiffWithStart(" 1700265600"), DiffWithEnd("1700265600"))
	_, duration, err := diff.CalculateDiff()
	if err != nil {
		t.Fatal(err)
	}
	if duration != 0 {
		t.Errorf("[computed: %v] != [correct: 0s]", duration)
	}
}

func TestSlashDateOrder(t *testing.T) {
	// no t.Parallel: t.Setenv is process-wide
	// ambiguous slash dates default to month/day/year
	t.Setenv(DateOrderEnvVar, "")
	testFormat(t, "01/02/2024", "%F", "2024-01-02")
	testFormat(t, "1/2/2024", "%F", "2024-01-02")
	testFormat(t, "01/02/2024 08:30", "%F %H:%M", "2024-01-02 08:30")
	testFormat(t, "01/02/2024 08:30:15", "%F %T", "2024-01-02 08:30:15")

	// DMY flips ambiguous dates to day/month/year
	t.Setenv(DateOrderEnvVar, "DMY")
	testFormat(t, "01/02/2024", "%F", "2024-02-01")
	testFormat(t, "1/2/2024", "%F", "2024-02-01")

	// a field greater than 12 disambiguates on its own, regardless of the variable
	testFormat(t, "25/12/2024", "%F", "2024-12-25")
	t.Setenv(DateOrderEnvVar, "MDY")
	testFormat(t, "25/12/2024", "%F", "2024-12-25")
	testFormat(t, "12/25/2024", "%F", "2024-12-25")

	// an invalid value errors only when an ambiguous date is actually parsed
	t.Setenv(DateOrderEnvVar, "YMD")
	if _, err := Reformat("01/02/2024", "%F"); err == nil {
		t.Error("expected an error for an invalid date-order value, got nil")
	}
	testFormat(t, "25/12/2024", "%F", "2024-12-25")

	// neither field can be a month
	t.Setenv(DateOrderEnvVar, "")
	if _, err := Reformat("13/13/2024", "%F"); err == nil {
		t.Error("expected an error when neither slash-date field can be a month, got nil")
	}
}

func TestParserAgreement(t *testing.T) {
	// no t.Parallel: TestSlashDateOrder mutates the date-order variable
	// Diff and Dur/Reformat must assign the same instant to the same input
	// (they used to try carbon and parsetime in opposite orders)
	for _, source := range []string{"01/02/2024", "1/2/2024", "2024-01", "20240101080102", "Jan 2, 2024"} {
		viaReformat, err := Reformat(source, "%s")
		if err != nil {
			t.Fatalf("Reformat(%q): %v", source, err)
		}
		sec, err := strconv.ParseInt(viaReformat, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		diff := NewDiff(DiffWithStart(source), DiffWithEnd(strconv.FormatInt(sec, 10)))
		_, duration, err := diff.CalculateDiff()
		if err != nil {
			t.Fatalf("Diff(%q): %v", source, err)
		}
		if duration != 0 {
			t.Errorf("parsers disagree on %q by %v", source, duration)
		}
	}
}

func TestExplicitZonePreservedAcrossDST(t *testing.T) {
	// fixLocalZone used to reinterpret times whose explicit zone matched
	// today's abbreviation, shifting instants across a DST boundary; probe
	// with today's zone abbreviation on a date in the opposite DST regime,
	// using a format that is not covered by zonedLayouts
	now := time.Now()
	name, offset := now.Zone()
	var probe time.Time
	for month := time.January; month <= time.December; month++ {
		candidate := time.Date(now.Year(), month, 15, 12, 0, 0, 0, time.Local)
		if _, candidateOffset := candidate.Zone(); candidateOffset != offset {
			probe = candidate
			break
		}
	}
	if probe.IsZero() {
		t.Skip("local time zone does not observe DST")
	}
	// e.g. "Jan 15 12:00:00 EDT 2026" while today's zone is EDT: the
	// explicit zone must be honored even though that date is in EST
	source := probe.Format("Jan 2 15:04:05") + " " + name + " " + probe.Format("2006")
	correct := time.Date(probe.Year(), probe.Month(), probe.Day(), 12, 0, 0, 0, time.FixedZone(name, offset))
	computed, err := Reformat(source, "%s")
	if err != nil {
		t.Fatal(err)
	}
	if computed != strconv.FormatInt(correct.Unix(), 10) {
		t.Errorf("[computed: %v] != [correct: %v] for %q", computed, correct.Unix(), source)
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

func TestRelativeDatesExactly24Hours(t *testing.T) {
	// yesterday and tomorrow are exactly -/+ 24 hours of the current time,
	// as documented, even across DST transitions (calendar-day arithmetic
	// would be 23 or 25 real hours on the two transition days); the
	// expansions come from separate Now() calls and the strings are
	// second-granular, so allow a small tolerance
	yesterday, err := parseDateTime(ConvertRelativeDateToActual("yesterday"))
	if err != nil {
		t.Fatal(err)
	}
	tomorrow, err := parseDateTime(ConvertRelativeDateToActual("tomorrow"))
	if err != nil {
		t.Fatal(err)
	}
	span := tomorrow.Sub(yesterday)
	if span < 48*time.Hour-2*time.Second || span > 48*time.Hour+2*time.Second {
		t.Errorf("yesterday to tomorrow spans %v, expected 48h", span)
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
