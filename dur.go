package DateTimeMate

import (
	"fmt"
	"github.com/golang-module/carbon/v2"
	"github.com/lestrrat-go/strftime"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// internal operation selectors for addOrSub
const (
	opAdd = iota
	opSub
)

type Dur struct {
	From         string
	Period       string
	Repeat       int
	Until        string
	OutputFormat string
}

type OptionsDur func(*Dur)

const (
	// a period is a series of amount/unit pairs; the amount must start at the
	// beginning of the string or after whitespace so that a fractional amount
	// such as "1.5" can never be partially matched as "5"
	expanded string = `(?:^|\s)(\d+(?:\.\d+)?)\s(years?|weeks?|days?|hours?|minutes?|seconds?|milliseconds?|microseconds?|nanoseconds?)`
	hintMsg  string = "Hint: duplicate durations not allowed; dates in uppercase; times in lowercase"

	// maxUntilIterations is a backstop against unbounded output from the until option
	maxUntilIterations = 1_000_000
)

var carbonFuncs = map[string]interface{}{
	"year":        [2]interface{}{carbon.Carbon.AddYears, carbon.Carbon.SubYears},
	"week":        [2]interface{}{carbon.Carbon.AddWeeks, carbon.Carbon.SubWeeks},
	"day":         [2]interface{}{carbon.Carbon.AddDays, carbon.Carbon.SubDays},
	"hour":        [2]interface{}{carbon.Carbon.AddHours, carbon.Carbon.SubHours},
	"minute":      [2]interface{}{carbon.Carbon.AddMinutes, carbon.Carbon.SubMinutes},
	"second":      [2]interface{}{carbon.Carbon.AddSeconds, carbon.Carbon.SubSeconds},
	"millisecond": [2]interface{}{carbon.Carbon.AddMilliseconds, carbon.Carbon.SubMilliseconds},
	"microsecond": [2]interface{}{carbon.Carbon.AddMicroseconds, carbon.Carbon.SubMicroseconds},
	"nanosecond":  [2]interface{}{carbon.Carbon.AddNanoseconds, carbon.Carbon.SubNanoseconds},
}

// unitNanoseconds is used to apply the fractional part of an amount, so the
// calendar-aware carbon functions still handle the integer part
var unitNanoseconds = map[string]float64{
	"year":        365.25 * 24 * float64(time.Hour),
	"week":        7 * 24 * float64(time.Hour),
	"day":         24 * float64(time.Hour),
	"hour":        float64(time.Hour),
	"minute":      float64(time.Minute),
	"second":      float64(time.Second),
	"millisecond": float64(time.Millisecond),
	"microsecond": float64(time.Microsecond),
	"nanosecond":  1,
}

var expandedRegexp = regexp.MustCompile(expanded)

func NewDur(options ...OptionsDur) *Dur {
	dur := &Dur{}
	for _, opt := range options {
		opt(dur)
	}
	return dur
}

func DurWithFrom(from string) OptionsDur {
	return func(dur *Dur) {
		dur.From = from
	}
}

func DurWithDur(d string) OptionsDur {
	return func(dur *Dur) {
		dur.Period = d
	}
}

func DurWithRepeat(repeat int) OptionsDur {
	return func(dur *Dur) {
		dur.Repeat = repeat
	}
}
func DurWithUntil(until string) OptionsDur {
	return func(dur *Dur) {
		dur.Until = until
	}
}

func DurWithOutputFormat(outputFormat string) OptionsDur {
	return func(dur *Dur) {
		dur.OutputFormat = outputFormat
	}
}

func (dur *Dur) String() string {
	return fmt.Sprintf("From:%v Period:%v Repeat:%v Until:%v OutputFormat:%v", dur.From, dur.Period, dur.Repeat, dur.Until, dur.OutputFormat)
}

func (dur *Dur) Add() ([]string, error) {
	return dur.addOrSub(opAdd)
}

func (dur *Dur) Sub() ([]string, error) {
	return dur.addOrSub(opSub)
}

// addOrSub - calculates a date/time when given a starting date/time and a duration
// also handle: the repeat and until options, relative dates, Unix timestamps,
// output formatting
func (dur *Dur) addOrSub(op int) ([]string, error) {
	if dur.Repeat < 0 {
		return nil, fmt.Errorf("repeat must not be negative: %d", dur.Repeat)
	}
	if dur.Repeat > 0 && dur.Until != "" {
		return nil, fmt.Errorf("repeat & until are mutually exclusive")
	}

	f, err := parseDateTimeOrUnix(dur.From)
	if err != nil {
		return nil, err
	}
	from := carbon.CreateFromStdTime(f)
	if from.Error != nil {
		return nil, from.Error
	}
	periodMatches, err := parsePeriod(dur.Period)
	if err != nil {
		return nil, err
	}

	var all []carbon.Carbon
	switch {
	case dur.Repeat == 0 && dur.Until == "":
		to, err := applyPeriod(from, periodMatches, op)
		if err != nil {
			return nil, err
		}
		all = append(all, to)
	case dur.Repeat > 0:
		to := from
		for i := 0; i < dur.Repeat; i++ {
			to, err = applyPeriod(to, periodMatches, op)
			if err != nil {
				return nil, err
			}
			all = append(all, to)
		}
	default: // until
		u, err := parseDateTimeOrUnix(dur.Until)
		if err != nil {
			return nil, err
		}
		to := from
		for i := 0; ; i++ {
			if i >= maxUntilIterations {
				return nil, fmt.Errorf("until would produce more than %d results", maxUntilIterations)
			}
			next, err := applyPeriod(to, periodMatches, op)
			if err != nil {
				return nil, err
			}
			if next.StdTime().Equal(to.StdTime()) {
				return nil, fmt.Errorf("duration %q does not advance toward the until date/time", dur.Period)
			}
			to = next
			if opAdd == op {
				if to.StdTime().After(u) {
					break
				}
			} else {
				if to.StdTime().Before(u) {
					break
				}
			}
			all = append(all, to)
		}
	}
	return dur.renderResults(all)
}

// renderResults converts computed date/times to strings, applying the
// optional strftime output format which also supports the unix time %s modifier
func (dur *Dur) renderResults(all []carbon.Carbon) ([]string, error) {
	rendered := make([]string, 0, len(all))
	if len(dur.OutputFormat) == 0 {
		for _, c := range all {
			rendered = append(rendered, c.ToString())
		}
		return rendered, nil
	}
	f, err := strftime.New(dur.OutputFormat, strftime.WithUnixSeconds('s'))
	if err != nil {
		return nil, err
	}
	for _, c := range all {
		rendered = append(rendered, f.FormatString(c.StdTime()))
	}
	return rendered, nil
}

// parsePeriod parses a period in either long or brief format into
// (amount, unit) pairs, erroring if any part of the period is not understood
func parsePeriod(period string) ([][2]string, error) {
	indexes := expandedRegexp.FindAllStringSubmatchIndex(period, -1)
	if len(indexes) == 0 {
		// brief format is being used so first expand it to the long format
		var err error
		period, err = expandPeriod(period)
		if nil != err {
			return nil, err
		}
		indexes = expandedRegexp.FindAllStringSubmatchIndex(period, -1)
		if len(indexes) == 0 {
			return nil, fmt.Errorf("[parsePeriod] Invalid duration: %s", period)
		}
	}

	// every character must belong to a match, otherwise part of the
	// period, such as the "2m" in "1 hour 2m", would be silently ignored
	var matches [][2]string
	var leftover strings.Builder
	prev := 0
	for _, m := range indexes {
		leftover.WriteString(period[prev:m[0]])
		prev = m[1]
		matches = append(matches, [2]string{period[m[2]:m[3]], period[m[4]:m[5]]})
	}
	leftover.WriteString(period[prev:])
	if remains := strings.TrimSpace(leftover.String()); remains != "" {
		return nil, fmt.Errorf("[parsePeriod] Invalid duration %q in: %s. %s", remains, period, hintMsg)
	}
	return matches, nil
}

// applyPeriod applies each (amount, unit) pair of a parsed period to a date/time
// when index==0, then add; when index==1, then subtract
// the integer part of an amount uses carbon's calendar-aware functions; any
// fractional part is applied as nanoseconds
func applyPeriod(to carbon.Carbon, periodMatches [][2]string, index int) (carbon.Carbon, error) {
	for _, match := range periodMatches {
		amount, word := match[0], removeTrailingS(match[1])
		value, err := strconv.ParseFloat(amount, 64)
		if err != nil {
			return to, err
		}
		if value > math.MaxInt32 {
			return to, fmt.Errorf("amount too large: %s", amount)
		}
		whole, frac := math.Modf(value)
		to = carbonFuncs[word].([2]interface{})[index].(func(carbon.Carbon, int) carbon.Carbon)(to, int(whole))
		if frac > 0 {
			ns := int(math.Round(frac * unitNanoseconds[word]))
			to = carbonFuncs["nanosecond"].([2]interface{})[index].(func(carbon.Carbon, int) carbon.Carbon)(to, ns)
		}
		if to.Error != nil {
			return to, to.Error
		}
	}
	return to, nil
}

// expandPeriod convert a brief style period into a long period
// only allow one replacement per each period
// Ex: 1h2m3s => 1 hour 2 minutes 3 seconds
func expandPeriod(period string) (string, error) {
	// a direct string replace will not work because some
	// periods have overlapping strings, such as 's' with 'ms, 'us', 'ns'
	// therefore convert each period to a unique string first
	s := period
	s = strings.Replace(s, "ns", "α", 1)
	s = strings.Replace(s, "us", "β", 1)
	s = strings.Replace(s, "µs", "β", 1)
	s = strings.Replace(s, "ms", "γ", 1)
	s = strings.Replace(s, "s", "δ", 1)
	s = strings.Replace(s, "m", "ε", 1)
	s = strings.Replace(s, "h", "ζ", 1)
	s = strings.Replace(s, "D", "η", 1)
	s = strings.Replace(s, "W", "θ", 1)
	// Month (M) not supported
	s = strings.Replace(s, "Y", "λ", 1)

	// now convert from the unique string back to the corresponding duration
	p := s
	p = strings.Replace(p, "α", " nanoseconds ", 1)
	p = strings.Replace(p, "β", " microseconds ", 1)
	p = strings.Replace(p, "γ", " milliseconds ", 1)
	p = strings.Replace(p, "δ", " seconds ", 1)
	p = strings.Replace(p, "ε", " minutes ", 1)
	p = strings.Replace(p, "ζ", " hours ", 1)
	p = strings.Replace(p, "η", " days ", 1)
	p = strings.Replace(p, "θ", " weeks ", 1)
	// Month (M) not supported
	p = strings.Replace(p, "λ", " years ", 1)

	// ensure each time & period was successfully replaced
	// len of Fields should always be even because is part
	// of the period is a two element tuple of
	// a numeric amount and a duration
	words := strings.Fields(p)
	if len(words)%2 == 1 {
		return "", fmt.Errorf("[expandPeriod] Invalid period: %s. %s", period, hintMsg)
	}

	// check that every other element is a number
	for i := 0; i < len(words); i += 2 {
		_, err := strconv.ParseFloat(words[i], 64)
		if err != nil {
			return "", fmt.Errorf("[expandPeriod] %v. %s", err, hintMsg)
		}
	}
	return p, nil
}
