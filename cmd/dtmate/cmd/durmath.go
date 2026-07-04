// durmath.go implements the 'durmath' sub-command, which adds or subtracts
// two durations that may be expressed in different units. The operation is
// selected with -a/--add or -s/--sub, and the signed result can optionally be
// converted to target units (-c), rounded (-d), output in brief form (-b),
// or rendered as an absolute (positive) duration (-A).

package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/jftuga/DateTimeMate"
	"github.com/spf13/cobra"
)

// durMathCmd represents the durmath command
var durMathCmd = &cobra.Command{
	Use:   "durmath [duration1] [duration2]",
	Short: "Add or subtract two durations",
	Example: `  dtmate durmath "1 hour 30 minutes" "45 minutes" -a
  dtmate durmath 1h30m 45m -a -b
  dtmate durmath "1 day" "90 minutes" -s -c minutes`,
	Args: cobra.MatchAll(cobra.ExactArgs(2)),
	RunE: func(cmd *cobra.Command, args []string) error {
		return outputDurMath(args[0], args[1])
	},
}

var (
	optDurMathAdd      bool
	optDurMathSub      bool
	optDurMathConv     string
	optDurMathBrief    bool
	optDurMathDecimals int
	optDurMathAbsolute bool
)

func init() {
	rootCmd.AddCommand(durMathCmd)
	durMathCmd.Flags().BoolVarP(&optDurMathAdd, "add", "a", false, "add the two durations")
	durMathCmd.Flags().BoolVarP(&optDurMathSub, "sub", "s", false, "subtract the second duration from the first")
	durMathCmd.Flags().StringVarP(&optDurMathConv, "conv", "c", "", "convert resulting duration to another group of units")
	durMathCmd.Flags().BoolVarP(&optDurMathBrief, "brief", "b", false, "output in brief format, such as: 1Y3W4D5h6m7s")
	durMathCmd.Flags().IntVarP(&optDurMathDecimals, "decimals", "d", 0, "with -c: show the smallest unit with this many decimal places, rounded")
	durMathCmd.Flags().BoolVarP(&optDurMathAbsolute, "absolute", "A", false, "always output an absolute (positive) duration")
	durMathCmd.MarkFlagsOneRequired("add", "sub")
	durMathCmd.MarkFlagsMutuallyExclusive("add", "sub")
	durMathCmd.SetFlagErrorFunc(negativeDurationHint("durmath", "Use -a/--add or -s/--sub to control the operation, e.g.:\n  dtmate durmath 2h 30m -s"))
}

// outputDurMath runs the requested duration arithmetic and prints the result;
// a negative-duration error is returned to cobra so the usage text is shown,
// consistent with the flag-parse path that catches leading-dash negatives
func outputDurMath(first, second string) error {
	if optDurMathDecimals != 0 && optDurMathConv == "" {
		fmt.Fprintln(os.Stderr, "-d/--decimals requires -c/--conv")
		os.Exit(1)
	}
	dm := DateTimeMate.NewDurMath(
		DateTimeMate.DurMathWithFirst(first),
		DateTimeMate.DurMathWithSecond(second),
		DateTimeMate.DurMathWithTarget(optDurMathConv),
		DateTimeMate.DurMathWithBrief(optDurMathBrief),
		DateTimeMate.DurMathWithDecimals(optDurMathDecimals),
		DateTimeMate.DurMathWithAbsolute(optDurMathAbsolute))

	var result string
	var err error
	if optDurMathAdd {
		result, err = dm.Add()
	} else {
		result, err = dm.Sub()
	}
	if err != nil {
		if errors.Is(err, DateTimeMate.ErrNegativeDuration) {
			// the trailing newline separates the error from cobra's usage
			// block, matching the negativeDurationHint output
			return fmt.Errorf("%w\n", err)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if optRootNoNewline {
		fmt.Print(result)
	} else {
		fmt.Println(result)
	}
	return nil
}
