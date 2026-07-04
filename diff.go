package DateTimeMate

import (
	"fmt"
	"github.com/golang-module/carbon/v2"
	"github.com/hako/durafmt"
	"time"
)

type Diff struct {
	Start    string
	End      string
	Brief    bool
	Absolute bool
}

type OptionsDiff func(*Diff)

func NewDiff(options ...OptionsDiff) *Diff {
	diff := &Diff{}
	for _, opt := range options {
		opt(diff)
	}
	return diff
}

func DiffWithStart(start string) OptionsDiff {
	return func(opt *Diff) {
		opt.Start = start
	}
}

func DiffWithEnd(end string) OptionsDiff {
	return func(opt *Diff) {
		opt.End = end
	}
}

func DiffWithBrief(brief bool) OptionsDiff {
	return func(opt *Diff) {
		opt.Brief = brief
	}
}

// DiffWithAbsolute makes CalculateDiff return an absolute (positive)
// duration and formatted string regardless of argument order
func DiffWithAbsolute(absolute bool) OptionsDiff {
	return func(opt *Diff) {
		opt.Absolute = absolute
	}
}

func (diff *Diff) String() string {
	return fmt.Sprintf("Start:%v End:%v Brief:%v Absolute:%v", diff.Start, diff.End, diff.Brief, diff.Absolute)
}

// parseDiffTime parses one side of a diff: 10-digit (seconds) and 13-digit
// (milliseconds) integers are treated as Unix timestamps; anything else is
// parsed with carbon first, falling back to parsetime if carbon fails
func parseDiffTime(source string) (time.Time, error) {
	if isUnixTimestamp(source) {
		return unixStringToTime(source)
	}
	converted := ConvertRelativeDateToActual(source)
	if c := carbon.Parse(converted); c.Error == nil {
		return c.StdTime(), nil
	}
	return parseDateTime(converted)
}

// CalculateDiff return the time difference and also set dt.Diff
// first try to parse with carbon, fallback to parsing with parsetime if carbon fails to parse
// when Absolute is set, both the formatted string and the returned duration are non-negative
func (diff *Diff) CalculateDiff() (string, time.Duration, error) {
	start, err := parseDiffTime(diff.Start)
	if err != nil {
		return "", 0, err
	}
	end, err := parseDiffTime(diff.End)
	if err != nil {
		return "", 0, err
	}

	yearDiff := end.Year() - start.Year()
	if yearDiff > 291 || yearDiff < -291 {
		return "", 0, fmt.Errorf("year difference of %d exceeds 291 year maximum", yearDiff)
	}

	duration := end.Sub(start)
	if diff.Absolute {
		duration = duration.Abs()
	}
	parsed := durafmt.Parse(duration)
	difference := fmt.Sprintf("%v", parsed)
	if diff.Brief {
		difference = shrinkPeriod(difference)
	}
	return difference, duration, nil
}
