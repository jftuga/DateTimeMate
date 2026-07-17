package cmd

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"time"

	"github.com/jftuga/DateTimeMate"
	"github.com/spf13/cobra"
)

var optTzListZones bool
var optTzListIANA bool
var optTzForce bool

var tzCmd = &cobra.Command{
	Use:   "tz [date/time] [target time zone]",
	Short: "Convert a date/time from one time zone to another",
	Args: func(cmd *cobra.Command, args []string) error {
		if optTzListZones || optTzListIANA {
			return cobra.NoArgs(cmd, args)
		}
		return cobra.ExactArgs(2)(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if optTzListZones {
			listZones()
			return
		}
		if optTzListIANA {
			listIANAZones()
			return
		}
		outputTzConversion(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(tzCmd)
	tzCmd.Flags().BoolVarP(&optTzListZones, "list-zones", "l", false, "list the supported time zone abbreviations and exit")
	tzCmd.Flags().BoolVarP(&optTzListIANA, "list-iana", "I", false, "list the IANA time zone names (e.g. America/New_York) and exit")
	tzCmd.Flags().BoolVarP(&optTzForce, "force", "f", false, "convert date/times before 1970 despite unreliable time zone data")
	tzCmd.MarkFlagsMutuallyExclusive("list-zones", "list-iana")
}

func newTimeZoneConverter() *DateTimeMate.TimeZoneConverter {
	aliases, err := DateTimeMate.ParseZoneAliases(os.Getenv(DateTimeMate.ZoneAliasesEnvVar))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", DateTimeMate.ZoneAliasesEnvVar, err)
		os.Exit(1)
	}
	return DateTimeMate.NewTimeZoneConverter(
		DateTimeMate.TimeZoneConverterWithZoneAbbrevs(DateTimeMate.LoadZoneDefinitions()),
		DateTimeMate.TimeZoneConverterWithAliases(aliases),
		DateTimeMate.TimeZoneConverterWithAllowPre1970(optTzForce))
}

func outputTzConversion(source, target string) {
	tz := newTimeZoneConverter()
	result, err := tz.ConvertTimeZone(source, target)
	if err != nil {
		if errors.Is(err, DateTimeMate.ErrPre1970) {
			fmt.Fprintln(os.Stderr, err, "(use --force to convert anyway)")
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
	for _, warning := range tz.Warnings(source, target) {
		fmt.Fprintln(os.Stderr, "warning:", warning)
	}
	formatted := result.Format("2006-01-02 15:04:05 -0700 MST")
	if optRootNoNewline {
		fmt.Print(formatted)
	} else {
		fmt.Println(formatted)
	}
}

func listZones() {
	zones := DateTimeMate.LoadZoneDefinitions()
	for _, abbrev := range slices.Sorted(maps.Keys(zones)) {
		def := zones[abbrev]
		note := ""
		if def.Ambiguous != "" {
			note = " [ambiguous; also: " + def.Ambiguous + "]"
		}
		fmt.Printf("%-6s UTC%s  %s%s\n", abbrev, DateTimeMate.FormatUTCOffset(def.Offset), def.Description, note)
	}
	fmt.Println()
	fmt.Println("Abbreviations always mean the fixed offsets shown, on any date: CET on a")
	fmt.Println("summer date stays UTC+01:00 rather than becoming CEST.")
	fmt.Println("IANA zone names such as America/New_York or Asia/Kolkata are also supported")
	fmt.Println("and preferred: unlike the fixed offsets above, they are DST aware.")
	fmt.Println("List them with: dtmate tz --list-iana")
	fmt.Printf("Override an ambiguous abbreviation with the %s environment\n", DateTimeMate.ZoneAliasesEnvVar)
	fmt.Printf("variable, e.g. %s=\"IST=Asia/Jerusalem|CST=Asia/Shanghai\"\n", DateTimeMate.ZoneAliasesEnvVar)
}

// listIANAZones prints each IANA zone name with the UTC offset and
// abbreviation currently in effect there
func listIANAZones() {
	now := time.Now()
	fmt.Printf("offsets and abbreviations are those currently in effect (%s)\n", now.Format("2006-01-02"))
	for _, name := range DateTimeMate.ListIANAZones() {
		loc, err := time.LoadLocation(name)
		if err != nil {
			continue
		}
		abbrev, offset := now.In(loc).Zone()
		fmt.Printf("%-32s UTC%s (%s)\n", name, DateTimeMate.FormatUTCOffset(offset), abbrev)
	}
}
