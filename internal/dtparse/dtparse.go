// Package dtparse is the unified fallback parser for date/time strings that
// the strict wall-clock, zoned, and slash-date layers of the root package do
// not claim. It replaces the parsetime and carbon fallbacks with a single
// ordered layout table tried top to bottom with the standard library, so
// ordering bugs between competing fallbacks are structurally impossible.
//
// Portions of the layout table are derived from the layout list in
// github.com/golang-module/carbon (MIT License, Copyright (c) gouguoyin),
// curated down to the shapes this repo documents as supported; the
// curation rationale is recorded in the comments on each table section.
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

// timeOfDaySuffixes are the time-of-day forms accepted after every entry in
// dateBases, ordered longest first so an out-of-range component is reported
// against the most specific shape; am/pm needs four spellings per shape
// because Go's "PM" element matches only uppercase and "pm" only lowercase,
// and the meridiem is accepted both joined to the time and separated by a
// space (the same coverage slashDateTimeSuffixes provides in the root
// package); the empty suffix accepts the bare date. Single-digit layout
// elements accept both padded and unpadded input, so "8:30" and "08:30"
// both match.
var timeOfDaySuffixes = []string{
	" 15:4:5", " 15:4",
	" 3:4:5PM", " 3:4:5pm", " 3:4:5 PM", " 3:4:5 pm",
	" 3:4PM", " 3:4pm", " 3:4 PM", " 3:4 pm",
	"",
}

// dateBases are the zone-less date shapes, each combined with every entry
// of timeOfDaySuffixes: the dashed, dotted, and 4-digit-year slashed
// separators from carbon's list ("1/2/2006" forms are deliberately absent:
// the root slash-date layer claims every d{1,2}/d{1,2}/d{2,4} shape first,
// which keeps DTMATE_DATE_ORDER authoritative), then the month-name shapes;
// "January" and "Jan" need separate bases because the short element
// consumes only the first three letters of a full month name and then
// fails on the remainder; "2-Jan-2006" is the shape strftime's %v renders
var dateBases = []string{
	"2006-1-2",
	"2006.1.2",
	"2006/1/2",
	"January 2, 2006",
	"Jan 2, 2006",
	"Mon, Jan 2, 2006",
	"2-Jan-2006",
}

// layouts is tried in order and the first success wins, so the order is part
// of the spec: within each family, longer layouts come before their prefixes
// (e.g. "2006-1-2 15:4:5" before "2006-1-2") so an out-of-range component is
// rejected against the most specific shape; zone-less ANSIC precedes the
// zoned variant so a trailing year is never mistaken for a zone token; bare
// times come last because no date layout can match a leading 1-2 digit hour.
// Go's time.Parse accepts a fractional second after any seconds element even
// when the layout has none, so no ".999999999" variants are needed.
var layouts = buildLayouts()

// buildLayouts assembles the ordered layout table from dateBases and
// timeOfDaySuffixes plus the standalone year-month, ANSIC, zoned, and
// bare-time entries
func buildLayouts() []layoutEntry {
	var table []layoutEntry
	// dashed date/times with the "T" separator
	table = append(table,
		layoutEntry{"2006-1-2T15:4:5", KindWallClock},
		layoutEntry{"2006-1-2T15:4", KindWallClock})
	// every date base with every time-of-day suffix, e.g. "2024-1-2 8:30:45",
	// "Jan 2, 2024 3:04PM", "Mon, Jan 2, 2024 15:04"
	for _, base := range dateBases {
		for _, suffix := range timeOfDaySuffixes {
			table = append(table, layoutEntry{base + suffix, KindWallClock})
		}
	}
	// year-month shapes, e.g. "2024-01"
	table = append(table,
		layoutEntry{"2006-1", KindWallClock},
		layoutEntry{"2006.1", KindWallClock},
		layoutEntry{"2006/1", KindWallClock})
	table = append(table,
		// ANSIC with and without weekday, zone-less first
		layoutEntry{"Mon Jan _2 15:04:05 2006", KindWallClock},
		layoutEntry{"Jan 2 15:04:05 2006", KindWallClock},
		// ANSIC-style with zone and trailing year, without weekday: this is
		// the one month-name zoned shape zonedLayouts does not cover and is
		// pinned by TestExplicitZonePreservedAcrossDST
		layoutEntry{"Jan 2 15:04:05 MST 2006", KindZoned},
		// ISO date/times whose offset form RFC3339 rejects: no colon, or
		// whole hours only ("+08")
		layoutEntry{"2006-01-02T15:04:05-0700", KindZoned},
		layoutEntry{"2006-01-02T15:04:05Z07", KindZoned},
		// the tz CLI output format ("2006-01-02 15:04:05 -0700 MST") renders
		// a zone whose abbreviation is numeric (e.g. Asia/Kathmandu's
		// "+0545") as "+0545 +0545", which Go's "MST" element cannot
		// re-parse; a second offset element reads the same value, so the
		// CLI's own output round-trips (pinned by
		// TestTimezoneOutputFormatRoundTrip)
		layoutEntry{"2006-01-02 15:04:05 -0700 -0700", KindZoned},
		// RFC822, RFC822Z, RFC850, Cookie, RFC1036
		layoutEntry{"02 Jan 06 15:04 MST", KindZoned},
		layoutEntry{"02 Jan 06 15:04 -0700", KindZoned},
		layoutEntry{"Monday, 02-Jan-06 15:04:05 MST", KindZoned},
		layoutEntry{"Monday, 02-Jan-2006 15:04:05 MST", KindZoned},
		layoutEntry{"Mon, 02 Jan 06 15:04:05 -0700", KindZoned})
	// bare times of day, stamped with today's date after parsing: the same
	// time-of-day shapes accepted after a date, minus the leading space
	for _, suffix := range timeOfDaySuffixes {
		if suffix == "" {
			continue
		}
		table = append(table, layoutEntry{strings.TrimPrefix(suffix, " "), KindTimeOnly})
	}
	return table
}

// outOfRange reports whether a time.Parse error means the input matched the
// layout's shape but a component value was invalid (e.g. a 61-minute time);
// such input is rejected immediately instead of falling through to layouts
// that might silently misread it. The exception is an out-of-range hour
// from a meridiem (12-hour) layout: hours 13-23 are valid in the 24-hour
// layouts, so a 12-hour element reading one (e.g. the "17" in
// "2024-01-15 17:45:00 +0545 +0545", or the "20" of a year) only means the
// meridiem layout does not apply, and later layouts must still be tried.
func outOfRange(layout string, err error) bool {
	if err == nil || !strings.Contains(err.Error(), "out of range") {
		return false
	}
	meridiem := strings.Contains(layout, "PM") || strings.Contains(layout, "pm")
	return !(meridiem && strings.Contains(err.Error(), "hour out of range"))
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
		if outOfRange(entry.layout, err) {
			return time.Time{}, entry.kind, fmt.Errorf("invalid date/time %q: %v", source, err)
		}
	}
	return time.Time{}, KindWallClock, fmt.Errorf("unable to parse date/time: %q", source)
}
