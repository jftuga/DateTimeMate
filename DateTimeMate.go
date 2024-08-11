package DateTimeMate

import (
	"github.com/golang-module/carbon/v2"
	"github.com/lestrrat-go/strftime"
	"github.com/tkuchiki/parsetime"
	"strings"
)

const (
	ModName    string = "DateTimeMate"
	ModVersion string = "1.2.2"
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

// Reformat the source string to match the strftime outputFormat
// Ex: "2024-07-22 08:21:44", "%v %r" => "22-Jul-2024 08:21:44 AM"
func Reformat(source string, outputFormat string) (string, error) {
	source = ConvertRelativeDateToActual(source)
	f, err := strftime.New(outputFormat)
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
