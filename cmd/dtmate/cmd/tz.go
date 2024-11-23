package cmd

import (
	"fmt"
	"github.com/jftuga/DateTimeMate"
	"os"

	"github.com/spf13/cobra"
)

var tzCmd = &cobra.Command{
	Use:   "tz [date/time] [target timezone]",
	Short: "Convert a date/time from one time zone to another",
	Args:  cobra.MatchAll(cobra.ExactArgs(2)),
	Run: func(cmd *cobra.Command, args []string) {
		outputTzConversion(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(tzCmd)
}

func outputTzConversion(source, target string) {
	defaultZones := DateTimeMate.LoadZoneDefinitions()
	tz, err := DateTimeMate.NewTimeZoneConverter(DateTimeMate.TimeZoneConverterWithSource(source), DateTimeMate.TimeZoneConverterWithTargetTZ(target), DateTimeMate.TimeZoneConverterWithZoneAbbrevs(defaultZones))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	result, err := tz.ConvertTimeZone(source, target) // FIXME: shouldn't use any args here
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if optRootNoNewline {
		fmt.Print(result)
	} else {
		fmt.Println(result)
	}
}
