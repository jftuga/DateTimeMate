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
		if strings.TrimSpace(pair) == "" {
			continue
		}
		abbrev, zone, found := strings.Cut(pair, "=")
		abbrev = strings.TrimSpace(abbrev)
		zone = strings.TrimSpace(zone)
		if !found || abbrev == "" || zone == "" {
			return nil, fmt.Errorf("invalid alias %q: expected ABBREVIATION=IANA-zone", pair)
		}
		if _, err := loadIANALocation(zone); err != nil {
			return nil, fmt.Errorf("invalid alias %q: %q is not an IANA time zone", pair, zone)
		}
		key := strings.ToUpper(abbrev)
		if existing, ok := aliases[key]; ok {
			return nil, fmt.Errorf("duplicate alias %q: %s and %s", key, existing, zone)
		}
		aliases[key] = zone
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
// resolvable time zone or a ±HH UTC offset, the preceding wall clock is
// interpreted in that zone, otherwise the whole string is parsed as a
// local date/time, with 10 and 13 digit integers read as unix timestamps
func (c *TimeZoneConverter) parseSourceTime(input string) (time.Time, error) {
	if idx := strings.LastIndex(input, " "); idx != -1 {
		token := input[idx+1:]
		wall := strings.TrimSpace(input[:idx])
		loc, shaped, err := parseOffsetSuffix(token)
		if err != nil {
			return time.Time{}, err
		}
		if !shaped && isZoneName(token) {
			if resolved, rerr := c.resolveLocation(token); rerr == nil {
				loc = resolved
			}
		}
		if loc != nil {
			return c.parseWallClockIn(input, wall, token, loc)
		}
	}
	return parseDateTimeOrUnix(input)
}

// parseWallClockIn interprets the wall clock preceding a trailing zone
// token in that zone; relative words and unix timestamps are rejected
// because they already denote an instant, and a wall clock carrying its
// own explicit zone or offset must agree with the trailing zone
func (c *TimeZoneConverter) parseWallClockIn(input, wall, zone string, loc *time.Location) (time.Time, error) {
	if ConvertRelativeDateToActual(wall) != wall {
		return time.Time{}, fmt.Errorf("relative date/times cannot carry a time zone: %q", input)
	}
	if isPureIntegerAtoi(wall) {
		if isUnixTimestamp(wall) {
			return time.Time{}, fmt.Errorf("a unix timestamp cannot carry a time zone: %q", input)
		}
		return parseIntegerDateTime(wall, loc)
	}
	for _, layout := range wallClockLayouts {
		if t, err := time.ParseInLocation(layout, wall, loc); err == nil {
			return t, nil
		}
	}
	t, err := parseDateTime(wall)
	if err != nil {
		return time.Time{}, err
	}
	if sourceHasExplicitZone(wall, t) {
		return reconcileZones(t, zone, loc)
	}
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc), nil
}

// reconcileZones handles a wall clock that carried its own zone or offset
// in addition to the trailing zone token: when both denote the same UTC
// offset at that instant the parsed instant is kept, otherwise the source
// is contradictory and rejected
func reconcileZones(t time.Time, zone string, loc *time.Location) (time.Time, error) {
	_, wallOffset := t.Zone()
	_, locOffset := t.In(loc).Zone()
	if wallOffset != locOffset {
		return time.Time{}, fmt.Errorf("source names both UTC%s and %s (UTC%s)", FormatUTCOffset(wallOffset), zone, FormatUTCOffset(locOffset))
	}
	return t.In(loc), nil
}

// parseOffsetSuffix interprets a ±NN field as a whole-hour UTC offset
// (e.g. "+08" or "-12", the abbreviation form many IANA zones print),
// used both for trailing source fields and for repairing zone tokens that
// time.Parse fabricated; ±NNNN fields are left to the date/time parser's
// -0700 layout. shaped reports whether the field has the ±NN form at all:
// a shaped but out-of-range field errors rather than falling through to
// parsers that would silently drop it
func parseOffsetSuffix(token string) (loc *time.Location, shaped bool, err error) {
	if len(token) != 3 || (token[0] != '+' && token[0] != '-') || !isAllDigits(token[1:]) {
		return nil, false, nil
	}
	hours, _ := strconv.Atoi(token[1:])
	seconds := hours * 3600
	if token[0] == '-' {
		seconds = -seconds
	}
	if seconds < -43200 || seconds > 50400 {
		return nil, true, fmt.Errorf("offset %q out of valid range (-12:00 to +14:00)", token)
	}
	return time.FixedZone(token, seconds), true, nil
}

