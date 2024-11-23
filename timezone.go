package DateTimeMate

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/jftuga/parsetime"
	"github.com/pkg/errors"
)

// Standard time zone abbreviation mapping errors
var (
	ErrInvalidTimezone   = errors.New("invalid timezone specification")
	ErrInvalidTimeFormat = errors.New("invalid time format")
	ErrEmptyInput        = errors.New("empty input provided")
	ErrUnsupportedFormat = errors.New("unsupported time format")
)

// TimeZoneConverter handles timezone conversion operations
type TimeZoneConverter struct {
	Source      string
	TargetTZ    string
	ZoneAbbrevs map[string]int
}

type OptionsTimeZoneConverter func(*TimeZoneConverter)

// NewTimeZoneConverter creates a new timezone converter with the given configuration
func NewTimeZoneConverter(options ...OptionsTimeZoneConverter) (*TimeZoneConverter, error) {
	tzc := &TimeZoneConverter{}
	for _, opt := range options {
		opt(tzc)
	}

	return tzc, nil
}

func TimeZoneConverterWithZoneAbbrevs(zoneAbbrevs map[string]int) OptionsTimeZoneConverter {
	return func(tzc *TimeZoneConverter) {
		tzc.ZoneAbbrevs = zoneAbbrevs
	}
}

func TimeZoneConverterWithSource(source string) OptionsTimeZoneConverter {
	return func(tzc *TimeZoneConverter) {
		tzc.Source = source
	}
}

func TimeZoneConverterWithTargetTZ(targetTZ string) OptionsTimeZoneConverter {
	return func(tzc *TimeZoneConverter) {
		tzc.TargetTZ = targetTZ
	}
}

// ConvertTimeZone converts a time from one timezone to another
func (c *TimeZoneConverter) ConvertTimeZone(sourceTime string, targetZone string) (time.Time, error) {
	if sourceTime == "" || targetZone == "" {
		return time.Time{}, ErrEmptyInput
	}

	// Validate input lengths to prevent buffer overflow
	const maxInputLength = 100
	if len(sourceTime) > maxInputLength || len(targetZone) > maxInputLength {
		return time.Time{}, errors.New("input exceeds maximum length")
	}

	// Parse source time
	// parsed is of type time.Time
	parsed, err := c.parseTime(sourceTime)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "failed to parse source time")
	}

	// Get target location
	// targetLoc is of type time.Location
	targetLoc, err := c.resolveLocation(targetZone)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "failed to resolve target timezone")
	}

	println("parsed   : ", parsed.String())
	println("targetLoc: ", targetLoc.String())

	// Convert to target timezone
	// result is of type time.Time
	result := parsed.In(targetLoc)
	return result, nil
}

// FormatTime formats a time.Time according to the specified format
func (c *TimeZoneConverter) FormatTime(t time.Time, format string) string {
	return t.Format(format)
}

// parseOffset parses a timezone offset string (e.g., "-14400" or "+0800") into seconds
func parseOffset(offset string) (int, error) {
	// Remove any leading '+' sign
	offset = strings.TrimPrefix(offset, "+")

	// Try parsing as a number of seconds
	seconds, err := strconv.Atoi(offset)
	if err == nil {
		// Validate the offset is within reasonable bounds (-12:00 to +14:00)
		if seconds >= -43200 && seconds <= 50400 {
			return seconds, nil
		}
		return 0, errors.New("offset out of valid range (-12:00 to +14:00)")
	}

	// Try parsing as HHMM format
	if len(offset) == 4 {
		hours, err := strconv.Atoi(offset[:2])
		if err != nil {
			return 0, errors.Wrap(err, "invalid hours in offset")
		}
		minutes, err := strconv.Atoi(offset[2:])
		if err != nil {
			return 0, errors.Wrap(err, "invalid minutes in offset")
		}

		// Validate the parts
		if hours < -12 || hours > 14 {
			return 0, errors.New("hours out of valid range (-12 to +14)")
		}
		if minutes < 0 || minutes > 59 {
			return 0, errors.New("minutes out of valid range (0 to 59)")
		}

		return (hours * 3600) + (minutes * 60), nil
	}

	return 0, errors.New("invalid offset format")
}

// parseTime attempts to parse the input time string using configured formats
func (c *TimeZoneConverter) parseTime(input string) (time.Time, error) {
	p, err := parsetime.NewParseTime()
	if err != nil {
		return time.Time{}, err
	}

	t, err := p.Parse(input)
	fmt.Println("input:", input)
	fmt.Println("    t:", t)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}

// resolveLocation resolves a timezone specification to a time.Location
func (c *TimeZoneConverter) resolveLocation(zone string) (*time.Location, error) {
	// Try as IANA timezone first
	fmt.Println("zone:", zone)
	loc, err := time.LoadLocation(zone)
	if err != nil && !strings.Contains(err.Error(), "unknown time zone") {
		fmt.Println("xxx1  err:", err.Error())
		return loc, nil
	}

	// Try as offset
	offset, err := parseOffset(zone)
	if err != nil && !strings.Contains(err.Error(), "invalid") {
		fmt.Println("xxx2 err:", err.Error())
		return time.FixedZone(fmt.Sprintf("UTC%+d", offset/3600), offset), nil
	}

	// Try as abbreviation
	if offset, ok := c.ZoneAbbrevs[strings.ToUpper(zone)]; ok {
		fmt.Println("xxx3 ok:", offset, ok)
		return time.FixedZone(zone, offset), nil
	}

	fmt.Println("CRAP")
	return nil, ErrInvalidTimezone
}
