package DateTimeMate

import (
	"fmt"
	"github.com/golang-module/carbon/v2"
	"strings"
	"testing"
)

func testStartEnd(t *testing.T, start, end string, brief bool, correct string) {
	diff := NewDiff(
		DiffWithStart(start),
		DiffWithEnd(end),
		DiffWithBrief(brief))
	result, _, err := diff.CalculateDiff()
	if err != nil {
		t.Error(err)
	}
	if result != correct {
		t.Errorf("[computed: %v] != [correct: %v]", result, correct)
	}
}

func testAddSubContains(t *testing.T, from, period, correctAdd, correctSub string) {
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

func testAddSubOutputFormat(t *testing.T, from, period, outputFormat, correctAdd, correctSub string) {
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
	return

	past, err := dur.Sub()
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(past[0], correctSub) {
		t.Errorf("[from: %v] [computed: %v] does not contain: [correct: %v]", from, past, correctSub)
	}
}

func testAddSubWithRepeat(t *testing.T, from, period string, correctAdd, correctSub []string, repeat int) {
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

func testAddUntil(t *testing.T, from, until, periodFuture string, correctAdd []string) {
	durFuture := NewDur(
		DurWithFrom(from),
		DurWithDur(periodFuture),
		DurWithUntil(until))
	future, err := durFuture.Add()
	if err != nil {
		t.Error(err)
	}

	for i := range len(future) {
		if !strings.Contains(future[i], correctAdd[i]) {
			t.Errorf("[from: %v] [computed: %v] does not contain: [correct: %v]", from, future[i], correctAdd[i])
			break
		}
	}
}

func testSubUntil(t *testing.T, from, until, periodPast string, correctSub []string) {
	durPast := NewDur(
		DurWithFrom(from),
		DurWithDur(periodPast),
		DurWithUntil(until))
	past, err := durPast.Sub()
	if err != nil {
		t.Error(err)
	}
	for i := range len(past) {
		if !strings.Contains(past[i], correctSub[i]) {
			t.Errorf("[from: %v] [computed: %v] does not contain: [correct: %v]", from, past[i], correctSub[i])
		}
	}
}

func TestTwoTimesSameDay(t *testing.T) {
	start := "12:00:00"
	end := "15:30:45"
	correct := "3 hours 30 minutes 45 seconds"
	correctBrief := "3h30m45s"
	testStartEnd(t, start, end, false, correct)
	testStartEnd(t, start, end, true, correctBrief)
}

func TestAmPm(t *testing.T) {
	start := "11:00AM"
	end := "11:00PM"
	correct := "12 hours"
	correctBrief := "12h"
	testStartEnd(t, start, end, false, correct)
	testStartEnd(t, start, end, true, correctBrief)
}

func TestIso8601(t *testing.T) {
	start := "2024-06-07T08:00:00Z"
	end := "2024-07-08T09:02:03Z"
	correct := "4 weeks 3 days 1 hour 2 minutes 3 seconds"
	correctBrief := "4W3D1h2m3s"
	testStartEnd(t, start, end, false, correct)
	testStartEnd(t, start, end, true, correctBrief)
}

func TestTimeZoneOffset(t *testing.T) {
	start := "2024-06-07T08:00:00Z"
	end := "2024-06-07T08:05:05-05:00"
	correct := "5 hours 5 minutes 5 seconds"
	correctBrief := "5h5m5s"
	testStartEnd(t, start, end, false, correct)
	testStartEnd(t, start, end, true, correctBrief)
}

func TestIncludeSpaces(t *testing.T) {
	start := "2024-06-07 08:01:02"
	end := "2024-06-07 08:02"
	correct := "58 seconds"
	correctBrief := "58s"
	testStartEnd(t, start, end, false, correct)
	testStartEnd(t, start, end, true, correctBrief)
}

func TestMicroSeconds(t *testing.T) {
	start := "2024-06-07T08:00:00Z"
	end := "2024-06-07T08:00:00.000123Z"
	correct := "123 microseconds"
	correctBrief := "123us"
	testStartEnd(t, start, end, false, correct)
	testStartEnd(t, start, end, true, correctBrief)
}

func TestMilliSeconds(t *testing.T) {
	start := "2024-06-07T08:00:00Z"
	end := "2024-06-07T08:01:02.345Z"
	correct := "1 minute 2 seconds 345 milliseconds"
	correctBrief := "1m2s345ms"
	testStartEnd(t, start, end, false, correct)
	testStartEnd(t, start, end, true, correctBrief)
}

func TestDurationHours(t *testing.T) {
	from := "11:00AM"
	period := "5 hours"
	briefPeriod := "5h"
	correctAdd := " 16:00:00 "
	correctSub := " 06:00:00 "
	testAddSubContains(t, from, period, correctAdd, correctSub)
	testAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	ofmt := " %H:%M:%S "
	testAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurationMillisecondsMicroseconds(t *testing.T) {
	from := "2024-01-01 00:00:00"
	period := "1 minute 2 seconds 123 milliseconds 456 microseconds"
	briefPeriod := "1m2s123ms456us"
	correctAdd := "2024-01-01 00:01:02.123456"
	correctSub := "2023-12-31 23:58:57.876544"
	testAddSubContains(t, from, period, correctAdd, correctSub)
	testAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	correctAdd = correctAdd[:19]
	correctSub = correctSub[:19]
	ofmt := "%Y-%m-%d %H:%M:%S"
	testAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurationHoursMinutesSeconds(t *testing.T) {
	from := "2024-01-01 00:00:00"
	period := "5 hours 5 minutes 5 seconds"
	briefPeriod := "5h5m5s"
	correctAdd := "2024-01-01 05:05:05"
	correctSub := "2023-12-31 18:54:55"
	testAddSubContains(t, from, period, correctAdd, correctSub)
	testAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	ofmt := "%Y-%m-%d %H:%M:%S"
	testAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurationYearsMonthsDays(t *testing.T) {
	from := "2000-01-01"
	period := "5 years 2 months 10 days"
	briefPeriod := "5Y2M10D"
	correctAdd := "2005-03-11"
	correctSub := "1994-10-22"
	testAddSubContains(t, from, period, correctAdd, correctSub)
	testAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	ofmt := "%Y-%m-%d"
	testAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurationYearsMonthsDaysHoursMinutesSeconds(t *testing.T) {
	from := "2024-01-01"
	period := "13 years 8 months 28 days 16 hours 15 minutes 15 seconds"
	briefPeriod := "13Y8M28D16h15m15s"
	correctAdd := "2037-09-29 16:15:15"
	correctSub := "2010-04-02 07:44:45"
	testAddSubContains(t, from, period, correctAdd, correctSub)
	testAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	ofmt := "%Y-%m-%d %H:%M:%S"
	testAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurationWeeksDays(t *testing.T) {
	from := "2024-01-01"
	period := "10 weeks 2 days"
	briefPeriod := "10W2D"
	correctAdd := "2024-03-13"
	correctSub := "2023-10-21"
	testAddSubContains(t, from, period, correctAdd, correctSub)
	testAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	ofmt := "%Y-%m-%d"
	testAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurationMonthsWeeksDays(t *testing.T) {
	from := "2024-06-15"
	period := "2 months 2 weeks 2 days"
	briefPeriod := "2M2W2D"
	correctAdd := "2024-08-31"
	correctSub := "2024-03-30"
	testAddSubContains(t, from, period, correctAdd, correctSub)
	testAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	ofmt := "%Y-%m-%d"
	testAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurationYearsMonthsWeeksDays(t *testing.T) {
	from := "2031-07-12"
	period := "2 years 2 months 2 weeks 2 days"
	briefPeriod := "2Y2M2W2D"
	correctAdd := "2033-09-28"
	correctSub := "2029-04-26"
	testAddSubContains(t, from, period, correctAdd, correctSub)
	testAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	ofmt := "%Y-%m-%d"
	testAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurationNanoseconds(t *testing.T) {
	from := "2031-07-11 05:00:00"
	period := "987654321 nanoseconds"
	briefPeriod := "987654321ns"
	correctAdd := "2031-07-11 05:00:00.987654321"
	correctSub := "2031-07-11 04:59:59.012345679"
	testAddSubContains(t, from, period, correctAdd, correctSub)
	testAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	correctAdd = correctAdd[:19]
	correctSub = correctSub[:19]
	ofmt := "%Y-%m-%d %H:%M:%S"
	testAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestWithRepeat(t *testing.T) {
	from := "2024-06-28T04:25:41Z"
	period := "1M1W1h1m2s"
	repeat := 3
	allCorrectAdd := []string{"2024-08-04 05:26:43", "2024-09-11 06:27:45", "2024-10-18 07:28:47"}
	allCorrectSub := []string{"2024-05-21 03:24:39", "2024-04-14 02:23:37", "2024-03-07 01:22:35"}
	testAddSubWithRepeat(t, from, period, allCorrectAdd, allCorrectSub, repeat)
}

func TestAddUntil(t *testing.T) {
	from := "2024-06-28T04:25:41Z"
	period := "1M1W1h1m2s"
	until := "2024-10-18 07:28:47"
	allCorrectAdd := []string{"2024-08-04 05:26:43", "2024-09-11 06:27:45", "2024-10-18 07:28:47"}
	testAddUntil(t, from, until, period, allCorrectAdd)
}

func TestSubUntil(t *testing.T) {
	from := "2024-10-18 07:28:47"
	period := "1M1W1h1m2s"
	until := "2024-05-28T04:25:41Z"
	allCorrectSub := []string{"2024-09-11 06:27:45", "2024-08-04 05:26:43", "2024-06-27 04:25:41"}
	testSubUntil(t, from, until, period, allCorrectSub)
}

func TestRelativeStartEnd(t *testing.T) {
	start := "yesterday"
	end := "Today"
	correct := "1 day"
	testStartEnd(t, start, end, false, correct)

	start = "Yesterday"
	end = "tomorrow"
	correct = "2 days"
	testStartEnd(t, start, end, false, correct)

	start = "now"
	end = "today"
	correct = "0 seconds"
	testStartEnd(t, start, end, false, correct)

	start = "today"
	end = "tomorrow"
	correct = "1 day"
	testStartEnd(t, start, end, false, correct)
}

func TestRelativeUntil(t *testing.T) {
	from := carbon.Now().StartOfDay().ToDateTimeString()
	period := "7h59m1s"
	until := "tomorrow"
	allCorrectAdd := []string{"", "", "", ""}
	allCorrectAdd[0] = fmt.Sprintf("%s", strings.Replace(from, "00:00:00", "07:59:01", 1))
	allCorrectAdd[1] = fmt.Sprintf("%s", strings.Replace(from, "00:00:00", "15:58:02", 1))
	allCorrectAdd[2] = fmt.Sprintf("%s", strings.Replace(from, "00:00:00", "23:57:03", 1))
	allCorrectAdd[3] = fmt.Sprintf("%s", strings.Replace(from, "00:00:00", "07:56:04", 1))
	allCorrectAdd[3] = fmt.Sprintf("%s", strings.Replace(allCorrectAdd[3], carbon.Now().ToDateString(), carbon.Tomorrow().ToDateString(), 1))
	testAddUntil(t, from, until, period, allCorrectAdd)
}
