package DateTimeMate

import (
	"fmt"
	"github.com/golang-module/carbon/v2"
	"github.com/lestrrat-go/strftime"
	"github.com/tkuchiki/parsetime"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// used for Dur.Op
const (
	Add = iota
	Sub
)

type Dur struct {
	From         string
	Op           int
	Period       string
	Repeat       int
	Until        string
	OutputFormat string
}

type OptionsDur func(*Dur)

const (
	expanded  string = `(\d+)\s(years?|months?|weeks?|days?|hours?|minutes?|seconds?|milliseconds?|microseconds?|nanoseconds?)`
	wordsOnly string = `\b[a-zA-Z]+\b`
	hintMsg   string = "Hint: duplicate durations not allowed; dates in uppercase; times in lowercase"
)

var carbonFuncs = map[string]interface{}{
	"year":        [2]interface{}{carbon.Carbon.AddYears, carbon.Carbon.SubYears},
	"month":       [2]interface{}{carbon.Carbon.AddMonths, carbon.Carbon.SubMonths},
	"week":        [2]interface{}{carbon.Carbon.AddWeeks, carbon.Carbon.SubWeeks},
	"day":         [2]interface{}{carbon.Carbon.AddDays, carbon.Carbon.SubDays},
	"hour":        [2]interface{}{carbon.Carbon.AddHours, carbon.Carbon.SubHours},
	"minute":      [2]interface{}{carbon.Carbon.AddMinutes, carbon.Carbon.SubMinutes},
	"second":      [2]interface{}{carbon.Carbon.AddSeconds, carbon.Carbon.SubSeconds},
	"millisecond": [2]interface{}{carbon.Carbon.AddMilliseconds, carbon.Carbon.SubMilliseconds},
	"microsecond": [2]interface{}{carbon.Carbon.AddMicroseconds, carbon.Carbon.SubMicroseconds},
	"nanosecond":  [2]interface{}{carbon.Carbon.AddNanoseconds, carbon.Carbon.SubNanoseconds},
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

func DurWithOp(op int) OptionsDur {
	return func(dur *Dur) {
		dur.Op = op
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
	return fmt.Sprintf("%v,%v,%v,%v,%v,%v", dur.From, dur.Period, dur.Op, dur.Repeat, dur.Until, dur.OutputFormat)
}

func (dur *Dur) Add() ([]string, error) {
	return dur.addOrSub(Add)
}

func (dur *Dur) Sub() ([]string, error) {
	return dur.addOrSub(Sub)
}

// addOrSub - calculates a date/time when given a starting date/time and a duration
// also handle: the repeat and until options, relative dates, output formatting
func (dur *Dur) addOrSub(op int) ([]string, error) {
	if dur.Repeat > 0 && dur.Until != "" {
		return nil, fmt.Errorf("repeat & until are mutually exclusive")
	}

	var all []string
	var err error
	if dur.Repeat == 0 && dur.Until == "" {
		var c string
		c, err = calculate(dur.From, dur.Period, op)
		if err != nil {
			return nil, err
		}
		all = []string{c}
	} else if dur.Repeat > 0 && dur.Until == "" {
		from := dur.From
		for i := 0; i < dur.Repeat; i++ {
			from, err = calculate(from, dur.Period, op)
			if err != nil {
				return nil, err
			}
			all = append(all, from)
		}
	} else if dur.Repeat == 0 && dur.Until != "" {
		var f, u time.Time
		var err error

		until := ConvertRelativeDateToActual(dur.Until)
		p, err := parsetime.NewParseTime()
		if err != nil {
			return nil, err
		}
		u, err = p.Parse(until)
		if err != nil {
			return nil, err
		}

		from := ConvertRelativeDateToActual(dur.From)
		for {
			from, err = calculate(from, dur.Period, op)
			if err != nil {
				return nil, err
			}

			p, err := parsetime.NewParseTime()
			if err != nil {
				return nil, err
			}
			f, err = p.Parse(from)
			if err != nil {
				return nil, err
			}

			if Add == op {
				if f.After(u) {
					break
				}
			} else {
				if f.Before(u) {
					break
				}
			}
			all = append(all, from)
		}
	}

	if len(dur.OutputFormat) > 0 && len(all) > 0 {
		var allWithFormat []string
		for _, a := range all {
			formatted, err := dur.setOutputFormat(a)
			if err != nil {
				return nil, err
			}
			allWithFormat = append(allWithFormat, formatted)
		}
		return allWithFormat, nil
	}
	return all, nil
}

// setOutputFormat use a strftime format string for the output date/time
func (dur *Dur) setOutputFormat(arg string) (string, error) {
	f, err := strftime.New(dur.OutputFormat)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	p, err := parsetime.NewParseTime()
	if err != nil {
		return "", err
	}
	s, err := p.Parse(arg)
	if err != nil {
		return "", err
	}
	output := f.FormatString(s)
	return output, nil
}

// calculate given a from date and period duration, compute a new date/time
// when index==0, then add; when index==1, then subtract
func calculate(from, period string, index int) (string, error) {
	periodMatches := expandedRegexp.FindAllStringSubmatch(period, -1)
	if len(periodMatches) == 0 {
		// brief format is being used so first expand it to the long format
		period, err := expandPeriod(period)
		if nil != err {
			return "", fmt.Errorf("%v", err)
		}
		periodMatches = expandedRegexp.FindAllStringSubmatch(period, -1)
		if len(periodMatches) == 0 {
			return "", fmt.Errorf("[validatePeriod] Invalid duration: %s", period)
		}
	}

	from = ConvertRelativeDateToActual(from)
	p, err := parsetime.NewParseTime()
	if err != nil {
		return "", err
	}
	f, err := p.Parse(from)
	if err != nil {
		return "", err
	}

	to := carbon.CreateFromStdTime(f)
	if to.Error != nil {
		return "", to.Error
	}
	err = validatePeriod(period)
	if err != nil {
		return "", err
	}

	for i := range periodMatches {
		amount := periodMatches[i][1]
		num, err := strconv.Atoi(amount)
		if err != nil {
			return "", err
		}
		word := periodMatches[i][2]
		// to understand this line of code, read: ChatGPT_Explanation.md
		to = carbonFuncs[removeTrailingS(word)].([2]interface{})[index].(func(carbon.Carbon, int) carbon.Carbon)(to, num)
		// fmt.Printf("    to: %v | %v | %v\n", num, word, to)
	}
	return to.ToString(), nil
}

// validatePeriod ensure all words in "period" are a valid time duration
func validatePeriod(period string) error {
	wordsOnlyRe := regexp.MustCompile(wordsOnly)
	matches := wordsOnlyRe.FindAllString(period, -1)
	for _, word := range matches {
		// fmt.Println("word:", word)
		_, ok := carbonFuncs[removeTrailingS(word)]
		if !ok {
			return fmt.Errorf("[validatePeriod] Invalid period: %s", word)
		}
	}
	return nil
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
	s = strings.Replace(s, "M", "ι", 1)
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
	p = strings.Replace(p, "ι", " months ", 1)
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
