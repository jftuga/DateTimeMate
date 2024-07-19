package DateTimeMate

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	secondsPerNanoseconds = 0.000000001
	secondsPerMicrosecond = 0.000001
	secondsPerMillisecond = 0.001
	secondsPerMinute      = 60
	secondsPerHour        = 60 * secondsPerMinute
	secondsPerDay         = 24 * secondsPerHour
	secondsPerWeek        = 7 * secondsPerDay
	secondsPerYear        = 365.25 * secondsPerDay
)

var unitMap = map[string]float64{
	"nanosecond":  secondsPerNanoseconds,
	"microsecond": secondsPerMicrosecond,
	"millisecond": secondsPerMillisecond,
	"second":      1,
	"minute":      secondsPerMinute,
	"hour":        secondsPerHour,
	"day":         secondsPerDay,
	"week":        secondsPerWeek,
	"year":        secondsPerYear,
}

var unitBriefMap = map[string]string{
	"ns": "nanosecond",
	"us": "microsecond",
	"ms": "millisecond",
	"s":  "second",
	"m":  "minute",
	"h":  "hour",
	"D":  "day",
	"W":  "week",
	"Y":  "year",
}

type Conv struct {
	Source string
	Target string
	Brief  bool
}

type OptionsConv func(*Conv)

func NewConv(options ...OptionsConv) *Conv {
	conv := &Conv{}
	for _, opt := range options {
		opt(conv)
	}
	return conv
}

func ConvWithSource(source string) OptionsConv {
	return func(conv *Conv) {
		conv.Source = source
	}
}

func ConvWithTarget(target string) OptionsConv {
	return func(conv *Conv) {
		conv.Target = target
	}
}

func ConvWithBrief(brief bool) OptionsConv {
	return func(conv *Conv) {
		conv.Brief = brief
	}
}

func (conv *Conv) parseSource() (float64, error) {
	parts := strings.Fields(conv.Source)
	totalSeconds := 0.0

	for i := 0; i < len(parts); i += 2 {
		value, err := strconv.ParseFloat(parts[i], 64)
		if err != nil {
			return 0, err
		}
		unit := strings.ToLower(strings.TrimSuffix(parts[i+1], "s")) // Remove plural 's'
		if seconds, ok := unitMap[unit]; ok {
			totalSeconds += value * seconds
		}
	}
	return totalSeconds, nil
}

func (conv *Conv) formatTarget(seconds float64, units []string) string {
	result := ""
	for _, unit := range units {
		unitInSeconds := unitMap[strings.TrimSuffix(unit, "s")]
		value := seconds / unitInSeconds
		intValue := int(value)
		if intValue > 0 {
			if intValue == 1 {
				unit = removeTrailingS(unit)
			} else if unit[len(unit)-1] != 's' {
				unit += "s"
			}
			result += fmt.Sprintf("%d %s ", intValue, unit)
			seconds -= float64(intValue) * unitInSeconds
		}
	}
	return strings.TrimSpace(result)
}

// expandBriefSourceDuration expands a brief period into a long period format
func expandBriefSourceDuration(period string) (string, error) {
	var err error

	// brief format is being used so first expand it to the long format
	period, err = expandPeriod(period)
	if nil != err {
		return "", fmt.Errorf("%v", err)
	}
	periodMatches := expandedRegexp.FindAllStringSubmatch(period, -1)
	if len(periodMatches) == 0 {
		return "", fmt.Errorf("[expandBriefSourceDuration] Invalid duration: %s", period)
	}
	return period, nil
}

// expandBriefTargetDuration give a brief duration, convert to long format
// example: WDhms.msusns => "week day hour minute second millisecond microsecond nanosecond"
func expandBriefTargetDuration(period string) ([]string, error) {
	usingSubSeconds := false
	var result []string
	var i int
	var ch rune
	for i, ch = range period {
		if ch == '.' {
			usingSubSeconds = true
			break
		}
		longName, ok := unitBriefMap[string(ch)]
		if !ok {
			return nil, fmt.Errorf("[expandBriefTargetDuration] Invalid time duration: %c", ch)
		}
		result = append(result, longName)
	}
	if usingSubSeconds {
		i += 1
		period := period[i:] // move past the dot
		if len(period)%2 != 0 {
			return nil, fmt.Errorf("[expandBriefTargetDuration] Invalid sub-second duration: %s", period)
		}
		for j := 0; j < len(period); j += 2 {
			subSecPeriod := period[j : j+2]
			longName, ok := unitBriefMap[subSecPeriod]
			if !ok {
				return nil, fmt.Errorf("[expandBriefTargetDuration] Invalid duration: %c", ch)
			}
			result = append(result, longName)
		}
	}
	return result, nil
}

func (conv *Conv) ConvertDuration() (string, error) {
	var err error
	fields := strings.Fields(conv.Source)
	if len(fields) == 1 {
		// brief format is being used so convert to long duration format
		conv.Source, err = expandBriefSourceDuration(conv.Source)
		if nil != err {
			return "", err
		}
	}
	seconds, err := conv.parseSource()
	if err != nil {
		return "", err
	}
	targetUnits := strings.Fields(conv.Target)
	if len(targetUnits) == 1 {
		_, ok := unitMap[removeTrailingS(conv.Target)]
		if !ok {
			// brief format is being used so convert to long duration format
			targetUnits, err = expandBriefTargetDuration(conv.Target)
			if nil != err {
				return "", err
			}
		}
	}
	result := conv.formatTarget(seconds, targetUnits)
	if conv.Brief {
		result = shrinkPeriod(result)
	}
	return result, nil
}
