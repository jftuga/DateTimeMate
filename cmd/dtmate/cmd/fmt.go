package cmd

import (
	"fmt"
	"github.com/jftuga/DateTimeMate"
	"github.com/spf13/cobra"
	"os"
)

var fmtCmd = &cobra.Command{
	Use:   "fmt [date/time] [format specifiers]",
	Short: "reformat a date/time",
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
		[]string{`%A`, `national representation of the full weekday name`},
		[]string{`%a`, `national representation of the abbreviated weekday`},
		[]string{`%B`, `national representation of the full month name`},
		[]string{`%b`, `national representation of the abbreviated month name`},
		[]string{`%C`, `(year / 100) as decimal number; single digits are preceded by a zero`},
		[]string{`%c`, `national representation of time and date`},
		[]string{`%D`, `equivalent to %m/%d/%y`},
		[]string{`%d`, `day of the month as a decimal number (01-31)`},
		[]string{`%e`, `the day of the month as a decimal number (1-31); single digits are preceded by a blank`},
		[]string{`%F`, `equivalent to %Y-%m-%d`},
		[]string{`%H`, `the hour (24-hour clock) as a decimal number (00-23)`},
		[]string{`%h`, `same as %b`},
		[]string{`%I`, `the hour (12-hour clock) as a decimal number (01-12)`},
		[]string{`%j`, `the day of the year as a decimal number (001-366)`},
		[]string{`%k`, `the hour (24-hour clock) as a decimal number (0-23); single digits are preceded by a blank`},
		[]string{`%l`, `the hour (12-hour clock) as a decimal number (1-12); single digits are preceded by a blank`},
		[]string{`%M`, `the minute as a decimal number (00-59)`},
		[]string{`%m`, `the month as a decimal number (01-12)`},
		[]string{`%n`, `a newline`},
		[]string{`%p`, `national representation of either 'ante meridiem' (a.m.)  or 'post meridiem' (p.m.)  as appropriate.`},
		[]string{`%R`, `equivalent to %H:%M`},
		[]string{`%r`, `equivalent to %I:%M:%S %p`},
		[]string{`%S`, `the second as a decimal number (00-60)`},
		[]string{`%T`, `equivalent to %H:%M:%S`},
		[]string{`%t`, `a tab`},
		[]string{`%U`, `the week number of the year (Sunday as the first day of the week) as a decimal number (00-53)`},
		[]string{`%u`, `the weekday (Monday as the first day of the week) as a decimal number (1-7)`},
		[]string{`%V`, `the week number of the year (Monday as the first day of the week) as a decimal number (01-53)`},
		[]string{`%v`, `equivalent to %e-%b-%Y`},
		[]string{`%W`, `the week number of the year (Monday as the first day of the week) as a decimal number (00-53)`},
		[]string{`%w`, `the weekday (Sunday as the first day of the week) as a decimal number (0-6)`},
		[]string{`%X`, `national representation of the time`},
		[]string{`%x`, `national representation of the date`},
		[]string{`%Y`, `the year with century as a decimal number`},
		[]string{`%y`, `the year without century as a decimal number (00-99)`},
		[]string{`%Z`, `the time zone name`},
		[]string{`%z`, `the time zone offset from UTC`},
		[]string{`%%`, `a '%'`},
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
