package dtparse

import (
	"strings"
	"testing"
	"time"
)

// testParseWallClock asserts that source parses as a KindWallClock result
// equal to want interpreted in UTC (the test location).
func testParseWallClock(t *testing.T, source, want string) {
	t.Helper()
	parsed, kind, err := Parse(source, time.UTC)
	if err != nil {
		t.Fatalf("Parse(%q): %v", source, err)
	}
	if kind != KindWallClock {
		t.Errorf("Parse(%q): kind = %v, want KindWallClock", source, kind)
	}
	got := parsed.Format("2006-01-02 15:04:05.999999999")
	if got != want {
		t.Errorf("Parse(%q) = %q, want %q", source, got, want)
	}
}

func TestParseFlexibleDates(t *testing.T) {
	t.Parallel()
	testParseWallClock(t, "2024-1-2", "2024-01-02 00:00:00")
	testParseWallClock(t, "2024-1-2 8:30", "2024-01-02 08:30:00")
	testParseWallClock(t, "2024-1-2 8:30:45", "2024-01-02 08:30:45")
	testParseWallClock(t, "2024-1-2T8:30:45", "2024-01-02 08:30:45")
	testParseWallClock(t, "2024-01-02T08:30", "2024-01-02 08:30:00")
	testParseWallClock(t, "2024-01", "2024-01-01 00:00:00")
	testParseWallClock(t, "2024-1", "2024-01-01 00:00:00")
	testParseWallClock(t, "2024.01.15", "2024-01-15 00:00:00")
	testParseWallClock(t, "2024.1.15 08:30:45", "2024-01-15 08:30:45")
	testParseWallClock(t, "2024/01/15", "2024-01-15 00:00:00")
	testParseWallClock(t, "2024/1/15 08:30", "2024-01-15 08:30:00")
	testParseWallClock(t, "2024/01", "2024-01-01 00:00:00")
}

func TestParseMonthNameDates(t *testing.T) {
	t.Parallel()
	testParseWallClock(t, "Jan 2, 2024", "2024-01-02 00:00:00")
	testParseWallClock(t, "Jan 2, 2024 08:30", "2024-01-02 08:30:00")
	testParseWallClock(t, "Jan 2, 2024 08:30:45", "2024-01-02 08:30:45")
	testParseWallClock(t, "January 2, 2024", "2024-01-02 00:00:00")
	testParseWallClock(t, "January 2, 2024 08:30:00", "2024-01-02 08:30:00")
	testParseWallClock(t, "Tue, Jan 2, 2024 3:04 PM", "2024-01-02 15:04:00")
	testParseWallClock(t, "2-Jan-2024", "2024-01-02 00:00:00")
	testParseWallClock(t, "22-Jul-2024 08:21:44", "2024-07-22 08:21:44")
	testParseWallClock(t, "22-Jul-2024 08:21", "2024-07-22 08:21:00")
}

func TestParseAnsic(t *testing.T) {
	t.Parallel()
	testParseWallClock(t, "Mon Jan  2 15:04:05 2006", "2006-01-02 15:04:05")
	testParseWallClock(t, "Jan 2 15:04:05 2024", "2024-01-02 15:04:05")
	testParseWallClock(t, "Jan 15 12:00:00 2026", "2026-01-15 12:00:00")
}

func TestParseFractionalSeconds(t *testing.T) {
	t.Parallel()
	// Go accepts a fraction after seconds without a dedicated layout
	testParseWallClock(t, "2024-1-2 8:30:45.123456789", "2024-01-02 08:30:45.123456789")
	testParseWallClock(t, "Jan 2, 2024 08:30:45.5", "2024-01-02 08:30:45.5")
}

