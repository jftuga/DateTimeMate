package cmd

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"

	"github.com/jftuga/DateTimeMate"
	"github.com/spf13/cobra"
)

var optTzListZones bool
var optTzForce bool

var tzCmd = &cobra.Command{
	Use:   "tz [date/time] [target time zone]",
	Short: "Convert a date/time from one time zone to another",
	Args: func(cmd *cobra.Command, args []string) error {
		if optTzListZones {
			return nil
		}
		return cobra.ExactArgs(2)(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if optTzListZones {
			listZones()
			return
		}
		outputTzConversion(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(tzCmd)
	tzCmd.Flags().BoolVarP(&optTzListZones, "list-zones", "l", false, "list the supported time zone abbreviations and exit")
	tzCmd.Flags().BoolVarP(&optTzForce, "force", "f", false, "convert date/times before 1970 despite unreliable time zone data")
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
	for _, warning := range tz.Warnings(source, target) {
		fmt.Fprintln(os.Stderr, "warning:", warning)
	}
	result, err := tz.ConvertTimeZone(source, target)
	if err != nil {
		if errors.Is(err, DateTimeMate.ErrPre1970) {
			fmt.Fprintln(os.Stderr, err, "(use --force to convert anyway)")
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
	formatted := result.Format("2006-01-02 15:04:05 MST")
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
	fmt.Println("IANA zone names such as America/New_York or Asia/Kolkata are also supported")
	fmt.Println("and preferred: unlike the fixed offsets above, they are DST aware.")
	fmt.Printf("Override an ambiguous abbreviation with the %s environment\n", DateTimeMate.ZoneAliasesEnvVar)
	fmt.Printf("variable, e.g. %s=\"IST=Asia/Jerusalem|CST=Asia/Shanghai\"\n", DateTimeMate.ZoneAliasesEnvVar)
}
