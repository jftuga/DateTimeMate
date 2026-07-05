package DateTimeMate

import (
	"github.com/golang-module/carbon/v2"
	"strings"
	"testing"
	"time"
)

func testDurAddSubContains(t *testing.T, from, period, correctAdd, correctSub string) {
	t.Helper()
	dur := NewDur(
		DurWithFrom(from),
		DurWithDur(period))
	future, err := dur.Add()
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(future[0], correctAdd) {
		t.Errorf("[from: %v] [computed: %v] does not contain: [correct: %v]", from, future, correctAdd)
	}

	past, err := dur.Sub()
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(past[0], correctSub) {
		t.Errorf("[from: %v] [computed: %v] does not contain: [correct: %v]", from, past, correctSub)
	}
}

func testDurAddSubOutputFormat(t *testing.T, from, period, outputFormat, correctAdd, correctSub string) {
	t.Helper()
	dur := NewDur(
		DurWithFrom(from),
		DurWithDur(period),
		DurWithOutputFormat(outputFormat))
	future, err := dur.Add()
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(future[0], correctAdd) {
		t.Errorf("\n[from: %v]\n[period: %v]\n[computed: %v] does not contain:\n[correct: %v]", from, period, future[0], correctAdd)
	}

	past, err := dur.Sub()
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(past[0], correctSub) {
		t.Errorf("[from: %v] [computed: %v] does not contain: [correct: %v]", from, past, correctSub)
	}
}

func testDurAddSubWithRepeat(t *testing.T, from, period string, correctAdd, correctSub []string, repeat int) {
	t.Helper()
	dur := NewDur(
		DurWithFrom(from),
		DurWithDur(period),
		DurWithRepeat(repeat))
	future, err := dur.Add()
	if err != nil {
		t.Error(err)
	}
	for i := range len(future) {
		if !strings.Contains(future[i], correctAdd[i]) {
			t.Errorf("[from: %v] [computed: %v] does not contain: [correct: %v]", from, future[i], correctAdd[i])
		}
	}

	past, err := dur.Sub()
	if err != nil {
		t.Error(err)
	}
	for j := range len(past) {
		if !strings.Contains(past[j], correctSub[j]) {
			t.Errorf("[from: %v] [computed: %v] does not contain: [correct: %v]", from, past[j], correctSub[j])
		}
	}
}

func testDurAddUntil(t *testing.T, from, until, periodFuture string, correctAdd []string) {
	t.Helper()
	durFuture := NewDur(
		DurWithFrom(from),
		DurWithDur(periodFuture),
		DurWithUntil(until))
	future, err := durFuture.Add()
	if err != nil {
		t.Error(err)
	}
	if len(future) != len(correctAdd) {
		t.Fatalf("[from: %v] expected %d results, got %d: %v", from, len(correctAdd), len(future), future)
	}

	for i := range len(future) {
		if !strings.Contains(future[i], correctAdd[i]) {
			t.Errorf("[from: %v] [computed: %v] does not contain: [correct: %v]", from, future[i], correctAdd[i])
			break
		}
	}
}

func testDurSubUntil(t *testing.T, from, until, periodPast string, correctSub []string) {
	t.Helper()
	durPast := NewDur(
		DurWithFrom(from),
		DurWithDur(periodPast),
		DurWithUntil(until))
	past, err := durPast.Sub()
	if err != nil {
		t.Error(err)
	}
	if len(past) != len(correctSub) {
		t.Fatalf("[from: %v] expected %d results, got %d: %v", from, len(correctSub), len(past), past)
	}
	for i := range len(past) {
		if !strings.Contains(past[i], correctSub[i]) {
			t.Errorf("[from: %v] [computed: %v] does not contain: [correct: %v]", from, past[i], correctSub[i])
		}
	}
}

func TestDurUnixTimestampFrom(t *testing.T) {
	t.Parallel()
	// a 10-digit integer from-value is a Unix timestamp in seconds; it used
	// to be silently misparsed by parsetime as a time-of-day on the current
	// date; "%s" output keeps the assertions timezone-independent
	testDurAddSubOutputFormat(t, "1700265600", "1 day", "%s", "1700352000", "1700179200")
	// a 13-digit integer is a Unix timestamp in milliseconds
	testDurAddSubOutputFormat(t, "1700265600123", "1 day", "%s", "1700352000", "1700179200")
}

