// Package humandur renders a time.Duration as a human-readable list of
// unit components, e.g. "4 weeks 3 days 1 hour 2 minutes 3 seconds".
// It replaces github.com/hako/durafmt with the exact output semantics the
// diff sub-command has always produced, plus a nanosecond unit so callers
// no longer need to hand-append sub-microsecond remainders.
package humandur

import (
	"fmt"
	"strings"
	"time"
)

// units are flat conversions applied by integer division, largest first:
// 1 year = 365 days exactly and 1 week = 7 days exactly, with no calendar
// awareness (matching durafmt).
var units = []struct {
	name string
	size uint64
}{
	{"year", uint64(365 * 24 * time.Hour)},
	{"week", uint64(7 * 24 * time.Hour)},
	{"day", uint64(24 * time.Hour)},
	{"hour", uint64(time.Hour)},
	{"minute", uint64(time.Minute)},
	{"second", uint64(time.Second)},
	{"millisecond", uint64(time.Millisecond)},
	{"microsecond", uint64(time.Microsecond)},
	{"nanosecond", 1},
}

// Format renders d largest-unit first, omitting zero-valued components,
// pluralizing unit names, and joining components with single spaces.
// A zero duration renders as "0 seconds"; a negative duration renders as
// "-" followed by the positive rendering.
func Format(d time.Duration) string {
	// negate via uint64 so time.Duration's minimum value cannot overflow
	negative := d < 0
	remainder := uint64(d)
	if negative {
		remainder = -remainder
	}
	if remainder == 0 {
		return "0 seconds"
	}
	var parts []string
	for _, unit := range units {
		count := remainder / unit.size
		remainder %= unit.size
		if count == 0 {
			continue
		}
		name := unit.name
		if count != 1 {
			name += "s"
		}
		parts = append(parts, fmt.Sprintf("%d %s", count, name))
	}
	result := strings.Join(parts, " ")
	if negative {
		result = "-" + result
	}
	return result
}
