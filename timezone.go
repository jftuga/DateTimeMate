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

// ZoneAliasesEnvVar names the environment variable holding pipe-delimited
// abbreviation overrides, e.g. DTMATE_TZ_ALIASES="IST=Asia/Jerusalem|CST=Asia/Shanghai"
const ZoneAliasesEnvVar = "DTMATE_TZ_ALIASES"

var (
	ErrInvalidTimezone = errors.New("invalid timezone specification")
	ErrEmptyInput      = errors.New("empty input provided")
	ErrPre1970         = errors.New("date/times before 1970 are not converted by default because time zone data is unreliable before then")
)

// TimeZoneConverter converts date/times between time zones; ZoneAbbrevs
// supplies fixed UTC offsets for abbreviations such as EST or JST that are
// not resolvable as IANA zone names, Aliases maps abbreviations to IANA
// zone names and takes precedence over every other resolution, and
// AllowPre1970 permits conversions of date/times before 1970
type TimeZoneConverter struct {
	ZoneAbbrevs  map[string]ZoneDefinition
	Aliases      map[string]string
	AllowPre1970 bool
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

func TimeZoneConverterWithZoneAbbrevs(zoneAbbrevs map[string]ZoneDefinition) OptionsTimeZoneConverter {
	return func(tzc *TimeZoneConverter) {
		tzc.ZoneAbbrevs = zoneAbbrevs
	}
}

func TimeZoneConverterWithAliases(aliases map[string]string) OptionsTimeZoneConverter {
	return func(tzc *TimeZoneConverter) {
		tzc.Aliases = aliases
	}
}

func TimeZoneConverterWithAllowPre1970(allow bool) OptionsTimeZoneConverter {
	return func(tzc *TimeZoneConverter) {
		tzc.AllowPre1970 = allow
	}
}

// ParseZoneAliases parses pipe-delimited abbreviation overrides such as
// "IST=Asia/Jerusalem|CST=Asia/Shanghai"; keys are uppercased and every
// value must name a valid IANA time zone
func ParseZoneAliases(spec string) (map[string]string, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, nil
	}
	aliases := make(map[string]string)
	for _, pair := range strings.Split(spec, "|") {
		abbrev, zone, found := strings.Cut(pair, "=")
		abbrev = strings.TrimSpace(abbrev)
		zone = strings.TrimSpace(zone)
		if !found || abbrev == "" || zone == "" {
			return nil, fmt.Errorf("invalid alias %q: expected ABBREVIATION=IANA-zone", pair)
		}
		if _, err := time.LoadLocation(zone); err != nil {
			return nil, fmt.Errorf("invalid alias %q: %q is not an IANA time zone", pair, zone)
		}
		aliases[strings.ToUpper(abbrev)] = zone
	}
	return aliases, nil
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
	if parsed.Year() < 1970 && !c.AllowPre1970 {
		return time.Time{}, fmt.Errorf("%w: %s", ErrPre1970, sourceTime)
	}

	targetLoc, err := c.resolveLocation(targetZone)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to resolve target timezone %q: %w", targetZone, err)
	}

	return parsed.In(targetLoc), nil
}

// Warnings reports the ambiguous zone abbreviations a conversion of the
// given source and target would rely on, excluding any overridden by an
// alias; each message names ZoneAliasesEnvVar so the user can override
func (c *TimeZoneConverter) Warnings(sourceTime, targetZone string) []string {
	var warnings []string
	if w := c.ambiguityWarning(strings.TrimSpace(targetZone)); w != "" {
		warnings = append(warnings, w)
	}
	sourceTime = strings.TrimSpace(sourceTime)
	if idx := strings.LastIndex(sourceTime, " "); idx != -1 {
		zone := sourceTime[idx+1:]
		if isZoneName(zone) {
			if w := c.ambiguityWarning(zone); w != "" && (len(warnings) == 0 || warnings[0] != w) {
				warnings = append(warnings, w)
			}
		}
	}
	return warnings
}

// ambiguityWarning returns a warning when the zone is an abbreviation with
// multiple real-world meanings and no alias pins down which one is wanted
func (c *TimeZoneConverter) ambiguityWarning(zone string) string {
	upper := strings.ToUpper(zone)
	if _, ok := c.Aliases[upper]; ok {
		return ""
	}
	def, ok := c.ZoneAbbrevs[upper]
	if !ok || def.Ambiguous == "" {
		return ""
	}
	return fmt.Sprintf("%s is ambiguous: using %s (UTC%s), not %s; set %s=\"%s=<IANA zone>\" to override",
		upper, def.Description, FormatUTCOffset(def.Offset), def.Ambiguous, ZoneAliasesEnvVar, upper)
}

// ListIANAZones returns the IANA time zone names (e.g. America/New_York)
// that resolve against the time zone database in use, sorted; names from
// the generated ianaZoneNames list that no longer resolve are dropped
func ListIANAZones() []string {
	zones := make([]string, 0, len(ianaZoneNames))
	for _, name := range ianaZoneNames {
		if _, err := time.LoadLocation(name); err == nil {
			zones = append(zones, name)
		}
	}
	return zones
}

// FormatUTCOffset renders an offset in seconds east of UTC as ±HH:MM
func FormatUTCOffset(seconds int) string {
	sign := "+"
	if seconds < 0 {
		sign = "-"
		seconds = -seconds
	}
	return fmt.Sprintf("%s%02d:%02d", sign, seconds/3600, (seconds%3600)/60)
}

// parseSourceTime parses a date/time string; when the last field names a
// resolvable time zone, the preceding wall clock is interpreted in that
// zone, otherwise the whole string is parsed as a local date/time
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

// resolveLocation resolves a time zone given as an aliased abbreviation,
// an IANA name (DST aware), an abbreviation from ZoneAbbrevs, or a UTC
// offset in seconds
func (c *TimeZoneConverter) resolveLocation(zone string) (*time.Location, error) {
	upper := strings.ToUpper(zone)
	if target, ok := c.Aliases[upper]; ok {
		loc, err := time.LoadLocation(target)
		if err != nil {
			return nil, fmt.Errorf("%w: alias %s=%s does not name an IANA time zone", ErrInvalidTimezone, upper, target)
		}
		return loc, nil
	}
	if loc, err := time.LoadLocation(zone); err == nil {
		return loc, nil
	}
	if def, ok := c.ZoneAbbrevs[upper]; ok {
		return time.FixedZone(upper, def.Offset), nil
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
