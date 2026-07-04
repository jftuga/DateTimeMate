package DateTimeMate

import "testing"

func testConv(t *testing.T, source, target string, brief bool, correct string) {
	t.Helper()
	conv := NewConv(
		ConvWithSource(source),
		ConvWithTarget(target),
		ConvWithBrief(brief))

	result, err := conv.ConvertDuration()
	if err != nil {
		t.Error(err)
	}
	if result != correct {
		t.Errorf("\n[computed: %v] !=\n[correct : %v]", result, correct)
	}
}

func testConvDecimals(t *testing.T, source, target string, brief bool, decimals int, correct string) {
	t.Helper()
	conv := NewConv(
		ConvWithSource(source),
		ConvWithTarget(target),
		ConvWithBrief(brief),
		ConvWithDecimals(decimals))

	result, err := conv.ConvertDuration()
	if err != nil {
		t.Error(err)
	}
	if result != correct {
		t.Errorf("\n[computed: %v] !=\n[correct : %v]", result, correct)
	}
}

func TestConvDecimals(t *testing.T) {
	t.Parallel()
	// the example from issue #31
	source := "2 years 37 weeks 2 days"
	target := "years"
	correct := "2.71 years"
	testConvDecimals(t, source, target, false, 2, correct)

	correct = "2.71Y"
	testConvDecimals(t, source, target, true, 2, correct)

	source = "-2 years 37 weeks 2 days"
	correct = "-2.71 years"
	testConvDecimals(t, source, target, false, 2, correct)

	// decimals only apply to the last (smallest) unit
	source = "2 years 37 weeks 2 days"
	target = "years weeks"
	correct = "2 years 37.29 weeks"
	testConvDecimals(t, source, target, false, 2, correct)

	source = "1 hour 30 minutes"
	target = "hours"
	correct = "1.5 hours"
	testConvDecimals(t, source, target, false, 1, correct)

	// a value rounding to exactly one is singular
	source = "365 days 6 hours"
	target = "years"
	correct = "1.00 year"
	testConvDecimals(t, source, target, false, 2, correct)

	// the last unit is included even when less than one
	source = "3 days"
	target = "years"
	correct = "0.01 years"
	testConvDecimals(t, source, target, false, 2, correct)

	// zero decimals preserves the default truncation behavior
	source = "2 years 37 weeks 2 days"
	target = "years"
	correct = "2 years"
	testConvDecimals(t, source, target, false, 0, correct)
}

func TestConvDecimalsOutOfRange(t *testing.T) {
	t.Parallel()
	for _, decimals := range []int{-1, 10} {
		conv := NewConv(
			ConvWithSource("90 minutes"),
			ConvWithTarget("hours"),
			ConvWithDecimals(decimals))
		_, err := conv.ConvertDuration()
		if err == nil {
			t.Errorf("expected an error for decimals=%d, got nil", decimals)
		}
	}
}

func TestConvInvalidInput(t *testing.T) {
	t.Parallel()
	cases := []struct{ source, target string }{
		{"1 hour 2", "hours"},           // dangling amount with no unit (used to panic)
		{"1 fortnight", "hours"},        // unknown source unit (used to output nothing)
		{"90 minutes", "hours bananas"}, // unknown target unit (used to divide by zero)
		{"90 minutes", "fortnights"},    // unknown single target unit
		{"1 month", "days"},             // months are deliberately unsupported; lengths vary
		{"", "hours"},                   // empty source
		{"abc days", "hours"},           // invalid numeric amount
		{"15", "hours"},                 // bare number with no unit
	}
	for _, c := range cases {
		conv := NewConv(ConvWithSource(c.source), ConvWithTarget(c.target))
		if _, err := conv.ConvertDuration(); err == nil {
			t.Errorf("expected an error for source %q target %q, got nil", c.source, c.target)
		}
	}
}

