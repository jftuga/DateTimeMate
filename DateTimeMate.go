package DateTimeMate

import (
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang-module/carbon/v2"
	"github.com/jftuga/parsetime"
	"github.com/lestrrat-go/strftime"
)

// ReadmeMd is the project README, embedded here (next to the file at the
// module root) because go:embed cannot reference parent directories from
// the cmd package; the CLI extracts its -e examples from it.
//
//go:embed README.md
var ReadmeMd string

const (
	ModName    string = "DateTimeMate"
	ModVersion string = "1.16.0"
	ModUrl     string = "https://github.com/jftuga/DateTimeMate"
)

// DateOrderEnvVar names the environment variable that controls how an
// ambiguous slash-separated date such as "01/02/2024" is interpreted:
// "MDY" (month/day/year, the default) or "DMY" (day/month/year).
// Unambiguous dates such as "25/12/2024" parse the same way regardless of
// this variable. When an ambiguous date is parsed and the variable is not
// set, a warning naming the variable is written to stderr.
const DateOrderEnvVar = "DTMATE_DATE_ORDER"

// this type of data structure is needed to preserve order as
// replacement always needs to be performed in this order
// otherwise you may get results such as:
// 1h12m1s123millis456us788ns (millis instead of ms)
// 123micros (micros instead of us)
var abbrevMap = [][]string{
	{"nanoseconds", "ns"},
	{"microseconds", "us"},
	{"milliseconds", "ms"},
	{"seconds", "s"},
	{"minutes", "m"},
	{"hours", "h"},
	{"days", "D"},
	{"weeks", "W"},
	{"years", "Y"},
	{"nanosecond", "ns"},
	{"microsecond", "us"},
	{"millisecond", "ms"},
	{"second", "s"},
	{"minute", "m"},
	{"hour", "h"},
	{"day", "D"},
	{"week", "W"},
	{"year", "Y"},
}

// ConvertRelativeDateToActual converts "yesterday", "today", "tomorrow"
// into actual dates; yesterday and tomorrow are -/+ 24 hours of current time
func ConvertRelativeDateToActual(from string) string {
	switch strings.ToLower(from) {
	case "now":
		return time.Now().Format("2006-01-02 15:04:05")
	case "today":
		return time.Now().Format("2006-01-02 15:04:05")
	case "yesterday":
		// Add keeps the documented "exactly 24 hours" promise; AddDate
		// is calendar-day arithmetic, which is 23 or 25 real hours on the
		// two DST transition days
		return time.Now().Add(-24 * time.Hour).Format("2006-01-02 15:04:05")
	case "tomorrow":
		return time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04:05")
	}
	return from
}

// shrinkPeriod convert a period into a brief period
// only allow one replacement per each period
// Ex: 1 hour 2 minutes 3 seconds => 1h2m3s
func shrinkPeriod(period string) string {
	for _, tuple := range abbrevMap {
		period = strings.Replace(period, tuple[0], tuple[1], 1)
	}

	return strings.ReplaceAll(period, " ", "")
}

// removeTrailingS convert plural to singular, such as "hours" to "hour"
func removeTrailingS(s string) string {
	return strings.TrimSuffix(s, "s")
}

// fixLocalZone corrects the offset of zone-less date/times: parsetime stamps
// them with a fixed snapshot of today's zone (e.g. EDT on a January date), so
// reinterpret the wall clock in time.Local, which resolves the DST offset in
// effect on that date; times that carry any other explicit zone are untouched
func fixLocalZone(t time.Time) time.Time {
	name, offset := t.Zone()
	nowName, nowOffset := time.Now().Zone()
	if name == nowName && offset == nowOffset {
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
	}
	return t
}

// zoneOffsetRegexp matches a numeric UTC offset written into a date/time
// string, such as "+0500", "-04:30", or "-0400": a sign followed by two
// digits, an optional colon, and two more digits; the separators inside
// dates like "2024-10-15" never have four digits after the sign, so plain
// dates cannot match
var zoneOffsetRegexp = regexp.MustCompile(`[+-][0-9]{2}:?[0-9]{2}`)

// sourceHasExplicitZone reports whether the source text itself names the
// zone found on the parsed time, either by abbreviation (e.g. "EDT") or as
// a numeric UTC offset; fixLocalZone must leave such times alone because
// their zone was written by the user, not stamped on by parsetime
func sourceHasExplicitZone(source string, t time.Time) bool {
	name, offset := t.Zone()
	if name != "" && strings.Contains(source, name) {
		return true
	}
	// an RFC3339 trailing "Z" (e.g. "2024-01-15T08:30:00Z") denotes UTC
	// explicitly even though the zone name "UTC" never appears in the text
	if offset == 0 && strings.HasSuffix(source, "Z") {
		return true
	}
	return zoneOffsetRegexp.MatchString(source)
}

