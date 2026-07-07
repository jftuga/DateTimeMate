package DateTimeMate

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// all duration math is carried out in integer nanoseconds so that integral
// amounts convert exactly; a year is 365.25 days, which is a whole number
// of nanoseconds, so even the fractional-day year stays exact
const (
	nanosPerMicrosecond int64 = 1_000
	nanosPerMillisecond int64 = 1_000_000
	nanosPerSecond      int64 = 1_000_000_000
	nanosPerMinute            = 60 * nanosPerSecond
	nanosPerHour              = 60 * nanosPerMinute
	nanosPerDay               = 24 * nanosPerHour
	nanosPerWeek              = 7 * nanosPerDay
	nanosPerYear              = 31_557_600_000_000_000 // 365.25 * nanosPerDay
)

var unitNanos = map[string]int64{
	"nanosecond":  1,
	"microsecond": nanosPerMicrosecond,
	"millisecond": nanosPerMillisecond,
	"second":      nanosPerSecond,
	"minute":      nanosPerMinute,
	"hour":        nanosPerHour,
	"day":         nanosPerDay,
	"week":        nanosPerWeek,
	"year":        nanosPerYear,
}

var unitBriefMap = map[string]string{
	"ns": "nanosecond",
	"us": "microsecond",
	"µs": "microsecond",
	"ms": "millisecond",
	"s":  "second",
	"m":  "minute",
	"h":  "hour",
	"D":  "day",
	"W":  "week",
	"Y":  "year",
}

// subSecondBriefUnits are the brief sub-second unit tokens accepted after
// the dot in a combined target such as "hms.msusns"; "ns", "us", and "µs"
// (but not "ms", which per-rune means minutes+seconds) are also accepted as
// a whole pre-dot segment
var subSecondBriefUnits = []string{"ms", "us", "µs", "ns"}

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

// isValidAmount reports whether s is a plain decimal amount matching
// an optional leading "-", digits, and an optional fractional part; this
// deliberately excludes the exponent ("1e2"), hex ("0x1p4"), "NaN", and
// "Inf" forms that strconv.ParseFloat would otherwise accept
func isValidAmount(s string) bool {
	s = strings.TrimPrefix(s, "-")
	whole, frac, hasDot := strings.Cut(s, ".")
	if !isAllDigits(whole) {
		return false
	}
	return !hasDot || isAllDigits(frac)
}

// errDurationRange is returned whenever a duration total or amount cannot
// be represented in int64 nanoseconds
var errDurationRange = errors.New("duration exceeds the supported range of about 292 years")

// addInt64Checked adds two int64 values, erroring on overflow
func addInt64Checked(a, b int64) (int64, error) {
	sum := a + b
	if (b > 0 && sum < a) || (b < 0 && sum > a) {
		return 0, errDurationRange
	}
	return sum, nil
}

// amountToNanos converts one textual amount of a unit to nanoseconds:
// integral amounts multiply exactly in int64, while fractional amounts fall
// back to float64 and are rounded to the nearest nanosecond
func amountToNanos(amount string, unitNs int64) (int64, error) {
	if !isValidAmount(amount) {
		return 0, fmt.Errorf("invalid amount: %q", amount)
	}
	if v, err := strconv.ParseInt(amount, 10, 64); err == nil {
		if v != 0 && (v > math.MaxInt64/unitNs || v < math.MinInt64/unitNs) {
			return 0, errDurationRange
		}
		return v * unitNs, nil
	}
	v, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return 0, err
	}
	ns := v * float64(unitNs)
	if ns >= math.MaxInt64 || ns <= math.MinInt64 {
		return 0, errDurationRange
	}
	return int64(math.Round(ns)), nil
}