func TestConvEmptyTarget(t *testing.T) {
	t.Parallel()
	// an empty or whitespace-only target used to panic with an
	// index-out-of-range error instead of returning an error
	for _, target := range []string{"", "   "} {
		conv := NewConv(ConvWithSource("90 minutes"), ConvWithTarget(target))
		if _, err := conv.ConvertDuration(); err == nil {
			t.Errorf("expected an error for empty target %q, got nil", target)
		}
	}
}

func TestConvCaseInsensitiveUnits(t *testing.T) {
	t.Parallel()
	// long-form units are case-insensitive; uppercase plurals such as
	// "DAYS" used to fail because the trailing "S" was not stripped
	testConv(t, "1 DAYS", "hours", false, "24 hours")
	testConv(t, "1 Week 2 DAYS", "days", false, "9 days")
	testConv(t, "1 DaY 2 houRS", "hours", false, "26 hours")
	testConv(t, "90 minutes", "HOURS", false, "1 hour")
}

func TestConvMixedSignSource(t *testing.T) {
	t.Parallel()
	// a negative amount mid-string subtracts from the total
	testConv(t, "1 year -30 days", "days", false, "335 days")
}

func TestConvZeroResult(t *testing.T) {
	t.Parallel()
	// a result that truncates to zero emits zero of the smallest unit
	// instead of an empty string
	testConv(t, "3599 seconds", "hours", false, "0 hours")
	testConv(t, "3599 seconds", "hours", true, "0h")
	testConv(t, "0 seconds", "days hours", false, "0 hours")
}

func TestConvBriefWithSpaces(t *testing.T) {
	t.Parallel()
	testConv(t, "1h 30m", "minutes", false, "90 minutes")
	testConv(t, "1Y 3W 4D", "days", false, "390 days")
}

func TestConvHoursMinutesSeconds(t *testing.T) {
	t.Parallel()
	source := "386 hours 24 minutes 36 seconds"
	target := "days hours minutes seconds"
	correct := "16 days 2 hours 24 minutes 36 seconds"
	testConv(t, source, target, false, correct)

	correct = "16D2h24m36s"
	testConv(t, source, target, true, correct)

	source = "-386 hours 24 minutes 36 seconds"
	correct = "-16 days 2 hours 24 minutes 36 seconds"
	testConv(t, source, target, false, correct)

	correct = "-16D2h24m36s"
	testConv(t, source, target, true, correct)

	source = "2 years 26 weeks 15 days 12 hours 30 minutes 30 seconds"
	target = "hours minutes seconds"
	correct = "22272 hours 30 minutes 30 seconds"
	testConv(t, source, target, false, correct)

	source = "-2 years 26 weeks 15 days 12 hours 30 minutes 30 seconds"
	correct = "-22272 hours 30 minutes 30 seconds"
	testConv(t, source, target, false, correct)
}

func TestConvSeconds(t *testing.T) {
	t.Parallel()
	source := "1198861 seconds"
	target := "days hours minutes seconds"
	correct := "13 days 21 hours 1 minute 1 second"
	testConv(t, source, target, false, correct)

	correct = "13D21h1m1s"
	testConv(t, source, target, true, correct)

	source = "-1198861 seconds"
	correct = "-13 days 21 hours 1 minute 1 second"
	testConv(t, source, target, false, correct)

	correct = "-13D21h1m1s"
	testConv(t, source, target, true, correct)

	source = "2 years 26 weeks 15 days 12 hours 30 minutes 30 seconds"
	target = "seconds"
	correct = "80181030 seconds"
	testConv(t, source, target, false, correct)

	correct = "80181030s"
	testConv(t, source, target, true, correct)

	source = "-2 years 26 weeks 15 days 12 hours 30 minutes 30 seconds"
	correct = "-80181030 seconds"
	testConv(t, source, target, false, correct)

	correct = "-80181030s"
	testConv(t, source, target, true, correct)
}