// wallClockLayouts are zone-less layouts interpreted in the local time
// zone; they are tried before parsetime because parsetime silently
// corrupts pre-1970 date/times; the 14-digit compact layout is included
// so it parses deterministically instead of relying on parsetime, whose
// year-mismatch guard cannot see a year inside a 14-digit run
var wallClockLayouts = []string{"2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02 15:04", "2006-01-02", "20060102150405"}

// zonedLayouts carry their own zone or offset, which must be preserved in
// the result (e.g. reformatting "...Z" with %Z must print UTC, not local)
var zonedLayouts = []string{time.RFC3339Nano, "2006-01-02 15:04:05 -0700 MST", "2006-01-02 15:04:05 -0700", "2006-01-02 15:04:05 MST", time.UnixDate, time.RFC1123Z, time.RFC1123, time.RubyDate}

// zeroOffsetZoneNames holds the zone names that legitimately denote UTC+0
// in a parsed date/time: the empty name of a bare numeric offset plus the
// zero-offset abbreviations from the zone table; time.Parse fabricates a
// zero-offset zone for any abbreviation not defined in the local time
// zone, so a zero offset under any other name signals a fabricated,
// silently-wrong parse
var zeroOffsetZoneNames = buildZeroOffsetZoneNames()

// buildZeroOffsetZoneNames collects the legitimate zero-offset zone names
// from LoadZoneDefinitions
func buildZeroOffsetZoneNames() map[string]bool {
	names := map[string]bool{"": true}
	for abbrev, def := range LoadZoneDefinitions() {
		if def.Offset == 0 {
			names[abbrev] = true
		}
	}
	return names
}

// isAllDigits reports whether s is non-empty and contains only ASCII digits
func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// slashDateFields splits the leading token of source into the two 1-2 digit
// fields and the 2 or 4 digit year of a slash-separated date such as
// "01/02/2024", "1/2/24", or "1/2/2024 08:30"; ok is false when the token
// does not have that shape; two-digit years must be claimed here, otherwise
// they fall through to parsers that ignore DateOrderEnvVar and can even
// replace the year with the current one ("12/31/24" used to parse as
// December 31 of the current year)
func slashDateFields(source string) (first, second, yearDigits int, ok bool) {
	token := source
	if i := strings.IndexAny(token, " T"); i != -1 {
		token = token[:i]
	}
	parts := strings.Split(token, "/")
	if len(parts) != 3 {
		return 0, 0, 0, false
	}
	if len(parts[0]) > 2 || len(parts[1]) > 2 || (len(parts[2]) != 4 && len(parts[2]) != 2) {
		return 0, 0, 0, false
	}
	for _, part := range parts {
		if !isAllDigits(part) {
			return 0, 0, 0, false
		}
	}
	first, _ = strconv.Atoi(parts[0])
	second, _ = strconv.Atoi(parts[1])
	return first, second, len(parts[2]), true
}

// slashDateTimeSuffixes are the time-of-day layouts accepted after a
// slash-separated date; the empty suffix accepts a bare date, and the
// space and "T" separated forms both accept times with and without seconds
var slashDateTimeSuffixes = []string{"", " 15:04:05", "T15:04:05", " 15:04", "T15:04", " 3:04:05PM", " 3:04PM", " 3:04:05pm", " 3:04pm"}

// parseSlashDate parses slash-separated dates, whose field order the layered
// parsers disagree on: month first (the default) or day first when
// DateOrderEnvVar is set to "DMY". A field greater than 12 disambiguates on
// its own, regardless of the variable; an ambiguous date parsed with the
// variable unset triggers a stderr warning naming it. The claimed return
// reports whether the input has the slash-date shape at all; when true the
// caller must not try other parsers, so that no shape can silently fall
// through to a parser with a different field order.
func parseSlashDate(source string) (time.Time, bool, error) {
	first, second, yearDigits, ok := slashDateFields(source)
	if !ok {
		return time.Time{}, false, nil
	}
	var monthFirst bool
	switch {
	case first > 12 && second > 12:
		return time.Time{}, true, fmt.Errorf("invalid date %q: neither %d nor %d can be a month", source, first, second)
	case first > 12:
		monthFirst = false
	case second > 12:
		monthFirst = true
	default:
		order := strings.ToUpper(strings.TrimSpace(os.Getenv(DateOrderEnvVar)))
		switch order {
		case "", "MDY":
			if order == "" {
				fmt.Fprintf(os.Stderr, "warning: %q is ambiguous: interpreting as month/day/year; set %s=DMY to override\n", source, DateOrderEnvVar)
			}
			monthFirst = true
		case "DMY":
			monthFirst = false
		default:
			return time.Time{}, true, fmt.Errorf("%s must be MDY or DMY, not %q", DateOrderEnvVar, order)
		}
	}
	yearLayout := "2006"
	if yearDigits == 2 {
		yearLayout = "06"
	}
	dateLayout := "1/2/" + yearLayout
	if !monthFirst {
		dateLayout = "2/1/" + yearLayout
	}
	for _, suffix := range slashDateTimeSuffixes {
		t, err := time.ParseInLocation(dateLayout+suffix, source, time.Local)
		if err == nil {
			return t, true, nil
		}
		if outOfRangeParseError(err) {
			return time.Time{}, true, fmt.Errorf("invalid date/time %q: %v", source, err)
		}
	}
	return time.Time{}, true, fmt.Errorf("unable to parse date/time: %q", source)
}