func TestParseTimeOnlyStamping(t *testing.T) {
	t.Parallel()
	loc := time.UTC
	for source, want := range map[string]string{
		"08:30":         "08:30:00",
		"15:16:15":      "15:16:15",
		"11:00AM":       "11:00:00",
		"11:00PM":       "23:00:00",
		"3:04pm":        "15:04:00",
		"3:04:05PM":     "15:04:05",
		"12:34:56.1234": "12:34:56.1234",
	} {
		parsed, kind, err := Parse(source, loc)
		if err != nil {
			t.Fatalf("Parse(%q): %v", source, err)
		}
		if kind != KindTimeOnly {
			t.Errorf("Parse(%q): kind = %v, want KindTimeOnly", source, kind)
		}
		now := time.Now().In(loc)
		wantDate := now.Format("2006-01-02")
		got := parsed.Format("2006-01-02 15:04:05.9999")
		want = wantDate + " " + want
		// tolerate a midnight rollover between the two Now() calls
		if got != want && !strings.HasSuffix(got, strings.TrimPrefix(want, wantDate)) {
			t.Errorf("Parse(%q) = %q, want %q", source, got, want)
		}
	}
}

func TestParseZonedKinds(t *testing.T) {
	t.Parallel()
	// each zoned layout must classify as KindZoned so the caller runs zone
	// validation; offsets must be preserved in the result
	tests := []struct {
		source     string
		wantOffset int
	}{
		{"2024-01-15T08:30:00-0500", -5 * 3600},
		{"2024-01-15T08:30:00+08", 8 * 3600},
		{"02 Jan 24 15:04 -0700", -7 * 3600},
		{"Mon, 02 Jan 24 15:04:05 -0700", -7 * 3600},
		// the tz CLI output format for a numeric zone abbreviation
		{"2024-01-15 17:45:00 +0545 +0545", 5*3600 + 45*60},
	}
	for _, tt := range tests {
		parsed, kind, err := Parse(tt.source, time.UTC)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tt.source, err)
		}
		if kind != KindZoned {
			t.Errorf("Parse(%q): kind = %v, want KindZoned", tt.source, kind)
		}
		if _, offset := parsed.Zone(); offset != tt.wantOffset {
			t.Errorf("Parse(%q): offset = %d, want %d", tt.source, offset, tt.wantOffset)
		}
	}
	// abbreviation-carrying zoned layouts classify as KindZoned even when
	// the abbreviation is unresolvable (validation is the caller's job)
	for _, source := range []string{
		"Jan 15 12:00:00 EDT 2026",
		"02 Jan 24 15:04 EST",
		"Monday, 02-Jan-24 15:04:05 GMT",
		"Monday, 02-Jan-2024 15:04:05 GMT",
	} {
		_, kind, err := Parse(source, time.UTC)
		if err != nil {
			t.Fatalf("Parse(%q): %v", source, err)
		}
		if kind != KindZoned {
			t.Errorf("Parse(%q): kind = %v, want KindZoned", source, kind)
		}
	}
}

func TestParseOutOfRangeRejected(t *testing.T) {
	t.Parallel()
	// a layout-shaped input with an invalid component errors immediately
	// instead of trying later layouts
	for _, source := range []string{"2024-13-01", "2024.02.30", "2024-1-32", "25:00", "08:61", "13:04PM", "Jan 32, 2024"} {
		if _, _, err := Parse(source, time.UTC); err == nil {
			t.Errorf("Parse(%q): expected an out-of-range error, got nil", source)
		}
	}
}

func TestParseUnparseableRejected(t *testing.T) {
	t.Parallel()
	for _, source := range []string{"not a date", "hello world", "1/2", ""} {
		if _, _, err := Parse(source, time.UTC); err == nil {
			t.Errorf("Parse(%q): expected an error, got nil", source)
		}
	}
}

func TestParsePre1970(t *testing.T) {
	t.Parallel()
	// no layout has a 1970 floor
	testParseWallClock(t, "Jan 2, 1950", "1950-01-02 00:00:00")
	testParseWallClock(t, "1950-1-2 12:00:00", "1950-01-02 12:00:00")
}

func TestParseTwoDigitYearPivot(t *testing.T) {
	t.Parallel()
	// Go's "06" element pivots 69-99 to 19xx and 00-68 to 20xx
	parsed, _, err := Parse("02 Jan 76 15:04 -0500", time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Year() != 1976 {
		t.Errorf("expected 1976, got %d", parsed.Year())
	}
	parsed, _, err = Parse("02 Jan 24 15:04 -0500", time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Year() != 2024 {
		t.Errorf("expected 2024, got %d", parsed.Year())
	}
}
