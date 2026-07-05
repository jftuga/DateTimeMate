package cmd

import (
	"fmt"
	"os"

	"github.com/jftuga/DateTimeMate"
	"github.com/spf13/cobra"
)

var tzCmd = &cobra.Command{
	Use:   "tz [date/time] [target time zone]",
	Short: "Convert a date/time from one time zone to another",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		outputTzConversion(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(tzCmd)
}

func outputTzConversion(source, target string) {
	tz := DateTimeMate.NewTimeZoneConverter(DateTimeMate.TimeZoneConverterWithZoneAbbrevs(DateTimeMate.LoadZoneDefinitions()))
	result, err := tz.ConvertTimeZone(source, target)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	formatted := result.Format("2006-01-02 15:04:05 MST")
	if optRootNoNewline {
		fmt.Print(formatted)
	} else {
		fmt.Println(formatted)
	}
}
