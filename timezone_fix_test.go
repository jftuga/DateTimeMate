// timezone_fix_test.go pins the behaviors introduced by the tz bug-fix
// batch (findings TZ-1 through TZ-12 of the v1.13.0 tz bug hunt): unix
// timestamp sources, rejection of contradictory zone combinations, ±HH
// offset suffixes, offset error propagation, case-insensitive IANA names,
// and alias parsing edge cases. Unlike the original timezone_test.go it
// uses only the standard testing package, per current repo policy.
// Instants are compared via Unix seconds so results do not depend on the
// time zone of the machine running the tests.
package DateTimeMate

import (
	"strings"
	"testing"
	"time"
)

// mustConvertUnix converts and returns the instant as Unix seconds,
// failing the test on error
func mustConvertUnix(t *testing.T, conv *TimeZoneConverter, source, target string) int64 {
	t.Helper()
	result, err := conv.ConvertTimeZone(source, target)
	if err != nil {
		t.Fatalf("ConvertTimeZone(%q, %q) unexpected error: %v", source, target, err)
	}
	return result.Unix()
}

// mustFailConvert converts, requires an error, and requires the error
// message to contain want
func mustFailConvert(t *testing.T, conv *TimeZoneConverter, source, target, want string) {
	t.Helper()
	_, err := conv.ConvertTimeZone(source, target)
	if err == nil {
		t.Fatalf("ConvertTimeZone(%q, %q) expected an error containing %q, got none", source, target, want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("ConvertTimeZone(%q, %q) error %q does not contain %q", source, target, err, want)
	}
}

func TestTimezoneUnixTimestampSources(t *testing.T) {
	conv := setupConverter()

	// TZ-1: 10-digit seconds and 13-digit milliseconds parse as unix
	// timestamps, consistent with the fmt sub-command
	if got := mustConvertUnix(t, conv, "1700265600", "UTC"); got != 1700265600 {
		t.Errorf("10-digit source: got unix %d, want 1700265600", got)
	}
	if got := mustConvertUnix(t, conv, "1700265600000", "UTC"); got != 1700265600 {
		t.Errorf("13-digit source: got unix %d, want 1700265600", got)
	}

	// TZ-1: other bare-integer digit counts error instead of being
	// misread as a time of day on the current date
	mustFailConvert(t, conv, "19800", "UTC", "ambiguous integer date/time")

	// TZ-1 (decided): a unix timestamp already denotes an instant, so a
	// zone suffix is contradictory
	mustFailConvert(t, conv, "1700265600 UTC", "America/New_York", "unix timestamp cannot carry a time zone")

	// compact integer date/times with a zone suffix keep working and are
	// interpreted in that zone
	want := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC).Unix()
	if got := mustConvertUnix(t, conv, "20240115120000 UTC", "EST"); got != want {
		t.Errorf("14-digit source with zone: got unix %d, want %d", got, want)
	}
}

func TestTimezoneRelativeWordsWithZone(t *testing.T) {
	conv := setupConverter()

	// TZ-4 (decided): relative words already denote an instant
	mustFailConvert(t, conv, "now UTC", "America/New_York", "relative date/times cannot carry a time zone")
	mustFailConvert(t, conv, "tomorrow EST", "UTC", "relative date/times cannot carry a time zone")

	// bare relative words keep working
	if _, err := conv.ConvertTimeZone("tomorrow", "UTC"); err != nil {
		t.Errorf("bare relative word: unexpected error: %v", err)
	}
}

