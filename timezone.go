package DateTimeMate

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"
	"unicode"
)

var (
	ErrInvalidTimezone = errors.New("invalid timezone specification")
	ErrEmptyInput      = errors.New("empty input provided")
)

// TimeZoneConverter converts date/times between time zones; ZoneAbbrevs
// supplies fixed UTC offsets for abbreviations such as EST or JST that
// are not resolvable as IANA zone names
type TimeZoneConverter struct {
	ZoneAbbrevs map[string]int
}

type OptionsTimeZoneConverter func(*TimeZoneConverter)

// NewTimeZoneConverter creates a new timezone converter with the given configuration
func NewTimeZoneConverter(options ...OptionsTimeZoneConverter) *TimeZoneConverter {
	tzc := new(TimeZoneConverter)
	for _, opt := range options {
		opt(tzc)
	}
	return tzc
}

func TimeZoneConverterWithZoneAbbrevs(zoneAbbrevs map[string]int) OptionsTimeZoneConverter {
	return func(tzc *TimeZoneConverter) {
		tzc.ZoneAbbrevs = zoneAbbrevs
	}
}

// ConvertTimeZone converts a date/time string to the target time zone; the
// source may end in its own zone (IANA name or abbreviation), otherwise it
// is parsed as a local date/time
func (c *TimeZoneConverter) ConvertTimeZone(sourceTime string, targetZone string) (time.Time, error) {
	sourceTime = strings.TrimSpace(sourceTime)
	targetZone = strings.TrimSpace(targetZone)
	if sourceTime == "" || targetZone == "" {
		return time.Time{}, ErrEmptyInput
	}

	parsed, err := c.parseSourceTime(sourceTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse source time: %w", err)
	}

	targetLoc, err := c.resolveLocation(targetZone)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to resolve target timezone %q: %w", targetZone, err)
	}

	return parsed.In(targetLoc), nil
}

// wallClockLayouts are tried when interpreting the date/time preceding an
// explicit source zone; time.ParseInLocation is preferred over parsetime
// because parsetime silently corrupts pre-1970 date/times
var wallClockLayouts = []string{"2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02 15:04", "2006-01-02"}

// parseSourceTime parses a date/time string; when the last field names a
// resolvable time zone, the preceding wall clock is interpreted in that
// zone, otherwise the whole string is parsed with parseDateTime, which
// assumes the local zone for zone-less date/times
func (c *TimeZoneConverter) parseSourceTime(input string) (time.Time, error) {
	if idx := strings.LastIndex(input, " "); idx != -1 {
		zone := input[idx+1:]
		if isZoneName(zone) {
			if loc, err := c.resolveLocation(zone); err == nil {
				wall := strings.TrimSpace(input[:idx])
				for _, layout := range wallClockLayouts {
					if t, err := time.ParseInLocation(layout, wall, loc); err == nil {
						return t, nil
					}
				}
				t, err := parseDateTime(wall)
				if err != nil {
					return time.Time{}, err
				}
				return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc), nil
			}
		}
	}
	return parseDateTime(ConvertRelativeDateToActual(input))
}

// isZoneName reports whether a field can name a time zone: an IANA path
// such as America/New_York or an alphabetic abbreviation such as EST;
// numeric fields (e.g. "-0500") are left to the date/time parser
func isZoneName(s string) bool {
	if strings.Contains(s, "/") {
		return true
	}
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return s != ""
}

// resolveLocation resolves a time zone given as an IANA name (DST aware),
// an abbreviation from ZoneAbbrevs, or a UTC offset in seconds
func (c *TimeZoneConverter) resolveLocation(zone string) (*time.Location, error) {
	if loc, err := time.LoadLocation(zone); err == nil {
		return loc, nil
	}
	if offset, ok := c.ZoneAbbrevs[strings.ToUpper(zone)]; ok {
		return time.FixedZone(strings.ToUpper(zone), offset), nil
	}
	if offset, err := parseOffset(zone); err == nil {
		return time.FixedZone(fmt.Sprintf("UTC%+d", offset/3600), offset), nil
	}
	return nil, ErrInvalidTimezone
}

// parseOffset parses a UTC offset given in seconds (e.g. "19800" or
// "-34200") and validates it falls within -12:00 to +14:00
func parseOffset(offset string) (int, error) {
	seconds, err := strconv.Atoi(strings.TrimPrefix(offset, "+"))
	if err != nil {
		return 0, fmt.Errorf("invalid offset %q: %w", offset, err)
	}
	if seconds < -43200 || seconds > 50400 {
		return 0, fmt.Errorf("offset %q out of valid range (-12:00 to +14:00)", offset)
	}
	return seconds, nil
}