func TestDurAmbiguousTimestamp(t *testing.T) {
	t.Parallel()
	// 11 and 12 digit integers are neither seconds (10) nor milliseconds (13)
	// and previously fell through to parsetime where they were misparsed
	for _, from := range []string{"17002656001", "170026560012"} {
		dur := NewDur(DurWithFrom(from), DurWithDur("1 day"))
		if _, err := dur.Add(); err == nil {
			t.Errorf("expected an error for ambiguous timestamp %q, got nil", from)
		}
	}
}

func TestDurUnixTimestampUntil(t *testing.T) {
	t.Parallel()
	dur := NewDur(
		DurWithFrom("1700265600"),
		DurWithDur("1 day"),
		DurWithUntil("1700438400"),
		DurWithOutputFormat("%s"))
	future, err := dur.Add()
	if err != nil {
		t.Fatal(err)
	}
	correct := []string{"1700352000", "1700438400"}
	if len(future) != len(correct) {
		t.Fatalf("[computed: %v] != [correct: %v]", future, correct)
	}
	for i := range correct {
		if future[i] != correct[i] {
			t.Errorf("[computed: %v] != [correct: %v]", future[i], correct[i])
		}
	}
}

func TestDurFractionalPeriod(t *testing.T) {
	t.Parallel()
	from := "2024-01-01 00:00:00"
	period := "1.5 hours"
	briefPeriod := "1.5h"
	correctAdd := "2024-01-01 01:30:00"
	correctSub := "2023-12-31 22:30:00"
	testDurAddSubContains(t, from, period, correctAdd, correctSub)
	testDurAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
}

func TestDurInvalidPeriodText(t *testing.T) {
	t.Parallel()
	// no part of a period may be silently ignored
	for _, period := range []string{"1 hour 2", "1 hour 2m", "1 hour bananas", "1.5.2 hours"} {
		dur := NewDur(DurWithFrom("2024-01-01"), DurWithDur(period))
		if _, err := dur.Add(); err == nil {
			t.Errorf("expected an error for period %q, got nil", period)
		}
	}
}

func TestDurNegativeRepeat(t *testing.T) {
	t.Parallel()
	dur := NewDur(DurWithFrom("2024-01-01"), DurWithDur("1 hour"), DurWithRepeat(-1))
	if _, err := dur.Add(); err == nil {
		t.Error("expected an error for a negative repeat, got nil")
	}
}

func TestDurZeroPeriodUntil(t *testing.T) {
	t.Parallel()
	// a period that does not advance must error instead of looping forever
	dur := NewDur(DurWithFrom("2024-01-01"), DurWithDur("0 minutes"), DurWithUntil("2024-01-02"))
	if _, err := dur.Add(); err == nil {
		t.Error("expected an error for a zero-length period with until, got nil")
	}
}

func TestDurSubSecondRepeat(t *testing.T) {
	t.Parallel()
	// sub-second precision must survive repeated application
	from := "2024-01-01 00:00:00"
	period := "500ms"
	repeat := 3
	allCorrectAdd := []string{"2024-01-01 00:00:00.5 ", "2024-01-01 00:00:01 ", "2024-01-01 00:00:01.5 "}
	allCorrectSub := []string{"2023-12-31 23:59:59.5 ", "2023-12-31 23:59:59 ", "2023-12-31 23:59:58.5 "}
	testDurAddSubWithRepeat(t, from, period, allCorrectAdd, allCorrectSub, repeat)
}

func TestDurWinterTimezoneOffset(t *testing.T) {
	t.Parallel()
	// a zone-less date must resolve with the local UTC offset in effect on
	// that date, not the offset in effect today
	dur := NewDur(DurWithFrom("2024-01-15 12:00:00"), DurWithDur("1 hour"))
	future, err := dur.Add()
	if err != nil {
		t.Fatal(err)
	}
	computed, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", future[0])
	if err != nil {
		t.Fatal(err)
	}
	correct, err := time.ParseInLocation("2006-01-02 15:04:05", "2024-01-15 13:00:00", time.Local)
	if err != nil {
		t.Fatal(err)
	}
	if !computed.Equal(correct) {
		t.Errorf("[computed: %v] != [correct: %v]", computed, correct)
	}
}

