package DateTimeMate

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-module/carbon/v2"
	"github.com/lestrrat-go/strftime"
	"github.com/tkuchiki/parsetime"
)

const (
	ModName    string = "DateTimeMate"
	ModVersion string = "1.3.1"
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
	{"months", "M"},
	{"years", "Y"},
	{"nanosecond", "ns"},
	{"microsecond", "us"},
	{"millisecond", "ms"},
	{"second", "s"},
	{"minute", "m"},
	{"hour", "h"},
	{"day", "D"},
	{"week", "W"},
	{"month", "M"},
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

// Reformat converts a date/time string into a specified format. The source can be:
//   - A Unix timestamp (e.g., "1700265600")
//   - A relative date (e.g., "yesterday", "now")
//   - Any other date format parseable by parsetime
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
	if isPureIntegerAtoi(source) {
		if source[0] == '-' {
			return "", fmt.Errorf("timestamps can't be negative: %v", source)
		}
		t, err := unixStringToTime(source)
		if err != nil {
			return "", err
		}
		source = t.String()
	} else {
		source = ConvertRelativeDateToActual(source)
	}

	// creates a new Strftime instance
	// outputFormat is a pattern string that follows strftime formatting
	// the additional formatting behavior allows this to also use the unix time %s modifier
	f, err := strftime.New(outputFormat, strftime.WithUnixSeconds('s'))
	if err != nil {
		return "", err
	}
	p, err := parsetime.NewParseTime()
	if err != nil {
		return "", err
	}
	s, err := p.Parse(source)
	if err != nil {
		return "", err

	}
	return f.FormatString(s), nil
}

// unixStringToTime converts a string containing a Unix timestamp to time.Time.
// It accepts timestamps in both seconds (10 digits) and milliseconds (13 digits).
// Returns the corresponding time.Time and any error encountered during conversion.
//
// If the input string is not a valid integer or is empty,
// it returns a zero time.Time and an error.
func unixStringToTime(timestamp string) (time.Time, error) {
	unixTime, err := strconv.ParseInt(strings.TrimSpace(timestamp), 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	if len(timestamp) == 13 {
		return time.UnixMilli(unixTime), nil
	}

	return time.Unix(unixTime, 0), nil
}

// isPureIntegerAtoi reports whether a string contains a valid base-10 integer.
// It returns true only if the string can be fully converted to an integer.
func isPureIntegerAtoi(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
