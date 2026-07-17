// Package datecalc applies a signed number of calendar or clock units to a
// time.Time. It replaces the arithmetic portion of
// github.com/golang-module/carbon: years, weeks, and days use the
// calendar-aware, overflow-normalizing time.Time.AddDate (Feb 29 + 1 year =
// Mar 1), while hours through nanoseconds use the absolute time.Time.Add.
package datecalc

import (
	"fmt"
	"math"
	"time"
)

// clockUnits are the units applied as absolute durations via time.Time.Add.
var clockUnits = map[string]time.Duration{
	"hour":        time.Hour,
	"minute":      time.Minute,
	"second":      time.Second,
	"millisecond": time.Millisecond,
	"microsecond": time.Microsecond,
	"nanosecond":  time.Nanosecond,
}

// Apply adds (sign=+1) or subtracts (sign=-1) n units to t. unit is one of:
// year, week, day, hour, minute, second, millisecond, microsecond,
// nanosecond (singular, lowercase).
func Apply(t time.Time, unit string, n int, sign int) (time.Time, error) {
	switch unit {
	case "year":
		return t.AddDate(sign*n, 0, 0), nil
	case "week":
		return t.AddDate(0, 0, sign*n*7), nil
	case "day":
		return t.AddDate(0, 0, sign*n), nil
	}
	if d, ok := clockUnits[unit]; ok {
		// the multiplication below is int64 nanoseconds, which silently
		// wraps for large counts of the bigger units (about 2.56 million
		// hours); reject anything that cannot be represented
		if limit := math.MaxInt64 / int64(d); int64(n) > limit || int64(n) < -limit {
			return t, fmt.Errorf("%d %ss exceeds the supported range of about 292 years", n, unit)
		}
		return t.Add(time.Duration(sign*n) * d), nil
	}
	return t, fmt.Errorf("unknown unit: %q", unit)
}