// outOfRangeParseError reports whether a time.Parse error means the input
// matched the layout's shape but a component value was invalid (e.g.
// "2024-02-30": day out of range); such input must be rejected immediately
// rather than falling through to lenient parsers that silently normalize
// or even mangle it ("08:61:00" used to come back as "08:06:01")
func outOfRangeParseError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "out of range")
}

// parseDateTime parses a date/time string: standard layouts are tried
// first because they preserve explicit zones and handle any year, then
// slash-separated dates (whose field order is settled by DateOrderEnvVar),
// then parsetime (with fixLocalZone correcting the DST offset of zone-less
// strings), and finally carbon; parsetime results naming a year that does
// not appear in the input are rejected as corrupt rather than returned
func parseDateTime(source string) (time.Time, error) {
	for _, layout := range wallClockLayouts {
		t, err := time.ParseInLocation(layout, source, time.Local)
		if err == nil {
			return t, nil
		}
		if outOfRangeParseError(err) {
			return time.Time{}, fmt.Errorf("invalid date/time %q: %v", source, err)
		}
	}
	for _, layout := range zonedLayouts {
		t, err := time.Parse(layout, source)
		if outOfRangeParseError(err) {
			return time.Time{}, fmt.Errorf("invalid date/time %q: %v", source, err)
		}
		if err == nil {
			name, offset := t.Zone()
			if offset != 0 || zeroOffsetZoneNames[strings.ToUpper(name)] {
				return t, nil
			}
			// time.Parse fabricates a zero offset for any zone token it
			// cannot resolve; ±NN tokens (e.g. "+08") carry their real
			// offset in the digits, so re-stamp the wall clock into the
			// zone they actually name instead of keeping the wrong instant
			if loc, shaped, serr := parseOffsetSuffix(name); shaped {
				if serr != nil {
					return time.Time{}, serr
				}
				return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc), nil
			}
			return time.Time{}, fmt.Errorf("zone abbreviation %q is not resolvable in this input position", name)
		}
	}
	if t, claimed, err := parseSlashDate(source); claimed {
		return t, err
	}
	p, err := parsetime.NewParseTime()
	if err != nil {
		return time.Time{}, err
	}
	t, err := p.Parse(source)
	if err == nil && !parsedYearMismatch(source, t) {
		if sourceHasExplicitZone(source, t) {
			return t, nil
		}
		return fixLocalZone(t), nil
	}
	if c := carbon.Parse(source); c.Error == nil {
		return c.StdTime(), nil
	}
	if err == nil {
		return time.Time{}, fmt.Errorf("refusing unreliable parse of %q: result year %d does not appear in the input", source, t.Year())
	}
	return time.Time{}, err
}

// parsedYearMismatch reports whether the input names a year (a standalone
// 4-digit run between 1000 and 2999, not preceded by a '.' as fractional
// seconds would be) that the parsed result does not match; parsetime
// signals corruption this way instead of returning an error
func parsedYearMismatch(source string, t time.Time) bool {
	sawYear := false
	digits := 0
	for i := 0; i <= len(source); i++ {
		if i < len(source) && source[i] >= '0' && source[i] <= '9' {
			digits++
			continue
		}
		if digits == 4 {
			start := i - digits
			year, _ := strconv.Atoi(source[start:i])
			if year >= 1000 && year <= 2999 && (start == 0 || source[start-1] != '.') {
				if year == t.Year() {
					return false
				}
				sawYear = true
			}
		}
		digits = 0
	}
	return sawYear
}

