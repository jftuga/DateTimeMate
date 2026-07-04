// durmath.go implements duration arithmetic: adding or subtracting two
// durations that may be expressed in different units. Results are signed
// (unless Absolute is set) and rendered either as a largest-to-smallest
// breakdown from years down to seconds (extended with sub-second units only
// when the result carries a sub-second remainder) or converted to
// caller-specified target units.

package DateTimeMate

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// ErrNegativeDuration is returned when either duration operand contains a
// negative amount; the sign of the operation is already carried by Add/Sub,
// so negative inputs are rejected.
var ErrNegativeDuration = errors.New("durmath does not accept negative durations; use -a/--add or -s/--sub to control the operation")

// DurMath adds or subtracts two durations, First and Second, which may be
// expressed in different units. Target optionally converts the result to a
// specific group of units; when empty, the result is a full breakdown from
// years down to seconds. Brief renders compact output such as "2h15m".
// Decimals sets the number of decimal places on the smallest output unit.
// Absolute renders the result without a sign, e.g. -15 minutes becomes
// 15 minutes.
type DurMath struct {
	First    string
	Second   string
	Target   string
	Brief    bool
	Decimals int
	Absolute bool
}

// OptionsDurMath is a functional option used to configure a DurMath.
type OptionsDurMath func(*DurMath)

// NewDurMath returns a DurMath configured with the given options.
func NewDurMath(options ...OptionsDurMath) *DurMath {
	dm := &DurMath{}
	for _, opt := range options {
		opt(dm)
	}
	return dm
}

// DurMathWithFirst sets the first duration operand.
func DurMathWithFirst(first string) OptionsDurMath {
	return func(dm *DurMath) {
		dm.First = first
	}
}

// DurMathWithSecond sets the second duration operand.
func DurMathWithSecond(second string) OptionsDurMath {
	return func(dm *DurMath) {
		dm.Second = second
	}
}

// DurMathWithTarget sets the target units for the result; an empty target
// selects the default full-breakdown output.
func DurMathWithTarget(target string) OptionsDurMath {
	return func(dm *DurMath) {
		dm.Target = target
	}
}

// DurMathWithBrief enables brief output, such as: 2h15m
func DurMathWithBrief(brief bool) OptionsDurMath {
	return func(dm *DurMath) {
		dm.Brief = brief
	}
}

// DurMathWithDecimals sets the number of decimal places used when formatting
// the last (smallest) output unit; 0 keeps the default integer truncation
func DurMathWithDecimals(decimals int) OptionsDurMath {
	return func(dm *DurMath) {
		dm.Decimals = decimals
	}
}

// DurMathWithAbsolute makes Add and Sub return an absolute (positive)
// duration; a negative result is rendered without the leading "-"
func DurMathWithAbsolute(absolute bool) OptionsDurMath {
	return func(dm *DurMath) {
		dm.Absolute = absolute
	}
}

// String returns a human-readable summary of the DurMath configuration.
func (dm *DurMath) String() string {
	return fmt.Sprintf("First:%v Second:%v Target:%v Brief:%v Decimals:%v Absolute:%v", dm.First, dm.Second, dm.Target, dm.Brief, dm.Decimals, dm.Absolute)
}

// Add returns the sum of the two durations.
func (dm *DurMath) Add() (string, error) {
	return dm.compute(false)
}

// Sub returns the signed difference of the two durations, First minus Second;
// a negative result is rendered with a single leading "-" unless Absolute
// is set.
func (dm *DurMath) Sub() (string, error) {
	return dm.compute(true)
}

// default output unit breakdowns: years down to seconds, extended through
// nanoseconds only when the result has a non-zero sub-second remainder
var (
	durMathDefaultUnits = []string{"years", "weeks", "days", "hours", "minutes", "seconds"}
	durMathAllUnits     = []string{"years", "weeks", "days", "hours", "minutes", "seconds", "milliseconds", "microseconds", "nanoseconds"}
)

// checkNegativeDuration rejects an operand containing "-" with
// ErrNegativeDuration. A legal duration never contains "-" (units are letters,
// amounts are digits and dots), so a whole-string check also catches
// mid-string negative amounts such as "1 year -30 days" that the parse
// machinery would otherwise accept.
func checkNegativeDuration(source string) error {
	if strings.Contains(source, "-") {
		return ErrNegativeDuration
	}
	return nil
}

// compute parses both operands, adds or subtracts them, and formats the
// signed result; the absolute value is formatted and a leading "-" is
// prepended when the result is negative
func (dm *DurMath) compute(subtract bool) (string, error) {
	if dm.Decimals < 0 || dm.Decimals > 9 {
		return "", fmt.Errorf("decimals must be between 0 and 9: %d", dm.Decimals)
	}
	if err := checkNegativeDuration(dm.First); err != nil {
		return "", err
	}
	if err := checkNegativeDuration(dm.Second); err != nil {
		return "", err
	}
	first, err := parseDurationSeconds(dm.First)
	if err != nil {
		return "", fmt.Errorf("first duration: %w", err)
	}
	second, err := parseDurationSeconds(dm.Second)
	if err != nil {
		return "", fmt.Errorf("second duration: %w", err)
	}

	result := first + second
	if subtract {
		result = first - second
	}
	abs := math.Abs(result)
	if math.Round(abs*1e9) == 0 {
		// snap float noise below half a nanosecond to exactly zero so a
		// zero result renders as "0 seconds", never "-0 seconds"
		result, abs = 0, 0
	}
	negative := result < 0

	units := durMathDefaultUnits
	if dm.Target != "" {
		units, err = resolveTargetUnits(dm.Target)
		if err != nil {
			return "", err
		}
	} else {
		// extend with sub-second units only when a sub-second remainder
		// survives rounding to whole nanoseconds, so binary float residue
		// does not trigger a bogus extension
		frac := abs - math.Floor(abs)
		if r := math.Round(frac * 1e9); r >= 1 && r < 1e9 {
			units = durMathAllUnits
		}
	}

	formatter := &Conv{Decimals: dm.Decimals}
	out := formatter.formatTarget(abs, units)
	if dm.Brief {
		out = shrinkPeriod(out)
	}
	out = strings.TrimSpace(out)
	if negative && !dm.Absolute {
		out = "-" + out
	}
	return out, nil
}
