package DateTimeMate

import "testing"

func testConv(t *testing.T, source, target string, brief bool, correct string) {
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

func TestConvHoursMinutesSeconds(t *testing.T) {
	source := "386 hours 24 minutes 36 seconds"
	target := "days hours minutes seconds"
	correct := "16 days 2 hours 24 minutes 36 seconds"
	testConv(t, source, target, false, correct)
	correct = "16D2h24m36s"
	testConv(t, source, target, true, correct)

	source = "2 years 26 weeks 15 days 12 hours 30 minutes 30 seconds"
	target = "hours minutes seconds"
	correct = "22272 hours 30 minutes 30 seconds"
	testConv(t, source, target, false, correct)
}

func TestConvSeconds(t *testing.T) {
	source := "1198861 seconds"
	target := "days hours minutes seconds"
	correct := "13 days 21 hours 1 minute 1 second"
	testConv(t, source, target, false, correct)
	correct = "13D21h1m1s"
	testConv(t, source, target, true, correct)

	source = "2 years 26 weeks 15 days 12 hours 30 minutes 30 seconds"
	target = "seconds"
	correct = "80181030 seconds"
	testConv(t, source, target, false, correct)
	correct = "80181030s"
	testConv(t, source, target, true, correct)
}

func TestConvMinutes(t *testing.T) {
	source := "15682 minutes 29 seconds"
	target := "weeks days hours minutes seconds"
	correct := "1 week 3 days 21 hours 22 minutes 29 seconds"
	testConv(t, source, target, false, correct)
	correct = "1W3D21h22m29s"
	testConv(t, source, target, true, correct)
}

func TestConvSingular(t *testing.T) {
	source := "694801 seconds 1 millisecond 1 microsecond 1 nanosecond"
	target := "weeks days hours minutes seconds milliseconds microseconds nanoseconds"
	correct := "1 week 1 day 1 hour 1 second 1 millisecond 1 microsecond 1 nanosecond"
	testConv(t, source, target, false, correct)

	target = "WDhms.msusns"
	correct = "1W1D1h1s1ms1us1ns"
	testConv(t, source, target, true, correct)
}

func TestConvMsUsNs(t *testing.T) {
	source := "4321s123456789ns"
	target := "hms.msusns"
	correct := "1 hour 12 minutes 1 second 123 milliseconds 456 microseconds 788 nanoseconds"
	testConv(t, source, target, false, correct)
	correct = "1h12m1s123ms456us788ns"
	testConv(t, source, target, true, correct)

	source = "4321s001001001ns"
	correct = "1 hour 12 minutes 1 second 1 millisecond 1 microsecond 1 nanosecond"
	testConv(t, source, target, false, correct)
	correct = "1h12m1s1ms1us1ns"
	testConv(t, source, target, true, correct)
}

func TestConvNanoseconds1(t *testing.T) {
	source := "1234567890987654321ns"
	target := "YWDhms.msusns"
	correct := "39 years 6 weeks 2 days 5 hours 31 minutes 30 seconds 987 milliseconds 654 microseconds 447 nanoseconds"
	testConv(t, source, target, false, correct)
	correct = "39Y6W2D5h31m30s987ms654us447ns"
	testConv(t, source, target, true, correct)
}

func TestConvNanoseconds2(t *testing.T) {
	source := "1234567890987654321ns"
	target := "YWDhms.msusns"
	correct := "39 years 6 weeks 2 days 5 hours 31 minutes 30 seconds 987 milliseconds 654 microseconds 447 nanoseconds"
	testConv(t, source, target, false, correct)
	correct = "39Y6W2D5h31m30s987ms654us447ns"
	testConv(t, source, target, true, correct)
}
