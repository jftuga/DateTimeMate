// durmath_test.go verifies the DurMath duration arithmetic API: adding and
// subtracting durations across mixed units, signed results, absolute results,
// target-unit conversion, decimals, brief output, and rejection of invalid or
// negative inputs.
package DateTimeMate

import (
	"errors"
	"testing"
)

func testDurMath(t *testing.T, first, second string, brief, absolute bool, correctAdd, correctSub string) {
	t.Helper()
	dm := NewDurMath(
		DurMathWithFirst(first),
		DurMathWithSecond(second),
		DurMathWithBrief(brief),
		DurMathWithAbsolute(absolute))

	result, err := dm.Add()
	if err != nil {
		t.Error(err)
	}
	if result != correctAdd {
		t.Errorf("\n[computed: %v] !=\n[correct : %v]", result, correctAdd)
	}

	result, err = dm.Sub()
	if err != nil {
		t.Error(err)
	}
	if result != correctSub {
		t.Errorf("\n[computed: %v] !=\n[correct : %v]", result, correctSub)
	}
}

func testDurMathConv(t *testing.T, first, second, target string, brief bool, decimals int, absolute bool, correctAdd, correctSub string) {
	t.Helper()
	dm := NewDurMath(
		DurMathWithFirst(first),
		DurMathWithSecond(second),
		DurMathWithTarget(target),
		DurMathWithBrief(brief),
		DurMathWithDecimals(decimals),
		DurMathWithAbsolute(absolute))

	result, err := dm.Add()
	if err != nil {
		t.Error(err)
	}
	if result != correctAdd {
		t.Errorf("\n[computed: %v] !=\n[correct : %v]", result, correctAdd)
	}

	result, err = dm.Sub()
	if err != nil {
		t.Error(err)
	}
	if result != correctSub {
		t.Errorf("\n[computed: %v] !=\n[correct : %v]", result, correctSub)
	}
}

func TestDurMathHoursMinutes(t *testing.T) {
	t.Parallel()
	testDurMath(t, "1 hour 30 minutes", "45 minutes", false, false, "2 hours 15 minutes", "45 minutes")
	testDurMath(t, "1h30m", "45m", false, false, "2 hours 15 minutes", "45 minutes")
	testDurMath(t, "1h30m", "45m", true, false, "2h15m", "45m")
}

func TestDurMathSignedResult(t *testing.T) {
	t.Parallel()
	// subtraction is signed when the second duration is larger
	testDurMath(t, "45 minutes", "1 hour", false, false, "1 hour 45 minutes", "-15 minutes")
	testDurMath(t, "45 minutes", "1 hour", true, false, "1h45m", "-15m")
	// the sign also survives target-unit conversion
	testDurMathConv(t, "90 minutes", "1 day", "minutes", false, 0, false, "1530 minutes", "-1350 minutes")
}

func TestDurMathAbsoluteResult(t *testing.T) {
	t.Parallel()
	// Absolute renders a negative result without the leading "-"
	testDurMath(t, "45 minutes", "1 hour", false, true, "1 hour 45 minutes", "15 minutes")
	testDurMath(t, "45 minutes", "1 hour", true, true, "1h45m", "15m")
	// Absolute is a no-op on positive results
	testDurMath(t, "1 hour 30 minutes", "45 minutes", false, true, "2 hours 15 minutes", "45 minutes")
	// Absolute composes with target units and decimals
	testDurMathConv(t, "90 minutes", "1 day", "minutes", false, 0, true, "1530 minutes", "1350 minutes")
	testDurMathConv(t, "30 minutes", "1 hour", "hours", false, 1, true, "1.5 hours", "0.5 hours")
	// a zero result still renders as "0 seconds", never "-0 seconds"
	testDurMath(t, "1 hour", "60 minutes", false, true, "2 hours", "0 seconds")
	testDurMath(t, "1 hour", "60 minutes", true, true, "2h", "0s")

	// negative inputs are still rejected when Absolute is set
	dm := NewDurMath(
		DurMathWithFirst("1 year -30 days"),
		DurMathWithSecond("1 hour"),
		DurMathWithAbsolute(true))
	if _, err := dm.Sub(); !errors.Is(err, ErrNegativeDuration) {
		t.Errorf("expected ErrNegativeDuration with Absolute set, got: %v", err)
	}
}

func TestDurMathMixedUnits(t *testing.T) {
	t.Parallel()
	testDurMath(t, "1 week", "3 days 12 hours", false, false, "1 week 3 days 12 hours", "3 days 12 hours")
	testDurMath(t, "1 day", "90 minutes", false, false, "1 day 1 hour 30 minutes", "22 hours 30 minutes")
}

func TestDurMathCaseInsensitiveUnits(t *testing.T) {
	t.Parallel()
	// long-form units are case-insensitive, singular or plural
	testDurMath(t, "1 HOUR 30 Minutes", "45 MINUTES", false, false, "2 hours 15 minutes", "45 minutes")
}