func TestConvMinutes(t *testing.T) {
	t.Parallel()
	source := "15682 minutes 29 seconds"
	target := "weeks days hours minutes seconds"
	correct := "1 week 3 days 21 hours 22 minutes 29 seconds"
	testConv(t, source, target, false, correct)

	source = "-15682 minutes 29 seconds"
	correct = "-1 week 3 days 21 hours 22 minutes 29 seconds"
	testConv(t, source, target, false, correct)

	source = "15682 minutes 29 seconds"
	correct = "1W3D21h22m29s"
	testConv(t, source, target, true, correct)

	source = "-15682 minutes 29 seconds"
	correct = "-1W3D21h22m29s"
	testConv(t, source, target, true, correct)
}

func TestConvSingular(t *testing.T) {
	t.Parallel()
	source := "694801 seconds 1 millisecond 1 microsecond 1 nanosecond"
	target := "weeks days hours minutes seconds milliseconds microseconds nanoseconds"
	correct := "1 week 1 day 1 hour 1 second 1 millisecond 1 microsecond 1 nanosecond"
	testConv(t, source, target, false, correct)

	source = "-694801 seconds 1 millisecond 1 microsecond 1 nanosecond"
	correct = "-1 week 1 day 1 hour 1 second 1 millisecond 1 microsecond 1 nanosecond"
	testConv(t, source, target, false, correct)

	source = "694801 seconds 1 millisecond 1 microsecond 1 nanosecond"
	target = "WDhms.msusns"
	correct = "1W1D1h1s1ms1us1ns"
	testConv(t, source, target, true, correct)

	source = "-694801 seconds 1 millisecond 1 microsecond 1 nanosecond"
	correct = "-1W1D1h1s1ms1us1ns"
	testConv(t, source, target, true, correct)
}

func TestConvMsUsNs(t *testing.T) {
	t.Parallel()
	source := "4321s123456789ns"
	target := "hms.msusns"
	correct := "1 hour 12 minutes 1 second 123 milliseconds 456 microseconds 788 nanoseconds"
	testConv(t, source, target, false, correct)

	source = "-4321s123456789ns"
	correct = "-1 hour 12 minutes 1 second 123 milliseconds 456 microseconds 788 nanoseconds"
	testConv(t, source, target, false, correct)

	source = "4321s123456789ns"
	correct = "1h12m1s123ms456us788ns"
	testConv(t, source, target, true, correct)

	source = "-4321s123456789ns"
	correct = "-1h12m1s123ms456us788ns"
	testConv(t, source, target, true, correct)

	source = "4321s001001001ns"
	correct = "1 hour 12 minutes 1 second 1 millisecond 1 microsecond 1 nanosecond"
	testConv(t, source, target, false, correct)

	source = "-4321s001001001ns"
	correct = "-1 hour 12 minutes 1 second 1 millisecond 1 microsecond 1 nanosecond"
	testConv(t, source, target, false, correct)

	source = "4321s001001001ns"
	correct = "1h12m1s1ms1us1ns"
	testConv(t, source, target, true, correct)

	source = "-4321s001001001ns"
	correct = "-1h12m1s1ms1us1ns"
	testConv(t, source, target, true, correct)
}

func TestConvNanoseconds1(t *testing.T) {
	t.Parallel()
	source := "1234567890987654321ns"
	target := "YWDhms.msusns"
	correct := "39 years 6 weeks 2 days 5 hours 31 minutes 30 seconds 987 milliseconds 654 microseconds 447 nanoseconds"
	testConv(t, source, target, false, correct)

	source = "-1234567890987654321ns"
	correct = "-39 years 6 weeks 2 days 5 hours 31 minutes 30 seconds 987 milliseconds 654 microseconds 447 nanoseconds"
	testConv(t, source, target, false, correct)

	source = "1234567890987654321ns"
	correct = "39Y6W2D5h31m30s987ms654us447ns"
	testConv(t, source, target, true, correct)

	source = "-1234567890987654321ns"
	correct = "-39Y6W2D5h31m30s987ms654us447ns"
	testConv(t, source, target, true, correct)
}
