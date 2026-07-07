package DateTimeMate

import (
	"fmt"
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

// CalculateDiff returns the time difference between Start and End, both as
// a formatted string and as a time.Duration; both sides are parsed with the
// same shared chain used by every other sub-command (parseDateTimeOrUnix)
// when Absolute is set, both the formatted string and the returned duration are non-negative
func (diff *Diff) CalculateDiff() (string, time.Duration, error) {
	start, err := parseDateTimeOrUnix(diff.Start)
	if err != nil {
		return "", 0, err
	}
	end, err := parseDateTimeOrUnix(diff.End)
	if err != nil {
		return "", 0, err
	}

	duration := end.Sub(start)
	// time.Time.Sub silently clamps a difference that overflows
	// time.Duration, so detect saturation by checking the round trip; this
	// accepts every representable span (about +/-292 years) instead of
	// rejecting by calendar-year count
	if !start.Add(duration).Equal(end) {
		return "", 0, fmt.Errorf("difference between %q and %q exceeds the representable range of about 292 years", diff.Start, diff.End)
	}
	if diff.Absolute {
		duration = duration.Abs()
	}
	parsed := durafmt.Parse(duration)
	difference := fmt.Sprintf("%v", parsed)
	// durafmt's smallest unit is the microsecond, so append any
	// sub-microsecond remainder: a 1500ns diff renders as
	// "1 microsecond 500 nanoseconds" and a sub-microsecond diff is a
	// nanosecond count instead of an empty string
	if rem := duration % time.Microsecond; rem != 0 {
		ns := rem.Nanoseconds()
		if ns < 0 {
			ns = -ns
		}
		unit := "nanoseconds"
		if ns == 1 {
			unit = "nanosecond"
		}
		nsPart := fmt.Sprintf("%d %s", ns, unit)
		// durafmt renders a sub-microsecond duration as "" ("-" when
		// negative), so the nanosecond count becomes the whole output
		if difference == "" || difference == "-" {
			if duration < 0 {
				nsPart = "-" + nsPart
			}
			difference = nsPart
		} else {
			difference += " " + nsPart
		}
	}
	if diff.Brief {
		difference = shrinkPeriod(difference)
	}
	return difference, duration, nil
}