// parseDurationNanos converts a duration string to a total number of
// nanoseconds.
//
// The source may be in long form ("1 hour 30 minutes") or brief form ("1h30m");
// brief input is first expanded to long form. Long form must contain alternating
// numeric values and time unit strings, with units (defined in unitNanos)
// accepted in both singular and plural forms. Integral amounts convert
// exactly; fractional amounts carry float64 precision (about 15-16
// significant digits). The total must fit in int64 nanoseconds, about
// +/-292 years.
func parseDurationNanos(source string) (int64, error) {
	if !isLongFormDuration(strings.Fields(source)) {
		// brief format is being used so convert to long duration format
		expandedSource, err := expandBriefSourceDuration(source)
		if nil != err {
			return 0, fmt.Errorf("invalid source duration %q: %v", source, err)
		}
		source = expandedSource
	}
	parts := strings.Fields(source)
	var total int64

	for i := 0; i < len(parts); i += 2 {
		if i+1 >= len(parts) {
			return 0, fmt.Errorf("missing unit after %q in: %s", parts[i], source)
		}
		unit := normalizeUnit(parts[i+1])
		unitNs, ok := unitNanos[unit]
		if !ok {
			return 0, fmt.Errorf("unknown source unit: %q", parts[i+1])
		}
		ns, err := amountToNanos(parts[i], unitNs)
		if err != nil {
			return 0, err
		}
		total, err = addInt64Checked(total, ns)
		if err != nil {
			return 0, err
		}
	}
	if total == math.MinInt64 {
		// reject the one value whose negation overflows, so callers can
		// always take the absolute value safely
		return 0, errDurationRange
	}
	return total, nil
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
		if _, ok := unitNanos[normalizeUnit(target)]; !ok {
			// brief format is being used so convert to long duration format
			var err error
			targetUnits, err = expandBriefTargetDuration(target)
			if nil != err {
				return nil, err
			}
			if len(targetUnits) == 0 {
				return nil, fmt.Errorf("no target units specified: %q", target)
			}
		}
	}
	for _, unit := range targetUnits {
		if _, ok := unitNanos[normalizeUnit(unit)]; !ok {
			return nil, fmt.Errorf("unknown target unit: %q", unit)
		}
	}
	return targetUnits, nil
}

// roundDiv divides a non-negative a by a positive b, rounding half up
func roundDiv(a, b int64) int64 {
	half := b / 2
	if a > math.MaxInt64-half {
		// avoid overflow at the very top of the range; plain truncation
		// only affects totals within half a quantum of 292 years
		return a / b
	}
	return (a + half) / b
}