func TestTimezoneFabricatedAbbreviationRejected(t *testing.T) {
	conv := setupConverter()

	// TZ-2: an abbreviation in non-final position that the local time
	// zone does not define must error loudly, not silently parse as UTC+0
	mustFailConvert(t, conv, "Mon Jan 15 07:00:00 XYZ 2024", "UTC", `zone abbreviation "XYZ" is not resolvable`)

	// the same guard protects the shared parse chain used by fmt
	if _, err := Reformat("Mon Jan 15 07:00:00 XYZ 2024", "%Y-%m-%d"); err == nil {
		t.Error("Reformat with a fabricated abbreviation: expected an error, got none")
	}

	// numeric ±NN zone tokens are repairable, not fabricated: the digits
	// carry the real offset, so the shared chain honors it instead of the
	// zero offset time.Parse invents
	if got, err := Reformat("2024-01-15 20:00:00 +08", "%Y-%m-%dT%H:%M:%S %z"); err != nil || got != "2024-01-15T20:00:00 +0800" {
		t.Errorf("Reformat +08: got %q, %v; want \"2024-01-15T20:00:00 +0800\"", got, err)
	}

	// "+00" legitimately denotes UTC+0 (e.g. Antarctica/Troll) and must
	// not be rejected
	if got, err := Reformat("2024-01-15 12:00:00 +00", "%z"); err != nil || got != "+0000" {
		t.Errorf("Reformat +00: got %q, %v; want \"+0000\"", got, err)
	}
}

func TestTimezoneOffsetSuffixSources(t *testing.T) {
	conv := setupConverter()
	want := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC).Unix()

	// TZ-3a: trailing ±NN tokens are whole-hour UTC offsets, the
	// abbreviation form many IANA zones print
	if got := mustConvertUnix(t, conv, "2024-01-15 20:00:00 +08", "UTC"); got != want {
		t.Errorf("+08 suffix: got unix %d, want %d", got, want)
	}
	if got := mustConvertUnix(t, conv, "2024-01-15 00:00:00 -12", "UTC"); got != want {
		t.Errorf("-12 suffix: got unix %d, want %d", got, want)
	}

	// a ±NN suffix outside -12:00..+14:00 errors instead of being
	// silently dropped by a fallback parser
	mustFailConvert(t, conv, "2024-01-15 12:00:00 +15", "UTC", "out of valid range")
}

func TestTimezoneOutputFormatRoundTrip(t *testing.T) {
	conv := setupConverter()
	const cliFormat = "2006-01-02 15:04:05 -0700 MST"
	want := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC).Unix()

	// TZ-3b: the CLI output format re-parses to the same instant, even
	// for zones with numeric abbreviations such as +08 and -12
	for _, target := range []string{"Antarctica/Casey", "Etc/GMT+12", "America/New_York", "Asia/Kathmandu"} {
		result, err := conv.ConvertTimeZone("2024-01-15 12:00:00 UTC", target)
		if err != nil {
			t.Fatalf("convert to %s: %v", target, err)
		}
		formatted := result.Format(cliFormat)
		if got := mustConvertUnix(t, conv, formatted, "UTC"); got != want {
			t.Errorf("round trip via %s (%q): got unix %d, want %d", target, formatted, got, want)
		}
	}
}

func TestTimezoneConflictingSourceZones(t *testing.T) {
	conv := setupConverter()

	// TZ-10 (decided): a wall clock carrying its own offset that
	// disagrees with the trailing zone is contradictory
	mustFailConvert(t, conv, "2024-01-15 12:00:00 +0900 EST", "UTC", "source names both")

	// when they agree, the parse is accepted and the offset is honored
	want := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC).Unix()
	if got := mustConvertUnix(t, conv, "2024-01-15 07:00:00 -0500 EST", "UTC"); got != want {
		t.Errorf("agreeing offset and zone: got unix %d, want %d", got, want)
	}

	// the Jerusalem round trip: without an alias, IST resolves to India
	// (+05:30) which disagrees with +0200 and must error loudly
	mustFailConvert(t, conv, "2024-01-15 14:00:00 +0200 IST", "UTC", "source names both")

	// with the alias, IST resolves to Asia/Jerusalem (+02:00 in January),
	// the offsets agree, and the round trip is exact
	aliases, err := ParseZoneAliases("IST=Asia/Jerusalem")
	if err != nil {
		t.Fatalf("ParseZoneAliases: %v", err)
	}
	aliased := NewTimeZoneConverter(
		TimeZoneConverterWithZoneAbbrevs(LoadZoneDefinitions()),
		TimeZoneConverterWithAliases(aliases))
	if got := mustConvertUnix(t, aliased, "2024-01-15 14:00:00 +0200 IST", "UTC"); got != want {
		t.Errorf("aliased IST round trip: got unix %d, want %d", got, want)
	}
}

