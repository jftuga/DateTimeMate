package DateTimeMate

import (
	"fmt"
	"github.com/golang-module/carbon/v2"
	"strings"
	"testing"
)

func testDurAddSubContains(t *testing.T, from, period, correctAdd, correctSub string) {
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

func testDurAddSubWithRepeat(t *testing.T, from, period string, correctAdd, correctSub []string, repeat int) {
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

func testDurSubUntil(t *testing.T, from, until, periodPast string, correctSub []string) {
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

func TestDurHours(t *testing.T) {
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

func TestDurYearsMonthsDays(t *testing.T) {
	from := "2000-01-01"
	period := "5 years 2 months 10 days"
	briefPeriod := "5Y2M10D"
	correctAdd := "2005-03-11"
	correctSub := "1994-10-22"
	testDurAddSubContains(t, from, period, correctAdd, correctSub)
	testDurAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	ofmt := "%Y-%m-%d"
	testDurAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurYearsMonthsDaysHoursMinutesSeconds(t *testing.T) {
	from := "2024-01-01"
	period := "13 years 8 months 28 days 16 hours 15 minutes 15 seconds"
	briefPeriod := "13Y8M28D16h15m15s"
	correctAdd := "2037-09-29 16:15:15"
	correctSub := "2010-04-02 07:44:45"
	testDurAddSubContains(t, from, period, correctAdd, correctSub)
	testDurAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	ofmt := "%Y-%m-%d %H:%M:%S"
	testDurAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurWeeksDays(t *testing.T) {
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

func TestDurMonthsWeeksDays(t *testing.T) {
	from := "2024-06-15"
	period := "2 months 2 weeks 2 days"
	briefPeriod := "2M2W2D"
	correctAdd := "2024-08-31"
	correctSub := "2024-03-30"
	testDurAddSubContains(t, from, period, correctAdd, correctSub)
	testDurAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	ofmt := "%Y-%m-%d"
	testDurAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurYearsMonthsWeeksDays(t *testing.T) {
	from := "2031-07-12"
	period := "2 years 2 months 2 weeks 2 days"
	briefPeriod := "2Y2M2W2D"
	correctAdd := "2033-09-28"
	correctSub := "2029-04-26"
	testDurAddSubContains(t, from, period, correctAdd, correctSub)
	testDurAddSubContains(t, from, briefPeriod, correctAdd, correctSub)
	ofmt := "%Y-%m-%d"
	testDurAddSubOutputFormat(t, from, period, ofmt, correctAdd, correctSub)
}

func TestDurNanoseconds(t *testing.T) {
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
	from := "2024-06-28T04:25:41Z"
	period := "1M1W1h1m2s"
	repeat := 3
	allCorrectAdd := []string{"2024-08-04 05:26:43", "2024-09-11 06:27:45", "2024-10-18 07:28:47"}
	allCorrectSub := []string{"2024-05-21 03:24:39", "2024-04-14 02:23:37", "2024-03-07 01:22:35"}
	testDurAddSubWithRepeat(t, from, period, allCorrectAdd, allCorrectSub, repeat)
}

func TestDurAddUntil(t *testing.T) {
	from := "2024-06-28T04:25:41Z"
	period := "1M1W1h1m2s"
	until := "2024-10-18 07:28:47"
	allCorrectAdd := []string{"2024-08-04 05:26:43", "2024-09-11 06:27:45", "2024-10-18 07:28:47"}
	testDurAddUntil(t, from, until, period, allCorrectAdd)
}

func TestDurSubUntil(t *testing.T) {
	from := "2024-10-18 07:28:47"
	period := "1M1W1h1m2s"
	until := "2024-05-28T04:25:41Z"
	allCorrectSub := []string{"2024-09-11 06:27:45", "2024-08-04 05:26:43", "2024-06-27 04:25:41"}
	testDurSubUntil(t, from, until, period, allCorrectSub)
}

func TestDurRelativeUntil(t *testing.T) {
	from := carbon.Now().StartOfDay().ToDateTimeString()
	period := "7h59m1s"
	until := "tomorrow"
	allCorrectAdd := []string{"", "", "", "", ""}
	allCorrectAdd[0] = fmt.Sprintf("%s", strings.Replace(from, "00:00:00", "07:59:01", 1))
	allCorrectAdd[1] = fmt.Sprintf("%s", strings.Replace(from, "00:00:00", "15:58:02", 1))
	allCorrectAdd[2] = fmt.Sprintf("%s", strings.Replace(from, "00:00:00", "23:57:03", 1))
	allCorrectAdd[3] = fmt.Sprintf("%s", strings.Replace(from, "00:00:00", "07:56:04", 1))
	allCorrectAdd[3] = fmt.Sprintf("%s", strings.Replace(allCorrectAdd[3], carbon.Now().ToDateString(), carbon.Tomorrow().ToDateString(), 1))
	testDurAddUntil(t, from, until, period, allCorrectAdd)
}