func TestDurHours(t *testing.T) {
	t.Parallel()
	from := "11:00AM"
	period := "5 hours"
	briefPeriod := "5h"
	correctAdd := " 16:00:00 "
	correctSub := " 06:00:00 "
	testDurAddSubContains(t, from, period, correctAdd, correctSub)
	testDurAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	ofmt := " %H:%M:%S "
	testDurAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurMillisecondsMicroseconds(t *testing.T) {
	t.Parallel()
	from := "2024-01-01 00:00:00"
	period := "1 minute 2 seconds 123 milliseconds 456 microseconds"
	briefPeriod := "1m2s123ms456us"
	correctAdd := "2024-01-01 00:01:02.123456"
	correctSub := "2023-12-31 23:58:57.876544"
	testDurAddSubContains(t, from, period, correctAdd, correctSub)
	testDurAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	correctAdd = correctAdd[:19]
	correctSub = correctSub[:19]
	ofmt := "%Y-%m-%d %H:%M:%S"
	testDurAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurHoursMinutesSeconds(t *testing.T) {
	t.Parallel()
	from := "2024-01-01 00:00:00"
	period := "5 hours 5 minutes 5 seconds"
	briefPeriod := "5h5m5s"
	correctAdd := "2024-01-01 05:05:05"
	correctSub := "2023-12-31 18:54:55"
	testDurAddSubContains(t, from, period, correctAdd, correctSub)
	testDurAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	ofmt := "%Y-%m-%d %H:%M:%S"
	testDurAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurYearsDays(t *testing.T) {
	t.Parallel()
	from := "2000-01-01"
	period := "5 years 70 days"
	briefPeriod := "5Y70D"
	correctAdd := "2005-03-12"
	correctSub := "1994-10-23"
	testDurAddSubContains(t, from, period, correctAdd, correctSub)
	testDurAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	ofmt := "%Y-%m-%d"
	testDurAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurYearsDaysHoursMinutesSeconds(t *testing.T) {
	t.Parallel()
	from := "2024-01-01"
	period := "13 years 272 days 16 hours 15 minutes 15 seconds"
	briefPeriod := "13Y272D16h15m15s"
	correctAdd := "2037-09-30 16:15:15"
	correctSub := "2010-04-03 07:44:45"
	testDurAddSubContains(t, from, period, correctAdd, correctSub)
	testDurAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	ofmt := "%Y-%m-%d %H:%M:%S"
	testDurAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurWeeksDays1(t *testing.T) {
	t.Parallel()
	from := "2024-01-01"
	period := "10 weeks 2 days"
	briefPeriod := "10W2D"
	correctAdd := "2024-03-13"
	correctSub := "2023-10-21"
	testDurAddSubContains(t, from, period, correctAdd, correctSub)
	testDurAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	ofmt := "%Y-%m-%d"
	testDurAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurWeeksDays2(t *testing.T) {
	t.Parallel()
	from := "2024-06-15"
	period := "11 weeks 2 days"
	briefPeriod := "11W2D"
	correctAdd := "2024-09-02"
	correctSub := "2024-03-28"
	testDurAddSubContains(t, from, period, correctAdd, correctSub)
	testDurAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	ofmt := "%Y-%m-%d"
	testDurAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurYearsWeeksDays(t *testing.T) {
	t.Parallel()
	from := "2031-07-12"
	period := "2 years 11 weeks 2 days"
	briefPeriod := "2Y11W2D"
	correctAdd := "2033-09-29"
	correctSub := "2029-04-24"
	testDurAddSubContains(t, from, period, correctAdd, correctSub)
	testDurAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	ofmt := "%Y-%m-%d"
	testDurAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurNanoseconds(t *testing.T) {
	t.Parallel()
	from := "2031-07-11 05:00:00"
	period := "987654321 nanoseconds"
	briefPeriod := "987654321ns"
	correctAdd := "2031-07-11 05:00:00.987654321"
	correctSub := "2031-07-11 04:59:59.012345679"
	testDurAddSubContains(t, from, period, correctAdd, correctSub)
	testDurAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	correctAdd = correctAdd[:19]
	correctSub = correctSub[:19]
	ofmt := "%Y-%m-%d %H:%M:%S"
	testDurAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurWithRepeat(t *testing.T) {
	t.Parallel()
	from := "2024-06-28T04:25:41Z"
	period := "5W1h1m2s"
	repeat := 3
	allCorrectAdd := []string{"2024-08-02 05:26:43", "2024-09-06 06:27:45", "2024-10-11 07:28:47"}
	allCorrectSub := []string{"2024-05-24 03:24:39", "2024-04-19 02:23:37", "2024-03-15 01:22:35"}
	testDurAddSubWithRepeat(t, from, period, allCorrectAdd, allCorrectSub, repeat)
}

func TestDurAddUntil(t *testing.T) {
	t.Parallel()
	from := "2024-06-28T04:25:41Z"
	period := "5W1h1m2s"
	until := "2024-10-11 07:28:47"
	allCorrectAdd := []string{"2024-08-02 05:26:43", "2024-09-06 06:27:45", "2024-10-11 07:28:47"}
	testDurAddUntil(t, from, until, period, allCorrectAdd)
}

func TestDurSubUntil(t *testing.T) {
	t.Parallel()
	from := "2024-10-18 07:28:47"
	period := "5W1h1m2s"
	until := "2024-06-28T04:25:41Z"
	allCorrectSub := []string{"2024-09-13 06:27:45", "2024-08-09 05:26:43", "2024-07-05 04:25:41"}
	testDurSubUntil(t, from, until, period, allCorrectSub)
}

func TestDurPre1970(t *testing.T) {
	t.Parallel()
	// entirely before 1970: parsetime silently corrupted these dates
	// before parseDateTime tried standard layouts first
	testDurAddSubContains(t, "1950-01-01 12:00:00", "1 day 2 hours", "1950-01-02 14:00:00", "1949-12-31 10:00:00")
	// crossing the unix epoch in both directions
	testDurAddSubContains(t, "1970-01-01 00:00:00", "1h", "1970-01-01 01:00:00", "1969-12-31 23:00:00")
}

func TestDurPre1970Repeat(t *testing.T) {
	t.Parallel()
	from := "1969-12-31 22:00:00"
	period := "1h"
	repeat := 3
	allCorrectAdd := []string{"1969-12-31 23:00:00", "1970-01-01 00:00:00", "1970-01-01 01:00:00"}
	allCorrectSub := []string{"1969-12-31 21:00:00", "1969-12-31 20:00:00", "1969-12-31 19:00:00"}
	testDurAddSubWithRepeat(t, from, period, allCorrectAdd, allCorrectSub, repeat)
}

func TestDurPre1970Until(t *testing.T) {
	t.Parallel()
	from := "1970-01-01 02:00:00"
	period := "1h"
	until := "1969-12-31 22:30:00"
	allCorrectSub := []string{"1970-01-01 01:00:00", "1970-01-01 00:00:00", "1969-12-31 23:00:00"}
	testDurSubUntil(t, from, until, period, allCorrectSub)
}

func TestDurRelativeUntil(t *testing.T) {
	t.Parallel()
	start := carbon.Now().StartOfDay()
	from := start.ToDateTimeString()
	dur := NewDur(
		DurWithFrom(from),
		DurWithDur("7h59m1s"),
		DurWithUntil("tomorrow"))
	future, err := dur.Add()
	if err != nil {
		t.Fatal(err)
	}
	// "tomorrow" resolves to now+24h, so the span covered from start of day
	// grows with the wall clock; the result count varies with the time of
	// day, but each entry is deterministic: start of day plus i+1 periods
	if len(future) < 3 || len(future) > 6 {
		t.Fatalf("[from: %v] expected 3 to 6 results, got %d: %v", from, len(future), future)
	}
	period := 7*time.Hour + 59*time.Minute + 1*time.Second
	for i := range len(future) {
		correct := start.StdTime().Add(time.Duration(i+1) * period).Format("2006-01-02 15:04:05")
		if !strings.Contains(future[i], correct) {
			t.Errorf("[from: %v] [computed: %v] does not contain: [correct: %v]", from, future[i], correct)
		}
	}
}