// formatTarget converts a duration in nanoseconds to a human-readable
// string representation using the specified time units.
//
// Parameters:
//   - totalNs: The signed duration to format, expressed in nanoseconds.
//   - units: A slice of strings representing the desired time units for the output.
//     These should correspond to keys in unitNanos (e.g., "hour", "minute", "second").
//
// The function walks the units from largest to smallest with integer
// division, including only non-zero values and handling singular and plural
// forms. A negative duration is rendered with a single leading "-", which
// is omitted when the formatted value itself is zero.
//
// When Conv.Decimals is greater than zero, the last (smallest) unit is
// formatted with that many decimal places and is always included even when
// less than one; the total is rounded once, to the display precision of
// that unit, before the breakdown so a rounding carry propagates into the
// larger units (119.96 seconds becomes "2 minutes 0.0 seconds"). When the
// units are not commensurate (a year is not a whole number of weeks or
// days), the smaller units still absorb the remainder exactly as truncated
// integer arithmetic dictates.
//
// Returns:
//   - string: A formatted string representing the duration using the specified units.
//     For example: "2 hours 30 minutes 45 seconds"
func (conv *Conv) formatTarget(totalNs int64, units []string) string {
	negative := totalNs < 0
	rem := totalNs
	if negative {
		rem = -rem
	}

	pow10 := int64(1)
	for i := 0; i < conv.Decimals; i++ {
		pow10 *= 10
	}
	// pre-round the total once, at the display granularity of the smallest
	// unit, so a rounding carry propagates through the larger units
	// (119.96 seconds -> "2 minutes 0.0 seconds"); this is only valid when
	// every larger unit is a whole multiple of that granularity, otherwise
	// (e.g. years over weeks) the pre-rounding would perturb the larger
	// units and the last unit is rounded on its own instead
	preRounded := false
	lastNs := unitNanos[normalizeUnit(units[len(units)-1])]
	if conv.Decimals > 0 && lastNs%pow10 == 0 {
		quantum := lastNs / pow10
		preRounded = true
		for _, unit := range units[:len(units)-1] {
			if unitNanos[normalizeUnit(unit)]%quantum != 0 {
				preRounded = false
				break
			}
		}
		if preRounded && quantum > 1 {
			if rounded := roundDiv(rem, quantum); rounded <= math.MaxInt64/quantum {
				rem = rounded * quantum
			}
		}
	}

	result := ""
	nonZero := false
	for i, unit := range units {
		unit = normalizeUnit(unit)
		unitNs := unitNanos[unit]

		if conv.Decimals > 0 && i == len(units)-1 {
			whole := rem / unitNs
			sub := rem % unitNs
			var ticks int64
			switch {
			case preRounded:
				ticks = sub / (unitNs / pow10)
			case unitNs%pow10 == 0:
				ticks = roundDiv(sub, unitNs/pow10)
			default:
				// the requested precision is finer than a nanosecond, so
				// scale up instead; the operands are small enough that the
				// product cannot overflow
				ticks = roundDiv(sub*pow10, unitNs)
			}
			if ticks == pow10 {
				// the last unit rounded up to a whole value; carry one level
				whole++
				ticks = 0
			}
			if whole != 0 || ticks != 0 {
				nonZero = true
			}
			if whole != 1 || ticks != 0 {
				unit += "s"
			}
			result += fmt.Sprintf("%d.%0*d %s ", whole, conv.Decimals, ticks, unit)
			continue
		}

		value := rem / unitNs
		if value == 0 {
			continue
		}
		nonZero = true

		if value > 1 {
			unit += "s"
		}

		result += fmt.Sprintf("%d %s ", value, unit)
		rem -= value * unitNs
	}
	if result == "" {
		// every unit truncated to zero, so emit zero of the smallest unit
		// instead of an empty string
		result = fmt.Sprintf("0 %ss", normalizeUnit(units[len(units)-1]))
	}
	result = strings.TrimSpace(result)
	if negative && nonZero {
		result = "-" + result
	}
	return result
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
// before the dot each rune is a unit, so "ms" is minutes+seconds there; a
// bare "ms" target keeps that meaning for backward compatibility but warns
// on stderr, because "ms" means milliseconds everywhere else. A pre-dot
// segment that is exactly "ns", "us", or "µs" means that sub-second unit
// (per-rune those were never valid). A dot must be followed by sub-second
// units.
func expandBriefTargetDuration(period string) ([]string, error) {
	pre, post, hasDot := strings.Cut(period, ".")
	var result []string
	if longName, ok := unitBriefMap[pre]; ok && len(pre) > 1 && pre != "ms" {
		result = append(result, longName)
	} else {
		if period == "ms" {
			fmt.Fprintf(os.Stderr, "warning: target \"ms\" is ambiguous: interpreting as minutes+seconds; use \".ms\" or \"milliseconds\" for milliseconds\n")
		}
		for _, ch := range pre {
			longName, ok := unitBriefMap[string(ch)]
			if !ok {
				return nil, fmt.Errorf("[expandBriefTargetDuration] invalid unit %q in target %q; valid units are Y W D h m s, with sub-second units after a dot, such as \"hms.msusns\"", string(ch), period)
			}
			result = append(result, longName)
		}
	}
	if hasDot {
		if post == "" {
			return nil, fmt.Errorf("[expandBriefTargetDuration] missing sub-second units after the dot in target %q", period)
		}
		for post != "" {
			matched := false
			for _, brief := range subSecondBriefUnits {
				if strings.HasPrefix(post, brief) {
					result = append(result, unitBriefMap[brief])
					post = post[len(brief):]
					matched = true
					break
				}
			}
			if !matched {
				return nil, fmt.Errorf("[expandBriefTargetDuration] invalid sub-second unit %q in target %q", post, period)
			}
		}
	}
	return result, nil
}

// ConvertDuration converts a duration from the format specified in Conv.Source
// to the format specified in Conv.Target.
//
// This function performs the following steps:
// 1. If the Source is in brief format (e.g., "1h30m"), it expands it to long format.
// 2. Parses the Source string to calculate the total duration in nanoseconds.
// 3. If the Target is in brief format, it expands it to a list of unit names.
// 4. Formats the duration according to the specified target units.
// 5. If Conv.Brief is true, it converts the result back to brief format.
//
// The function handles both brief (e.g., "1h30m") and long (e.g., "1 hour 30 minutes") formats
// for both input and output. A leading "-" negates the whole source, and a
// net-negative total (e.g. "30 minutes -2 hours") is rendered with a
// leading "-".
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
	total, err := parseDurationNanos(source)
	if err != nil {
		return "", err
	}
	if isNegativeDuration {
		total = -total
	}
	targetUnits, err := resolveTargetUnits(conv.Target)
	if err != nil {
		return "", err
	}
	result := conv.formatTarget(total, targetUnits)
	if conv.Brief {
		result = shrinkPeriod(result)
	}
	return strings.TrimSpace(result), nil
}
