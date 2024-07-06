package cmd

import (
	"fmt"
	"github.com/jftuga/DateTimeMate"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// durCmd represents the dur command
var durCmd = &cobra.Command{
	Use:   "dur [from] [duration]",
	Short: "Output a date/time when given a starting date/time and duration",
	Args:  cobra.MatchAll(cobra.ExactArgs(2)),
	Run: func(cmd *cobra.Command, args []string) {
		outputDur(args[0], args[1], optDurUntil, optDurFormat, optDurRepeat)
	},
}

var (
	//optDurFrom   string
	optDurAdd    bool
	optDurSub    bool
	optDurUntil  string
	optDurFormat string
	optDurRepeat int
)

func init() {
	rootCmd.AddCommand(durCmd)
	durCmd.Flags().BoolVarP(&optDurAdd, "add", "a", false, "add: a duration to use with -f, such as '1D2h3s' or '1 day 2 hours 3 seconds'")
	durCmd.Flags().BoolVarP(&optDurSub, "sub", "s", false, "subtract: a duration to use with -f, such as '5 months 4 weeks 3 days'")
	durCmd.Flags().StringVarP(&optDurUntil, "until", "u", "", "repeat duration until this date/time is exceeded")
	durCmd.Flags().StringVarP(&optDurFormat, "format", "f", "", "output results with strftime formatting")
	durCmd.Flags().IntVarP(&optDurRepeat, "repeat", "r", 0, "repeat the -a or -s duration this number of times (mutually exclusive with -u)")
	durCmd.MarkFlagsOneRequired("add", "sub")
	durCmd.MarkFlagsMutuallyExclusive("add", "sub")
	durCmd.MarkFlagsMutuallyExclusive("repeat", "until")
}

func outputDur(from, duration, until, format string, repeat int) {
	dur := DateTimeMate.NewDur(
		DateTimeMate.DurWithFrom(from),
		DateTimeMate.DurWithDur(duration),
		DateTimeMate.DurWithUntil(until),
		DateTimeMate.DurWithRepeat(repeat),
		DateTimeMate.DurWithOutputFormat(format))

	var allResults []string
	var err error
	if optDurAdd {
		allResults, err = dur.Add()
	} else {
		allResults, err = dur.Sub()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	delim := "\n"
	if optRootNoNewline {
		delim = ","
	}
	output := strings.Join(allResults, delim)
	fmt.Print(output)
	if !optRootNoNewline {
		fmt.Println()
	}
}