func TestTimezoneSecondsOffsetTargets(t *testing.T) {
	conv := setupConverter()

	// TZ-5 (decided): sign-prefixed 3-4 digit targets read as ±HHMM, so
	// reject them with the equivalent value in seconds; TZ-8 makes the
	// detail visible instead of a bare "invalid timezone specification"
	mustFailConvert(t, conv, "2024-01-15 12:00:00 UTC", "+0530", "use 19800 for +05:30")
	mustFailConvert(t, conv, "2024-01-15 12:00:00 UTC", "-0500", "use -18000 for -05:00")

	// TZ-8: the range message reaches the user
	mustFailConvert(t, conv, "2024-01-15 12:00:00 UTC", "50401", "out of valid range")

	// bare digit strings remain seconds, now labeled without truncation
	// (TZ-9)
	result, err := conv.ConvertTimeZone("2024-01-15 12:00:00 UTC", "19800")
	if err != nil {
		t.Fatalf("seconds target: %v", err)
	}
	if name, _ := result.Zone(); name != "UTC+05:30" {
		t.Errorf("seconds target zone name: got %q, want \"UTC+05:30\"", name)
	}
}

func TestTimezoneCaseInsensitiveIANANames(t *testing.T) {
	conv := setupConverter()
	want := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC).Unix()

	// TZ-7 (decided): IANA names are case-insensitive on every platform,
	// as both target and source
	if got := mustConvertUnix(t, conv, "2024-01-15 12:00:00 UTC", "america/new_york"); got != want {
		t.Errorf("lowercase target: got unix %d, want %d", got, want)
	}
	if got := mustConvertUnix(t, conv, "2024-01-15 21:00:00 asia/tokyo", "UTC"); got != want {
		t.Errorf("lowercase source: got unix %d, want %d", got, want)
	}

	// alias values are canonicalized the same way
	if _, err := ParseZoneAliases("IST=asia/jerusalem"); err != nil {
		t.Errorf("lowercase alias value: unexpected error: %v", err)
	}
}

func TestTimezoneDigitBearingZoneNames(t *testing.T) {
	conv := setupConverter()

	// TZ-11: IANA names containing digits work as a source zone suffix
	want := time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC).Unix()
	if got := mustConvertUnix(t, conv, "2024-07-15 08:00:00 EST5EDT", "UTC"); got != want {
		t.Errorf("EST5EDT source: got unix %d, want %d", got, want)
	}
}

func TestTimezoneAliasSpecEdgeCases(t *testing.T) {
	// TZ-12 (decided): empty segments from trailing or doubled pipes are
	// skipped
	aliases, err := ParseZoneAliases("IST=Asia/Jerusalem|")
	if err != nil {
		t.Fatalf("trailing pipe: unexpected error: %v", err)
	}
	if len(aliases) != 1 || aliases["IST"] != "Asia/Jerusalem" {
		t.Errorf("trailing pipe: got %v, want map[IST:Asia/Jerusalem]", aliases)
	}

	// TZ-12 (decided): duplicate keys error instead of silently last-winning
	_, err = ParseZoneAliases("IST=Asia/Jerusalem|IST=Asia/Kolkata")
	if err == nil {
		t.Fatal("duplicate alias keys: expected an error, got none")
	}
	if !strings.Contains(err.Error(), `duplicate alias "IST"`) {
		t.Errorf("duplicate alias keys: error %q does not name the duplicate", err)
	}
}
