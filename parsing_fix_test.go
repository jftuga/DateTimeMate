package DateTimeMate

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTwoDigitYearSlashDates(t *testing.T) {
	// no t.Parallel: t.Setenv is process-wide
	// two-digit years must go through the same slash-date machinery as
	// four-digit years; they used to fall through to parsetime, which
	// ignored DateOrderEnvVar and replaced the year of "12/31/24" with the
	// current year
	t.Setenv(DateOrderEnvVar, "MDY")
	testFormat(t, "1/2/24", "%F", "2024-01-02")
	testFormat(t, "12/31/24", "%F", "2024-12-31")
	testFormat(t, "31/12/24", "%F", "2024-12-31")
	testFormat(t, "1/2/24 08:30", "%F %H:%M", "2024-01-02 08:30")

	// Go's "06" layout convention: 69-99 are 19xx
	testFormat(t, "7/4/76", "%F", "1976-07-04")

	t.Setenv(DateOrderEnvVar, "DMY")
	testFormat(t, "1/2/24", "%F", "2024-02-01")

	// neither field can be a month
	if _, err := Reformat("13/13/24", "%F"); err == nil {
		t.Error("expected an error when neither slash-date field can be a month, got nil")
	}
}

func TestOutOfRangeDateTimeRejected(t *testing.T) {
	// no t.Parallel: t.Setenv is process-wide
	// components the strict layouts reject as out of range must error
	// instead of falling through to parsers that silently normalize
	// ("2024-02-30" became March 1) or mangle ("08:61:00" became "08:06:01")
	t.Setenv(DateOrderEnvVar, "MDY")
	invalid := []string{
		"2024-02-30",
		"2024-04-31",
		"2024-00-15",
		"2024-01-15 08:61:00",
		"2024-01-15 08:30:61",
		"2024-01-15 25:00:00",
		"2024-02-30T12:00:00Z",
		"2/30/2024",
		"1/32/2024",
		"1/2/2024 13:04PM",
		"1/2/2024 13:04 PM",
	}
	for _, source := range invalid {
		if _, err := Reformat(source, "%F"); err == nil {
			t.Errorf("expected an error for %q, got nil", source)
		}
	}

	// valid edge values must still parse
	testFormat(t, "2024-02-29", "%F", "2024-02-29")
	testFormat(t, "2024-12-31 23:59:59", "%F %T", "2024-12-31 23:59:59")

	// the same rejection must hold when a time zone conversion parses the
	// wall clock
	conv := setupConverter()
	if _, err := conv.ConvertTimeZone("2024-02-30 12:00:00 EST", "UTC"); err == nil {
		t.Error("expected an error for an out-of-range date in tz, got nil")
	}
}

func TestEmptyDateTimeRejected(t *testing.T) {
	// empty and whitespace-only date/times used to silently parse as the
	// current time
	for _, source := range []string{"", "   "} {
		if _, err := Reformat(source, "%F"); err == nil {
			t.Errorf("Reformat: expected an error for %q, got nil", source)
		}
	}
	dur := NewDur(DurWithFrom(""), DurWithDur("1 hour"))
	if _, err := dur.Add(); err == nil {
		t.Error("Dur: expected an error for an empty From, got nil")
	}
	diff := NewDiff(DiffWithStart(""), DiffWithEnd("2024-01-01"))
	if _, _, err := diff.CalculateDiff(); err == nil {
		t.Error("Diff: expected an error for an empty Start, got nil")
	}
	diff = NewDiff(DiffWithStart("2024-01-01"), DiffWithEnd(" "))
	if _, _, err := diff.CalculateDiff(); err == nil {
		t.Error("Diff: expected an error for a blank End, got nil")
	}
}

func TestZuluWithContradictoryZoneRejected(t *testing.T) {
	// a trailing "Z" is an explicit UTC designator: combining it with a
	// different trailing zone used to silently discard the Z and reinterpret
	// the wall clock in that zone
	conv := setupConverter()
	for _, source := range []string{"2024-01-15T08:30:00Z EST", "2024-01-15T08:30:00.5Z EST"} {
		_, err := conv.ConvertTimeZone(source, "UTC")
		if err == nil {
			t.Errorf("expected an error for %q, got nil", source)
			continue
		}
		if !strings.Contains(err.Error(), "names both") {
			t.Errorf("expected a contradictory-zone error for %q, got: %v", source, err)
		}
	}

	// agreeing zones must still convert
	result, err := conv.ConvertTimeZone("2024-01-15T08:30:00Z UTC", "America/New_York")
	assert.NoError(t, err)
	assert.Equal(t, "2024-01-15 03:30:00 EST", result.Format("2006-01-02 15:04:05 MST"))
}

