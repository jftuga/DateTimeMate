package DateTimeMate

import (
	_ "embed"
	"fmt"
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
	ModVersion string = "1.9.0"
	ModUrl     string = "https://github.com/jftuga/DateTimeMate"
)

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
		return carbon.Now().String()
	case "today":
		return carbon.Now().String()
	case "yesterday":
		return carbon.Yesterday().String()
	case "tomorrow":
		return carbon.Tomorrow().String()
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

// parseDateTime parses a date/time string with parsetime, correcting the
// DST offset of zone-less strings via fixLocalZone
func parseDateTime(source string) (time.Time, error) {
	p, err := parsetime.NewParseTime()
	if err != nil {
		return time.Time{}, err
	}
	t, err := p.Parse(source)
	if err != nil {
		return time.Time{}, err
	}
	return fixLocalZone(t), nil
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

	var t time.Time
	if isPureIntegerAtoi(source) {
		if source[0] == '-' {
			return "", fmt.Errorf("timestamps can't be negative: %v", source)
		}
		t, err = unixStringToTime(source)
	} else {
		t, err = parseDateTime(ConvertRelativeDateToActual(source))
	}
	if err != nil {
		return "", err
	}
	return f.FormatString(t), nil
}

// unixStringToTime converts a string containing a Unix timestamp to time.Time.
// It accepts timestamps in seconds (up to 10 digits) and milliseconds (13 digits).
// Returns the corresponding time.Time and any error encountered during conversion.
//
// If the input string is not a valid integer, is empty, or has an ambiguous
// length (11, 12, or more than 13 digits), it returns a zero time.Time and an error.
func unixStringToTime(timestamp string) (time.Time, error) {
	timestamp = strings.TrimSpace(timestamp)
	unixTime, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	switch {
	case len(timestamp) <= 10:
		return time.Unix(unixTime, 0), nil
	case len(timestamp) == 13:
		return time.UnixMilli(unixTime), nil
	default:
		return time.Time{}, fmt.Errorf("ambiguous timestamp length %d for %q: expected up to 10 digits (seconds) or exactly 13 (milliseconds)", len(timestamp), timestamp)
	}
}

// isPureIntegerAtoi reports whether a string contains a valid base-10 integer.
// It returns true only if the string can be fully converted to an integer.
func isPureIntegerAtoi(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// isUnixTimestamp reports whether a string should be treated as a Unix
// timestamp: a pure integer of 10 to 13 characters; 10 characters are
// seconds and 13 are milliseconds, while ambiguous 11 and 12 character
// values are rejected with an error by unixStringToTime instead of falling
// through to the date/time parser; other lengths are excluded so values
// such as "2024" or a 14-digit compact date/time like "20240101080102"
// are still parsed as date/times
func isUnixTimestamp(s string) bool {
	return isPureIntegerAtoi(s) && len(s) >= 10 && len(s) <= 13
}

// parseDateTimeOrUnix parses a date/time string, treating 10-digit (seconds)
// and 13-digit (milliseconds) integers as Unix timestamps; anything else is
// converted from a relative date and parsed with parsetime
func parseDateTimeOrUnix(source string) (time.Time, error) {
	if isUnixTimestamp(source) {
		return unixStringToTime(source)
	}
	return parseDateTime(ConvertRelativeDateToActual(source))
}