// Reformat converts a date/time string into a specified format. The source can be:
//   - A Unix timestamp (e.g., "1700265600")
//   - A relative date (e.g., "yesterday", "now")
//   - Any other parseable date format
//
// The outputFormat parameter uses strftime format specifiers, with additional
// support for Unix seconds via '%s'.
//
// Example usage:
//
//	s, err := Reformat("1700265600", "%Y-%m-%d")         // Unix timestamp to date
//	s, err := Reformat("yesterday", "%Y-%m-%d %H:%M:%S") // Relative date to datetime
//	s, err := Reformat("2024-01-01", "%s")               // Date to Unix timestamp
//
// Returns an error if:
//   - The outputFormat is invalid
//   - The source date cannot be parsed
//   - The time parser initialization fails
func Reformat(source string, outputFormat string) (string, error) {
	source = strings.TrimSpace(source)

	// creates a new Strftime instance
	// outputFormat is a pattern string that follows strftime formatting
	// the additional formatting behavior allows this to also use the unix time %s modifier
	f, err := strftime.New(outputFormat, strftime.WithUnixSeconds('s'))
	if err != nil {
		return "", err
	}

	t, err := parseDateTimeOrUnix(source)
	if err != nil {
		return "", err
	}
	return f.FormatString(t), nil
}

// timestampDigits returns the number of digits in a pure integer string,
// ignoring any leading sign
func timestampDigits(s string) int {
	s = strings.TrimPrefix(s, "+")
	s = strings.TrimPrefix(s, "-")
	return len(s)
}

// unixStringToTime converts a string containing a Unix timestamp to time.Time.
// It accepts timestamps in seconds (up to 10 digits) and milliseconds (13 digits).
// Returns the corresponding time.Time and any error encountered during conversion.
//
// If the input string is not a valid integer, is empty, is negative, or has an
// ambiguous digit count (11, 12, or more than 13 digits), it returns a zero
// time.Time and an error.
func unixStringToTime(timestamp string) (time.Time, error) {
	timestamp = strings.TrimSpace(timestamp)
	unixTime, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	if unixTime < 0 {
		return time.Time{}, fmt.Errorf("timestamps can't be negative: %v", timestamp)
	}

	switch digits := timestampDigits(timestamp); {
	case digits <= 10:
		return time.Unix(unixTime, 0), nil
	case digits == 13:
		return time.UnixMilli(unixTime), nil
	default:
		return time.Time{}, fmt.Errorf("ambiguous timestamp length %d for %q: expected up to 10 digits (seconds) or exactly 13 (milliseconds)", digits, timestamp)
	}
}

// isPureIntegerAtoi reports whether a string contains a valid base-10 integer.
// It returns true only if the string can be fully converted to an integer.
func isPureIntegerAtoi(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// isUnixTimestamp reports whether a string should be treated as a Unix
// timestamp: a pure integer of 10 to 13 digits; 10 digits are seconds and
// 13 are milliseconds, while ambiguous 11 and 12 digit values are rejected
// with an error by unixStringToTime instead of falling through to the
// date/time parser; other digit counts are excluded so values such as
// "2024" or a 14-digit compact date/time like "20240101080102" are still
// parsed as date/times
func isUnixTimestamp(s string) bool {
	digits := timestampDigits(s)
	return isPureIntegerAtoi(s) && digits >= 10 && digits <= 13
}

// parseIntegerDateTime parses a pure-integer date/time that is not a Unix
// timestamp: 4 digits are a year, 8 digits a compact date, and 14 digits a
// compact date/time, all interpreted in the given time zone; any other
// digit count errors rather than falling through to a parser that would
// misread the digits as a time of day on the current date
func parseIntegerDateTime(source string, loc *time.Location) (time.Time, error) {
	var layout string
	switch timestampDigits(source) {
	case 4:
		layout = "2006"
	case 8:
		layout = "20060102"
	case 14:
		layout = "20060102150405"
	default:
		return time.Time{}, fmt.Errorf("ambiguous integer date/time %q: expected 4 digits (year), 8 (date), 10 (seconds), 13 (milliseconds), or 14 (date/time)", source)
	}
	t, err := time.ParseInLocation(layout, source, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to parse integer date/time %q: %w", source, err)
	}
	return t, nil
}

// parseDateTimeOrUnix parses a date/time string, treating 10-digit (seconds)
// and 13-digit (milliseconds) integers as Unix timestamps; negative integers
// are rejected because timestamps can't be negative, other integers are
// parsed as compact date/times, and anything else is converted from a
// relative date and parsed as a date/time; empty input is rejected because
// the lenient fallback parsers silently read it as the current time
func parseDateTimeOrUnix(source string) (time.Time, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return time.Time{}, ErrEmptyInput
	}
	if isPureIntegerAtoi(source) {
		if strings.HasPrefix(source, "-") {
			return time.Time{}, fmt.Errorf("timestamps can't be negative: %v", source)
		}
		if isUnixTimestamp(source) {
			return unixStringToTime(source)
		}
		return parseIntegerDateTime(source, time.Local)
	}
	return parseDateTime(ConvertRelativeDateToActual(source))
}