func TestAbbreviationsKeepFixedOffsets(t *testing.T) {
	// CET, EET, and WET are also IANA legacy zones with DST rules; as
	// abbreviations they must mean their fixed table offsets on any date,
	// otherwise "08:30 CET" on a summer date silently becomes 08:30 CEST
	conv := setupConverter()
	tests := []struct {
		sourceTime string
		expected   string
	}{
		{"2024-07-15 08:30:00 CET", "2024-07-15 07:30:00"},
		{"2024-07-15 08:30:00 EET", "2024-07-15 06:30:00"},
		{"2024-07-15 08:30:00 WET", "2024-07-15 08:30:00"},
		{"2024-01-15 08:30:00 CET", "2024-01-15 07:30:00"},
	}
	for _, tt := range tests {
		result, err := conv.ConvertTimeZone(tt.sourceTime, "UTC")
		assert.NoError(t, err)
		assert.Equal(t, tt.expected, result.Format("2006-01-02 15:04:05"), "source: %s", tt.sourceTime)
	}

	// full IANA names stay DST-aware
	result, err := conv.ConvertTimeZone("2024-07-15 08:30:00 Europe/Paris", "UTC")
	assert.NoError(t, err)
	assert.Equal(t, "2024-07-15 06:30:00", result.Format("2006-01-02 15:04:05"))
}

func TestDurCaseInsensitiveLongUnits(t *testing.T) {
	// conv and durmath already accept any case for long-form units; dur
	// used to reject anything but lowercase
	for _, period := range []string{"1 hour 30 minutes", "1 Hour 30 Minutes", "1 HOUR 30 MINUTES"} {
		dur := NewDur(DurWithFrom("2024-01-15 08:00:00"), DurWithDur(period), DurWithOutputFormat("%F %T"))
		add, err := dur.Add()
		if err != nil {
			t.Fatalf("Dur(%q): %v", period, err)
		}
		if len(add) != 1 || add[0] != "2024-01-15 09:30:00" {
			t.Errorf("Dur(%q): [computed: %v] != [correct: 2024-01-15 09:30:00]", period, add)
		}
	}

	// brief units stay case-sensitive: "D" means days while "d" is invalid
	dur := NewDur(DurWithFrom("2024-01-15 08:00:00"), DurWithDur("1d"))
	if _, err := dur.Add(); err == nil {
		t.Error("expected an error for the invalid brief unit \"d\", got nil")
	}
	dur = NewDur(DurWithFrom("2024-01-15 08:00:00"), DurWithDur("1D"), DurWithOutputFormat("%F %T"))
	add, err := dur.Add()
	if err != nil {
		t.Fatal(err)
	}
	if len(add) != 1 || add[0] != "2024-01-16 08:00:00" {
		t.Errorf("[computed: %v] != [correct: 2024-01-16 08:00:00]", add)
	}
}

func TestSlashDateTimeSuffixVariants(t *testing.T) {
	// no t.Parallel: t.Setenv is process-wide
	// the "T" separator and lowercase am/pm are accepted alongside the
	// space-separated and uppercase forms
	t.Setenv(DateOrderEnvVar, "MDY")
	testFormat(t, "1/2/2024T08:30", "%F %H:%M", "2024-01-02 08:30")
	testFormat(t, "1/2/2024T08:30:45", "%F %T", "2024-01-02 08:30:45")
	testFormat(t, "1/2/24T08:30", "%F %H:%M", "2024-01-02 08:30")
	testFormat(t, "1/2/2024 3:04pm", "%F %H:%M", "2024-01-02 15:04")
	testFormat(t, "1/2/2024 3:04:05pm", "%F %T", "2024-01-02 15:04:05")
	testFormat(t, "1/2/2024 3:04PM", "%F %H:%M", "2024-01-02 15:04")
	testFormat(t, "1/2/2024 3:04 PM", "%F %H:%M", "2024-01-02 15:04")
	testFormat(t, "1/2/2024 3:04:05 pm", "%F %T", "2024-01-02 15:04:05")
}

func TestSpacedAmPmParsing(t *testing.T) {
	// a space before am/pm parses the same as the joined form, in bare
	// times and month-name dates alike (regression from the parsetime
	// removal: its US format allowed whitespace before the meridiem)
	testFormat(t, "Jan 2, 2024 3:04 PM", "%F %H:%M", "2024-01-02 15:04")
	testFormat(t, "January 2, 2024 3:04:05 pm", "%F %T", "2024-01-02 15:04:05")
	diff := NewDiff(DiffWithStart("11:00 AM"), DiffWithEnd("11:00PM"))
	result, _, err := diff.CalculateDiff()
	if err != nil {
		t.Fatal(err)
	}
	if result != "12 hours" {
		t.Errorf("[computed: %v] != [correct: 12 hours]", result)
	}
}
