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

func (conv *Conv) String() string {
	return fmt.Sprintf("Source:%v Target:%v Brief:%v", conv.Source, conv.Target, conv.Brief)
}

// parseSource parses the Source field of the Conv struct
// and converts it to a total number of seconds.
//
// The function expects the Source string to contain alternating numeric values and time unit strings
// (e.g., "1 hour 30 minutes"). It supports various time units (defined in unitMap) in both singular
// and plural forms.
//
// The function iterates through the Source string, parsing each numeric value and its corresponding
// time unit. It then converts each time value to seconds and accumulates the total.
//
// Returns:
//   - float64: The total time converted to seconds.
//   - error: An error if parsing fails, typically due to invalid numeric input.
func (conv *Conv) parseSource() (float64, error) {
	parts := strings.Fields(conv.Source)
	var totalSeconds float64

	for i := 0; i < len(parts); i += 2 {
		value, err := strconv.ParseFloat(parts[i], 64)
		if err != nil {
			return 0, err
		}
		unit := strings.ToLower(removeTrailingS(parts[i+1]))
		if seconds, ok := unitMap[unit]; ok {
			totalSeconds += value * seconds
		}
	}
	return totalSeconds, nil
}

// formatTarget converts a duration in seconds to a human-readable
// string representation using the specified time units.
//
// Parameters:
//   - seconds: The duration to format, expressed in seconds.
//   - units: A slice of strings representing the desired time units for the output.
//     These should correspond to keys in the unitMap (e.g., "hour", "minute", "second").
//
// The function iterates through the provided units, converting the input seconds into
// each unit as appropriate. It builds a string representation, including only non-zero
// values. The function handles singular and plural forms of the units.
//
// Returns:
//   - string: A formatted string representing the duration using the specified units.
//     For example: "2 hours 30 minutes 45 seconds"
func (conv *Conv) formatTarget(seconds float64, units []string) string {
	result := ""
	for _, unit := range units {
		unitInSeconds := unitMap[strings.TrimSuffix(unit, "s")]
		value := seconds / unitInSeconds
		intValue := int(value)

		if intValue <= 0 {
			continue
		}

		unit = removeTrailingS(unit)
		if intValue > 1 {
			unit += "s"
		}

		result += fmt.Sprintf("%d %s ", intValue, unit)
		seconds -= float64(intValue) * unitInSeconds
	}
	return strings.TrimSpace(result)
}

// expandBriefSourceDuration expands a brief period into a long period format
// example: Dhm => "days hours minutes"
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
	return strings.TrimSpace(period), nil
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
		i++
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

// ConvertDuration converts a duration from the format specified in Conv.Source
// to the format specified in Conv.Target.
//
// This function performs the following steps:
// 1. If the Source is in brief format (e.g., "1h30m"), it expands it to long format.
// 2. Parses the Source string to calculate the total duration in seconds.
// 3. If the Target is in brief format, it expands it to a list of unit names.
// 4. Formats the duration in seconds according to the specified target units.
// 5. If Conv.Brief is true, it converts the result back to brief format.
//
// The function handles both brief (e.g., "1h30m") and long (e.g., "1 hour 30 minutes") formats
// for both input and output.
//
// Returns:
//   - string: The converted duration in the specified target format.
//   - error: An error if any step of the conversion process fails.
func (conv *Conv) ConvertDuration() (string, error) {
	var err error
	isNegativeDuration := false
	if s, found := strings.CutPrefix(conv.Source, "-"); found {
		conv.Source = s
		isNegativeDuration = true
	}
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
	result = strings.TrimSpace(result)
	if isNegativeDuration {
		result = "-" + result
	}
	return result, nil
}
