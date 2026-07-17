// library_fix_test.go pins the behaviors introduced by the v1.19.0 library
// bug-hunt batch: uniform am/pm acceptance across every zone-less date
// shape, out-of-range errors that no longer misfire from 12-hour layouts,
// bare-time sources with a zone suffix stamped with that zone's current
// date, whole-hour ±HH wall-clock offsets that conflict with a trailing
// zone token, and duration errors that quote the caller's original input.
// It uses only the standard testing package, per current repo policy.
package DateTimeMate

import (
	"strings"
	"testing"
	"time"
)

func TestMeridiemAcceptedAcrossDateShapes(t *testing.T) {
	// the meridiem is accepted joined to the time and space-separated, in
	// both spellings, after every zone-less date shape - the same coverage
	// slash dates and bare times have always had ("Jan 2, 2024 3:04PM"
	// used to be rejected while "Jan 2, 2024 3:04 PM" parsed)
	testFormat(t, "Jan 2, 2024 3:04PM", "%F %H:%M", "2024-01-02 15:04")
	testFormat(t, "Jan 2, 2024 3:04pm", "%F %H:%M", "2024-01-02 15:04")
	testFormat(t, "January 2, 2024 3:04:05PM", "%F %T", "2024-01-02 15:04:05")
	testFormat(t, "2-Jan-2024 3:04 PM", "%F %H:%M", "2024-01-02 15:04")
	testFormat(t, "2024-01-02 3:04 PM", "%F %H:%M", "2024-01-02 15:04")
	testFormat(t, "2024-1-2 3:04PM", "%F %H:%M", "2024-01-02 15:04")
	testFormat(t, "2024.01.15 3:04 pm", "%F %H:%M", "2024-01-15 15:04")

	// the weekday-comma shape accepts the same time-of-day forms as every
	// other date shape, not just "3:04 PM" (its only pre-fix layout)
	testFormat(t, "Mon, Jan 2, 2024 3:04 pm", "%F %H:%M", "2024-01-02 15:04")
	testFormat(t, "Mon, Jan 2, 2024 15:04", "%F %H:%M", "2024-01-02 15:04")
	testFormat(t, "Mon, Jan 2, 2024", "%F", "2024-01-02")
}

func TestOutOfRangeHourNoLongerMisfires(t *testing.T) {
	// a 12-hour layout reading a 24-hour value (or the "20" of a year) used
	// to abort the whole parse with "hour out of range" even though every
	// component of the input was valid; such inputs now parse
	testFormat(t, "2024-01-02 3:04 PM", "%F %H:%M", "2024-01-02 15:04")

	// genuinely invalid components still fail fast with the real reason
	for source, want := range map[string]string{
		"Feb 30, 2024 3:04PM": "day out of range",
		"2024-02-30":          "day out of range",
		"08:61":               "minute out of range",
	} {
		_, err := Reformat(source, "%F")
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("Reformat(%q): expected an error containing %q, got: %v", source, want, err)
		}
	}

	// an hour that is invalid for a meridiem still errors, just without a
	// misleading out-of-range message
	for _, source := range []string{"13:04PM", "13:04 PM"} {
		if _, err := Reformat(source, "%T"); err == nil {
			t.Errorf("Reformat(%q): expected an error, got nil", source)
		}
	}
}

func TestTimeOnlySourceUsesZoneCurrentDate(t *testing.T) {
	// a bare time with a zone suffix means that time on the zone's current
	// day: the date used to come from the machine's zone, so "08:30 CET"
	// requested from a machine near the date line could land on the wrong
	// CET day
	conv := setupConverter()
	for _, zone := range []string{"Etc/GMT-14", "Etc/GMT+12"} {
		loc, err := time.LoadLocation(zone)
		if err != nil {
			t.Fatalf("LoadLocation(%s): %v", zone, err)
		}
		for attempt := 0; ; attempt++ {
			before := time.Now().In(loc).Format("2006-01-02")
			result, err := conv.ConvertTimeZone("08:30 "+zone, zone)
			if err != nil {
				t.Fatalf("ConvertTimeZone(08:30 %s): %v", zone, err)
			}
			after := time.Now().In(loc).Format("2006-01-02")
			if before != after && attempt == 0 {
				continue // midnight in the zone rolled over mid-test; retry once
			}
			if got := result.Format("2006-01-02 15:04:05"); got != before+" 08:30:00" {
				t.Errorf("08:30 %s = %q, want %q", zone, got, before+" 08:30:00")
			}
			break
		}
	}
}

func TestWholeHourOffsetConflictsWithZone(t *testing.T) {
	// TZ-10 pinned the 4-digit form: a wall clock carrying +0900 with a
	// trailing EST is contradictory; the ISO8601 short form "+05" used to
	// slip past that check and the wall clock was silently reinterpreted
	// in the trailing zone
	conv := setupConverter()
	_, err := conv.ConvertTimeZone("2024-01-02T15:04:05+05 EST", "UTC")
	if err == nil || !strings.Contains(err.Error(), "source names both") {
		t.Errorf("+05 with EST: expected a contradictory-zone error, got: %v", err)
	}

	// when the short offset and the trailing zone agree, the instant is
	// honored exactly
	result, err := conv.ConvertTimeZone("2024-01-02T15:04:05-05 EST", "UTC")
	if err != nil {
		t.Fatalf("-05 with EST: %v", err)
	}
	want := time.Date(2024, 1, 2, 20, 4, 5, 0, time.UTC).Unix()
	if result.Unix() != want {
		t.Errorf("-05 with EST: got unix %d, want %d", result.Unix(), want)
	}
}

func TestInvalidPeriodErrorQuotesOriginalInput(t *testing.T) {
	// the brief-period expansion mangles unrecognized text with placeholder
	// replacements, and the mangled form used to leak into the error
	// ("1 month" errored with `parsing "ont": invalid syntax`)
	dur := NewDur(DurWithFrom("2024-01-15"), DurWithDur("1 month"))
	_, err := dur.Add()
	if err == nil {
		t.Fatal("expected an error for the unsupported unit \"month\", got nil")
	}
	if !strings.Contains(err.Error(), "1 month") || strings.Contains(err.Error(), "ont\"") {
		t.Errorf("error must quote the original period, not the mangled expansion: %v", err)
	}
}
