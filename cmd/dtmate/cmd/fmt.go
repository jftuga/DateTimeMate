package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/jftuga/DateTimeMate"
	"github.com/spf13/cobra"
)

var fmtCmd = &cobra.Command{
	Use:   "fmt [date/time] [format specifiers]",
	Short: "reformat a date/time",
	Args: func(cmd *cobra.Command, args []string) error {
		if optFmtList {
			return nil
		}
		if len(args) != 2 {
			return errors.New("requires two arguments: [date/time] [format specifiers]")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if optFmtList {
			listConversionsSpecifiers()
			return
		}
		reformat(args[0], args[1])
	},
}

var optFmtList bool

func init() {
	rootCmd.AddCommand(fmtCmd)
	fmtCmd.Flags().BoolVarP(&optFmtList, "list", "l", false, "list supported conversion specifiers")
}

// listConversionsSpecifiers the list was copied from:
// https://github.com/lestrrat-go/strftime
func listConversionsSpecifiers() {
	list := [][]string{
		{`%A`, `national representation of the full weekday name`},
		{`%a`, `national representation of the abbreviated weekday`},
		{`%B`, `national representation of the full month name`},
		{`%b`, `national representation of the abbreviated month name`},
		{`%C`, `(year / 100) as decimal number; single digits are preceded by a zero`},
		{`%c`, `national representation of time and date`},
		{`%D`, `equivalent to %m/%d/%y`},
		{`%d`, `day of the month as a decimal number (01-31)`},
		{`%e`, `the day of the month as a decimal number (1-31); single digits are preceded by a blank`},
		{`%F`, `equivalent to %Y-%m-%d`},
		{`%H`, `the hour (24-hour clock) as a decimal number (00-23)`},
		{`%h`, `same as %b`},
		{`%I`, `the hour (12-hour clock) as a decimal number (01-12)`},
		{`%j`, `the day of the year as a decimal number (001-366)`},
		{`%k`, `the hour (24-hour clock) as a decimal number (0-23); single digits are preceded by a blank`},
		{`%l`, `the hour (12-hour clock) as a decimal number (1-12); single digits are preceded by a blank`},
		{`%M`, `the minute as a decimal number (00-59)`},
		{`%m`, `the month as a decimal number (01-12)`},
		{`%n`, `a newline`},
		{`%p`, `national representation of either 'ante meridiem' (a.m.)  or 'post meridiem' (p.m.)  as appropriate.`},
		{`%R`, `equivalent to %H:%M`},
		{`%r`, `equivalent to %I:%M:%S %p`},
		{`%S`, `the second as a decimal number (00-60)`},
		{`%s`, `the number of seconds since the Epoch, 1970-01-01 00:00:00 +0000 (UTC)`},
		{`%T`, `equivalent to %H:%M:%S`},
		{`%t`, `a tab`},
		{`%U`, `the week number of the year (Sunday as the first day of the week) as a decimal number (00-53)`},
		{`%u`, `the weekday (Monday as the first day of the week) as a decimal number (1-7)`},
		{`%V`, `the week number of the year (Monday as the first day of the week) as a decimal number (01-53)`},
		{`%v`, `equivalent to %e-%b-%Y`},
		{`%W`, `the week number of the year (Monday as the first day of the week) as a decimal number (00-53)`},
		{`%w`, `the weekday (Sunday as the first day of the week) as a decimal number (0-6)`},
		{`%X`, `national representation of the time`},
		{`%x`, `national representation of the date`},
		{`%Y`, `the year with century as a decimal number`},
		{`%y`, `the year without century as a decimal number (00-99)`},
		{`%Z`, `the time zone name`},
		{`%z`, `the time zone offset from UTC`},
		{`%%`, `a '%'`},
	}

	for i := 0; i < len(list); i++ {
		fmt.Printf("%s  %s\n", list[i][0], list[i][1])
	}
}

func reformat(source, outputFormat string) {
	result, err := DateTimeMate.Reformat(source, outputFormat)
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