// isZoneName reports whether a field can name a time zone: an IANA path
// such as America/New_York, an alphabetic abbreviation such as EST, or a
// letters-and-digits IANA name such as EST5EDT; pure-numeric and signed
// fields (e.g. "-0500") are left to the offset and date/time parsers
func isZoneName(s string) bool {
	if strings.Contains(s, "/") {
		return true
	}
	hasLetter := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
		default:
			return false
		}
	}
	return hasLetter
}

// resolveLocation resolves a time zone given as an aliased abbreviation,
// an abbreviation from ZoneAbbrevs (a fixed offset), an IANA name
// (DST aware, case-insensitive), or a UTC offset in seconds; abbreviations
// are checked before IANA names because CET, EET, and WET are also IANA
// legacy zones with DST rules, which would silently turn "08:30 CET" on a
// summer date into 08:30 CEST; every other colliding name (EST, MST, HST,
// GMT, UTC) is a fixed IANA zone with the same offset as the table entry
func (c *TimeZoneConverter) resolveLocation(zone string) (*time.Location, error) {
	upper := strings.ToUpper(zone)
	if target, ok := c.Aliases[upper]; ok {
		loc, err := loadIANALocation(target)
		if err != nil {
			return nil, fmt.Errorf("%w: alias %s=%s does not name an IANA time zone", ErrInvalidTimezone, upper, target)
		}
		return loc, nil
	}
	if def, ok := c.ZoneAbbrevs[upper]; ok {
		return time.FixedZone(upper, def.Offset), nil
	}
	if loc, err := loadIANALocation(zone); err == nil {
		return loc, nil
	}
	offset, err := parseOffset(zone)
	if err == nil {
		return time.FixedZone("UTC"+FormatUTCOffset(offset), offset), nil
	}
	if len(zone) > 0 && (zone[0] == '+' || zone[0] == '-' || (zone[0] >= '0' && zone[0] <= '9')) {
		return nil, fmt.Errorf("%w: %v", ErrInvalidTimezone, err)
	}
	return nil, ErrInvalidTimezone
}

// ianaZoneFold maps each generated IANA zone name, lowercased, to its
// canonical spelling so lookups are case-insensitive on every platform:
// filesystem lookups happen to be case-insensitive on macOS but not on
// Linux or in the embedded time/tzdata fallback
var ianaZoneFold = buildIANAZoneFold()

// buildIANAZoneFold builds the lowercased-to-canonical zone name map from
// the generated ianaZoneNames list
func buildIANAZoneFold() map[string]string {
	fold := make(map[string]string, len(ianaZoneNames))
	for _, name := range ianaZoneNames {
		fold[strings.ToLower(name)] = name
	}
	return fold
}

// loadIANALocation loads an IANA time zone by name, case-insensitively:
// the name is canonicalized against the generated zone list first, and
// names not on the list fall back to a direct lookup
func loadIANALocation(name string) (*time.Location, error) {
	if canonical, ok := ianaZoneFold[strings.ToLower(name)]; ok {
		name = canonical
	}
	return time.LoadLocation(name)
}

// parseOffset parses a UTC offset given in seconds (e.g. "19800" or
// "-34200") and validates it falls within -12:00 to +14:00; sign-prefixed
// 3-4 digit values are rejected because the user almost certainly meant
// ±HHMM, not seconds
func parseOffset(offset string) (int, error) {
	if len(offset) > 1 && (offset[0] == '+' || offset[0] == '-') {
		digits := offset[1:]
		if isAllDigits(digits) && (len(digits) == 3 || len(digits) == 4) {
			return 0, hhmmOffsetError(offset, digits)
		}
	}
	seconds, err := strconv.Atoi(strings.TrimPrefix(offset, "+"))
	if err != nil {
		return 0, fmt.Errorf("invalid offset %q: %w", offset, err)
	}
	if seconds < -43200 || seconds > 50400 {
		return 0, fmt.Errorf("offset %q out of valid range (-12:00 to +14:00)", offset)
	}
	return seconds, nil
}

// hhmmOffsetError builds the rejection for a sign-prefixed 3-4 digit
// offset, suggesting the equivalent value in seconds when the digits form
// a valid ±HHMM reading
func hhmmOffsetError(offset, digits string) error {
	hours, _ := strconv.Atoi(digits[:len(digits)-2])
	minutes, _ := strconv.Atoi(digits[len(digits)-2:])
	if minutes > 59 {
		return fmt.Errorf("offset %q looks like HH:MM; offsets are in seconds", offset)
	}
	seconds := hours*3600 + minutes*60
	if offset[0] == '-' {
		seconds = -seconds
	}
	return fmt.Errorf("offset %q looks like HH:MM; offsets are in seconds (use %d for %s)", offset, seconds, FormatUTCOffset(seconds))
}