func TestDurMathTargetUnits(t *testing.T) {
	t.Parallel()
	testDurMathConv(t, "1 day", "90 minutes", "minutes", false, 0, false, "1530 minutes", "1350 minutes")
	// brief target specification
	testDurMathConv(t, "1 day", "90 minutes", "hm", false, 0, false, "25 hours 30 minutes", "22 hours 30 minutes")
	testDurMathConv(t, "1 day", "90 minutes", "hm", true, 0, false, "25h30m", "22h30m")
}

func TestDurMathDecimals(t *testing.T) {
	t.Parallel()
	testDurMathConv(t, "1 hour", "30 minutes", "hours", false, 1, false, "1.5 hours", "0.5 hours")
	testDurMathConv(t, "1 hour", "30 minutes", "hours", false, 2, false, "1.50 hours", "0.50 hours")
}

func TestDurMathSubSecond(t *testing.T) {
	t.Parallel()
	// sub-second units appear only when the result has a sub-second remainder
	testDurMath(t, "1.5 seconds", "250 milliseconds", false, false, "1 second 750 milliseconds", "1 second 250 milliseconds")
	testDurMath(t, "1 second", "500 milliseconds", false, false, "1 second 500 milliseconds", "500 milliseconds")
	// whole-second results stay at second granularity
	testDurMath(t, "750 milliseconds", "250 milliseconds", false, false, "1 second", "500 milliseconds")
	// float-imprecise operands round to the nearest nanosecond, so the
	// result is exact instead of 199ms 999us 999ns
	testDurMath(t, "0.3 seconds", "0.1 seconds", false, false, "400 milliseconds", "200 milliseconds")
	// decimals rounding carries into the larger unit
	testDurMathConv(t, "1 minute 59.96 seconds", "0 seconds", "minutes seconds", false, 1, false, "2 minutes 0.0 seconds", "2 minutes 0.0 seconds")
}

func TestDurMathZeroResult(t *testing.T) {
	t.Parallel()
	// a zero result renders as seconds, not nanoseconds
	testDurMath(t, "1 hour", "60 minutes", false, false, "2 hours", "0 seconds")
	testDurMath(t, "1 hour", "60 minutes", true, false, "2h", "0s")
}

func TestDurMathInvalidInput(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, first, second, target string
		decimals                    int
		negative                    bool // true when ErrNegativeDuration is expected
	}{
		{name: "unknown unit in first", first: "1 fortnight", second: "1 hour"},
		{name: "unknown unit in second", first: "1 hour", second: "1 fortnight"},
		{name: "unknown target unit", first: "1 hour", second: "1 hour", target: "bananas"},
		{name: "empty first", first: "", second: "1 hour"},
		{name: "empty second", first: "1 hour", second: ""},
		{name: "whitespace-only target", first: "1 hour", second: "1 hour", target: "   "},
		{name: "bare-dot target", first: "1 hour", second: "1 hour", target: "."},
		{name: "month in first", first: "1 month", second: "1 hour"},
		{name: "month in second", first: "1 hour", second: "1 month"},
		{name: "decimals below range", first: "1 hour", second: "1 hour", target: "hours", decimals: -1},
		{name: "decimals above range", first: "1 hour", second: "1 hour", target: "hours", decimals: 10},
		{name: "missing unit in first", first: "1 hour 2", second: "1 hour"},
		{name: "missing unit in second", first: "1 hour", second: "1 hour 2"},
		{name: "leading negative long first", first: "-1 hour", second: "1 hour", negative: true},
		{name: "leading negative long second", first: "1 hour", second: "-1 hour", negative: true},
		{name: "leading negative brief first", first: "-1h30m", second: "1 hour", negative: true},
		{name: "leading negative brief second", first: "1 hour", second: "-1h30m", negative: true},
		{name: "mid-string negative long first", first: "1 year -30 days", second: "1 hour", negative: true},
		{name: "mid-string negative long second", first: "1 hour", second: "1 year -30 days", negative: true},
		{name: "mid-string negative brief first", first: "1Y-30D", second: "1 hour", negative: true},
		{name: "mid-string negative brief second", first: "1 hour", second: "1Y-30D", negative: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dm := NewDurMath(
				DurMathWithFirst(c.first),
				DurMathWithSecond(c.second),
				DurMathWithTarget(c.target),
				DurMathWithDecimals(c.decimals))
			if _, err := dm.Add(); err == nil {
				t.Errorf("Add: expected an error for first %q second %q target %q decimals %d, got nil", c.first, c.second, c.target, c.decimals)
			} else if c.negative && !errors.Is(err, ErrNegativeDuration) {
				t.Errorf("Add: expected ErrNegativeDuration, got: %v", err)
			}
			if _, err := dm.Sub(); err == nil {
				t.Errorf("Sub: expected an error for first %q second %q target %q decimals %d, got nil", c.first, c.second, c.target, c.decimals)
			} else if c.negative && !errors.Is(err, ErrNegativeDuration) {
				t.Errorf("Sub: expected ErrNegativeDuration, got: %v", err)
			}
		})
	}
}
