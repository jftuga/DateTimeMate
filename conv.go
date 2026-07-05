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
	Source   string
	Target   string
	Brief    bool
	Decimals int
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

// ConvWithDecimals sets the number of decimal places used when formatting
// the last (smallest) target unit; 0 keeps the default integer truncation
func ConvWithDecimals(decimals int) OptionsConv {
	return func(conv *Conv) {
		conv.Decimals = decimals
	}
}

func (conv *Conv) String() string {
	return fmt.Sprintf("Source:%v Target:%v Brief:%v Decimals:%v", conv.Source, conv.Target, conv.Brief, conv.Decimals)
}

// normalizeUnit lowercases a unit name and strips a plural trailing "s"
// so that forms such as "DAYS", "Days", and "days" all match the
// singular lowercase keys used in unitMap
func normalizeUnit(unit string) string {
	return removeTrailingS(strings.ToLower(unit))
}

// parseDurationSeconds converts a duration string to a total number of seconds.
//
// The source may be in long form ("1 hour 30 minutes") or brief form ("1h30m");
// brief input is first expanded to long form. Long form must contain alternating
// numeric values and time unit strings, with units (defined in unitMap) accepted
// in both singular and plural forms.
//
// Returns:
//   - float64: The total time converted to seconds.
//   - error: An error if parsing fails, typically due to invalid numeric input
//     or an unknown unit.
func parseDurationSeconds(source string) (float64, error) {
	if !isLongFormDuration(strings.Fields(source)) {
		// brief format is being used so convert to long duration format
		expandedSource, err := expandBriefSourceDuration(source)
		if nil != err {
			return 0, fmt.Errorf("invalid source duration %q: %v", source, err)
		}
		source = expandedSource
	}
	parts := strings.Fields(source)
	var totalSeconds float64

	for i := 0; i < len(parts); i += 2 {
		value, err := strconv.ParseFloat(parts[i], 64)
		if err != nil {
			return 0, err
		}
		if i+1 >= len(parts) {
			return 0, fmt.Errorf("missing unit after %q in: %s", parts[i], source)
		}
		unit := normalizeUnit(parts[i+1])
		seconds, ok := unitMap[unit]
		if !ok {
			return 0, fmt.Errorf("unknown source unit: %q", parts[i+1])
		}
		totalSeconds += value * seconds
	}
	return totalSeconds, nil
}

// resolveTargetUnits validates a target unit specification and expands it into
// a list of long-form unit names. The target may be space-separated long-form
// units ("hours minutes"), a single long-form unit, or a brief specification
// ("hm" or "hms.msusns"). An empty or whitespace-only target and any unknown
// unit produce an error.
func resolveTargetUnits(target string) ([]string, error) {
	targetUnits := strings.Fields(target)
	if len(targetUnits) == 0 {
		return nil, fmt.Errorf("no target units specified")
	}
	if len(targetUnits) == 1 {
		if _, ok := unitMap[normalizeUnit(target)]; !ok {
			// brief format is being used so convert to long duration format
			var err error
			targetUnits, err = expandBriefTargetDuration(target)
			if nil != err {
				return nil, err
			}
		}
	}
	for _, unit := range targetUnits {
		if _, ok := unitMap[normalizeUnit(unit)]; !ok {
			return nil, fmt.Errorf("unknown target unit: %q", unit)
		}
	}
	return targetUnits, nil
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
// When Conv.Decimals is greater than zero, the last (smallest) unit is formatted with
// that many decimal places, rounded, and is always included even when less than one.
//
// Returns:
//   - string: A formatted string representing the duration using the specified units.
//     For example: "2 hours 30 minutes 45 seconds"
func (conv *Conv) formatTarget(seconds float64, units []string) string {
	result := ""
	for i, unit := range units {
		unit = normalizeUnit(unit)
		unitInSeconds := unitMap[unit]
		value := seconds / unitInSeconds

		if conv.Decimals > 0 && i == len(units)-1 {
			if value < 0 { // guard against float underflow producing "-0.00"
				value = 0
			}
			formatted := strconv.FormatFloat(value, 'f', conv.Decimals, 64)
			if rounded, _ := strconv.ParseFloat(formatted, 64); rounded != 1 {
				unit += "s"
			}
			result += fmt.Sprintf("%s %s ", formatted, unit)
			continue
		}

		intValue := int(value)
		if intValue <= 0 {
			continue
		}

		if intValue > 1 {
			unit += "s"
		}

		result += fmt.Sprintf("%d %s ", intValue, unit)
		seconds -= float64(intValue) * unitInSeconds
	}
	if result == "" {
		// every unit truncated to zero, so emit zero of the smallest unit
		// instead of an empty string
		result = fmt.Sprintf("0 %ss", normalizeUnit(units[len(units)-1]))
	}
	return strings.TrimSpace(result)
}

// isLongFormDuration reports whether the fields form alternating
// amount/unit pairs, such as "1 hour 30 minutes"; anything else, including
// brief formats with spaces such as "1h 30m", needs brief expansion
func isLongFormDuration(fields []string) bool {
	if len(fields) == 0 || len(fields)%2 != 0 {
		return false
	}
	for i := 0; i < len(fields); i += 2 {
		if _, err := strconv.ParseFloat(fields[i], 64); err != nil {
			return false
		}
	}
	return true
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
	if conv.Decimals < 0 || conv.Decimals > 9 {
		return "", fmt.Errorf("decimals must be between 0 and 9: %d", conv.Decimals)
	}
	source := conv.Source
	isNegativeDuration := false
	if s, found := strings.CutPrefix(source, "-"); found {
		source = s
		isNegativeDuration = true
	}
	seconds, err := parseDurationSeconds(source)
	if err != nil {
		return "", err
	}
	targetUnits, err := resolveTargetUnits(conv.Target)
	if err != nil {
		return "", err
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
