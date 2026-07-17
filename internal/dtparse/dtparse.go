// Package dtparse is the unified fallback parser for date/time strings that
// the strict wall-clock, zoned, and slash-date layers of the root package do
// not claim. It replaces the parsetime and carbon fallbacks with a single
// ordered layout table tried top to bottom with the standard library, so
// ordering bugs between competing fallbacks are structurally impossible.
//
// Portions of the layout table are derived from the layout list in
// github.com/golang-module/carbon (MIT License, Copyright (c) gouguoyin),
// curated as described in DEPENDENCY_REPLACEMENT.md.
package dtparse

import (
	"fmt"
	"strings"
	"time"
)

// Kind classifies a layout's shape, which decides the post-parse behavior:
// wall-clock results are complete as is, time-only results need today's
// date stamped on, and zoned results need the caller's zone validation.
type Kind int

const (
	// KindWallClock is a zone-less date or date+time, parsed in the
	// caller-supplied location.
	KindWallClock Kind = iota
	// KindTimeOnly is a zone-less time of day; Parse stamps it with
	// today's date in the caller-supplied location.
	KindTimeOnly
	// KindZoned carries its own zone abbreviation or offset; validating
	// that time.Parse resolved the zone (rather than fabricating a
	// zero-offset placeholder) is the caller's job.
	KindZoned
)

type layoutEntry struct {
	layout string
	kind   Kind
}

// layouts is tried in order and the first success wins, so the order is part
// of the spec. Within each family, longer layouts come before their prefixes
// (e.g. "2006-1-2 15:4:5" before "2006-1-2") so a full match is never
// shadowed by a partial one; zone-less ANSIC precedes the zoned variant so a
// trailing year is never mistaken for a zone token; bare times come last
// because no date layout can match a leading 1-2 digit hour. Go's time.Parse
// accepts a fractional second after any seconds element even when the layout
// has none, so no ".999999999" variants are needed.
var layouts = []layoutEntry{
	// flexible zone-less dates and date/times: single-digit layout elements
	// accept both padded and unpadded input, covering "2024-1-2 8:30:45",
	// "2024-01" (year-month), and the dotted and 4-digit-year slashed
	// separators from carbon's list ("1/2/2006" forms are deliberately
	// absent: the root slash-date layer claims every d{1,2}/d{1,2}/d{2,4}
	// shape first, which keeps DTMATE_DATE_ORDER authoritative)
	{"2006-1-2T15:4:5", KindWallClock},
	{"2006-1-2T15:4", KindWallClock},
	{"2006-1-2 15:4:5", KindWallClock},
	{"2006-1-2 15:4", KindWallClock},
	{"2006-1-2", KindWallClock},
	{"2006-1", KindWallClock},
	{"2006.1.2 15:4:5", KindWallClock},
	{"2006.1.2 15:4", KindWallClock},
	{"2006.1.2", KindWallClock},
	{"2006.1", KindWallClock},
	{"2006/1/2 15:4:5", KindWallClock},
	{"2006/1/2 15:4", KindWallClock},
	{"2006/1/2", KindWallClock},
	{"2006/1", KindWallClock},
	// month-name dates; "January" and "Jan" need separate layouts because
	// the short element consumes only the first three letters of a full
	// month name and then fails on the remainder
	{"January 2, 2006 15:04:05", KindWallClock},
	{"January 2, 2006 15:04", KindWallClock},
	{"January 2, 2006", KindWallClock},
	{"Jan 2, 2006 15:04:05", KindWallClock},
	{"Jan 2, 2006 15:04", KindWallClock},
	{"Jan 2, 2006", KindWallClock},
	{"Mon, Jan 2, 2006 3:04 PM", KindWallClock},
	// day-first month-name dates; strftime's %v renders this shape
	{"2-Jan-2006 15:04:05", KindWallClock},
	{"2-Jan-2006 15:04", KindWallClock},
	{"2-Jan-2006", KindWallClock},
	// ANSIC with and without weekday, zone-less first
	{"Mon Jan _2 15:04:05 2006", KindWallClock},
	{"Jan 2 15:04:05 2006", KindWallClock},
	// ANSIC-style with zone and trailing year, without weekday: this is the
	// one month-name zoned shape zonedLayouts does not cover and is pinned
	// by TestExplicitZonePreservedAcrossDST
	{"Jan 2 15:04:05 MST 2006", KindZoned},
	// ISO date/times whose offset form RFC3339 rejects: no colon, or
	// whole hours only ("+08")
	{"2006-01-02T15:04:05-0700", KindZoned},
	{"2006-01-02T15:04:05Z07", KindZoned},
	// the tz CLI output format ("2006-01-02 15:04:05 -0700 MST") renders a
	// zone whose abbreviation is numeric (e.g. Asia/Kathmandu's "+0545") as
	// "+0545 +0545", which Go's "MST" element cannot re-parse; a second
	// offset element reads the same value, so the CLI's own output
	// round-trips (pinned by TestTimezoneOutputFormatRoundTrip)
	{"2006-01-02 15:04:05 -0700 -0700", KindZoned},
	// RFC822, RFC822Z, RFC850, Cookie, RFC1036
	{"02 Jan 06 15:04 MST", KindZoned},
	{"02 Jan 06 15:04 -0700", KindZoned},
	{"Monday, 02-Jan-06 15:04:05 MST", KindZoned},
	{"Monday, 02-Jan-2006 15:04:05 MST", KindZoned},
	{"Mon, 02 Jan 06 15:04:05 -0700", KindZoned},
	// bare times of day, stamped with today's date after parsing; Go's "PM"
	// element matches only uppercase and "pm" only lowercase, so both
	// spellings need a layout (same reason slashDateTimeSuffixes in the
	// root package carries both)
	{"15:04:05", KindTimeOnly},
	{"15:04", KindTimeOnly},
	{"3:04:05PM", KindTimeOnly},
	{"3:04PM", KindTimeOnly},
	{"3:04:05pm", KindTimeOnly},
	{"3:04pm", KindTimeOnly},
}

// outOfRange reports whether a time.Parse error means the input matched the
// layout's shape but a component value was invalid (e.g. a 61-minute time);
// such input is rejected immediately instead of falling through to layouts
// that might silently misread it.
func outOfRange(err error) bool {
	return err != nil && strings.Contains(err.Error(), "out of range")
}

// Parse tries the ordered layout table and returns the first successful
// parse along with the matched layout's Kind. loc is the location used for
// zone-less layouts (the root package passes time.Local); KindZoned results
// are parsed with time.Parse and the caller must validate the zone.
func Parse(source string, loc *time.Location) (time.Time, Kind, error) {
	for _, entry := range layouts {
		var t time.Time
		var err error
		if entry.kind == KindZoned {
			t, err = time.Parse(entry.layout, source)
		} else {
			t, err = time.ParseInLocation(entry.layout, source, loc)
		}
		if err == nil {
			if entry.kind == KindTimeOnly {
				now := time.Now().In(loc)
				t = time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
			}
			return t, entry.kind, nil
		}
		if outOfRange(err) {
			return time.Time{}, entry.kind, fmt.Errorf("invalid date/time %q: %v", source, err)
		}
	}
	return time.Time{}, KindWallClock, fmt.Errorf("unable to parse date/time: %q", source)
}
