package datecalc

import (
	"math"
	"testing"
	"time"
)

func TestApply(t *testing.T) {
	t.Parallel()
	base := time.Date(2024, 6, 15, 12, 30, 45, 123456789, time.UTC)
	tests := []struct {
		name string
		t    time.Time
		unit string
		n    int
		sign int
		want time.Time
	}{
		{"add year", base, "year", 1, +1, time.Date(2025, 6, 15, 12, 30, 45, 123456789, time.UTC)},
		{"sub year", base, "year", 2, -1, time.Date(2022, 6, 15, 12, 30, 45, 123456789, time.UTC)},
		{"leap day plus one year normalizes to Mar 1", time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), "year", 1, +1, time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)},
		{"add week", base, "week", 2, +1, time.Date(2024, 6, 29, 12, 30, 45, 123456789, time.UTC)},
		{"sub week crosses month", base, "week", 3, -1, time.Date(2024, 5, 25, 12, 30, 45, 123456789, time.UTC)},
		{"add day", base, "day", 20, +1, time.Date(2024, 7, 5, 12, 30, 45, 123456789, time.UTC)},
		{"sub day", base, "day", 15, -1, time.Date(2024, 5, 31, 12, 30, 45, 123456789, time.UTC)},
		{"add hour", base, "hour", 13, +1, time.Date(2024, 6, 16, 1, 30, 45, 123456789, time.UTC)},
		{"sub minute", base, "minute", 31, -1, time.Date(2024, 6, 15, 11, 59, 45, 123456789, time.UTC)},
		{"add second", base, "second", 15, +1, time.Date(2024, 6, 15, 12, 31, 0, 123456789, time.UTC)},
		{"add millisecond", base, "millisecond", 2, +1, time.Date(2024, 6, 15, 12, 30, 45, 125456789, time.UTC)},
		{"sub microsecond", base, "microsecond", 456, -1, time.Date(2024, 6, 15, 12, 30, 45, 123000789, time.UTC)},
		{"add nanosecond", base, "nanosecond", 211, +1, time.Date(2024, 6, 15, 12, 30, 45, 123457000, time.UTC)},
		{"zero count is identity", base, "day", 0, +1, base},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Apply(tt.t, tt.unit, tt.n, tt.sign)
			if err != nil {
				t.Fatal(err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("Apply(%v, %q, %d, %d) = %v, want %v", tt.t, tt.unit, tt.n, tt.sign, got, tt.want)
			}
		})
	}
}

func TestApplyDST(t *testing.T) {
	t.Parallel()
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	// 2024-03-10 02:00 EST -> 03:00 EDT: the day is only 23 real hours
	before := time.Date(2024, 3, 9, 12, 0, 0, 0, loc)
	// day arithmetic is calendar-aware: same wall clock the next day
	gotDay, err := Apply(before, "day", 1, +1)
	if err != nil {
		t.Fatal(err)
	}
	if gotDay.Hour() != 12 {
		t.Errorf("adding 1 day across DST should keep the wall clock at 12, got %d", gotDay.Hour())
	}
	// hour arithmetic is absolute: 24 real hours lands at 13:00 EDT
	gotHours, err := Apply(before, "hour", 24, +1)
	if err != nil {
		t.Fatal(err)
	}
	if gotHours.Hour() != 13 {
		t.Errorf("adding 24 hours across DST should land at 13, got %d", gotHours.Hour())
	}
}

func TestApplyOverflowRejected(t *testing.T) {
	t.Parallel()
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	// counts whose nanosecond total exceeds int64 used to wrap silently
	// (3,000,000 hours moved the date backward to 1781)
	for _, tt := range []struct {
		unit string
		n    int
	}{
		{"hour", 3_000_000},
		{"minute", 200_000_000},
	} {
		if _, err := Apply(base, tt.unit, tt.n, +1); err == nil {
			t.Errorf("Apply(%d %ss): expected an overflow error, got nil", tt.n, tt.unit)
		}
		if _, err := Apply(base, tt.unit, tt.n, -1); err == nil {
			t.Errorf("Apply(-%d %ss): expected an overflow error, got nil", tt.n, tt.unit)
		}
	}
	// the largest representable hour count still works
	limit := int(int64(math.MaxInt64) / int64(time.Hour))
	got, err := Apply(base, "hour", limit, +1)
	if err != nil {
		t.Fatal(err)
	}
	if !got.After(base) {
		t.Errorf("Apply(%d hours) = %v, expected a time after %v", limit, got, base)
	}
}

func TestApplyUnknownUnit(t *testing.T) {
	t.Parallel()
	if _, err := Apply(time.Now(), "month", 1, +1); err == nil {
		t.Error("expected an error for an unsupported unit, got nil")
	}
}
