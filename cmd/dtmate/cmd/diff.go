package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/jftuga/DateTimeMate"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff [start] [end]",
	Short: "Output the difference between two date/times",
	Example: `  dtmate diff 12:00:00 15:30:45
  dtmate diff 2024-06-07T08:00:00Z 2024-06-08T09:02:03Z --conv s -b`,
	Args: func(cmd *cobra.Command, args []string) error {
		if optDiffReadFromStdin {
			if len(args) == 0 {
				return nil
			} else {
				return errors.New("invalid number of arguments")
			}
		}
		if len(args) != 2 {
			return errors.New("requires two arguments: start, end date/times")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if optDiffReadFromStdin {
			start, end, err := getInput(os.Stdin)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			outputDiff(start, end, optDiffBrief)
			return
		}
		outputDiff(args[0], args[1], optDiffBrief)
	},
}

var optDiffBrief bool
var optDiffReadFromStdin bool
var optDiffConv string
var optDiffDecimals int
var optDiffAbsolute bool

func init() {
	rootCmd.AddCommand(diffCmd)
	diffCmd.Flags().BoolVarP(&optDiffBrief, "brief", "b", false, "output in brief format, such as: 1Y3W4D5h6m7s")
	diffCmd.Flags().BoolVarP(&optDiffReadFromStdin, "stdin", "i", false, "read from STDIN instead of using -s/-e")
	diffCmd.Flags().StringVarP(&optDiffConv, "conv", "c", "", "convert resulting duration to another group of units")
	diffCmd.Flags().IntVarP(&optDiffDecimals, "decimals", "d", 0, "with -c: show the smallest unit with this many decimal places, rounded")
	diffCmd.Flags().BoolVarP(&optDiffAbsolute, "absolute", "A", false, "always output an absolute (positive) duration")
}

// getInput reads the start and end date/times from r: either one line
// containing "start,end" or start on line one and end on line two; a
// non-empty second line takes precedence over a comma in the first line
// so that dates containing commas (e.g. "Jan 2, 2024") work in two-line mode
func getInput(r io.Reader) (string, string, error) {
	const usage = "expected 'start,end' on one line or start and end on two lines"
	input := bufio.NewScanner(r)
	if !input.Scan() {
		if err := input.Err(); err != nil {
			return "", "", err
		}
		return "", "", errors.New("no input on stdin: " + usage)
	}
	line := strings.TrimSpace(input.Text())
	var end string
	if input.Scan() {
		end = strings.TrimSpace(input.Text())
	}
	if err := input.Err(); err != nil {
		return "", "", err
	}
	if end != "" {
		if line == "" {
			return "", "", errors.New("invalid stdin input: first line is empty")
		}
		return line, end, nil
	}
	split := strings.Split(line, ",")
	if len(split) != 2 {
		return "", "", fmt.Errorf("invalid stdin input %q: %s", line, usage)
	}
	start := strings.TrimSpace(split[0])
	end = strings.TrimSpace(split[1])
	if start == "" || end == "" {
		return "", "", fmt.Errorf("invalid stdin input %q: %s", line, usage)
	}
	return start, end, nil
}

// convert duration from one group of units to another
func convDuration(source, target string, brief bool, decimals int) string {
	conv := DateTimeMate.NewConv(DateTimeMate.ConvWithSource(source), DateTimeMate.ConvWithTarget(target), DateTimeMate.ConvWithBrief(brief), DateTimeMate.ConvWithDecimals(decimals))
	result, err := conv.ConvertDuration()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return result
}

// outputDiff compute the duration between two dates, times, and/or date/times
func outputDiff(start, end string, brief bool) {
	if optDiffDecimals != 0 && optDiffConv == "" {
		fmt.Fprintln(os.Stderr, "-d/--decimals requires -c/--conv")
		os.Exit(1)
	}
	diff := DateTimeMate.NewDiff(DateTimeMate.DiffWithStart(start), DateTimeMate.DiffWithEnd(end), DateTimeMate.DiffWithBrief(brief), DateTimeMate.DiffWithAbsolute(optDiffAbsolute))
	result, duration, err := diff.CalculateDiff()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if optDiffConv != "" {
		// convert from the exact duration, not the human-readable string:
		// the formatted string truncates sub-unit remainders and durafmt
		// uses 365-day years while conv uses 365.25
		result = convDuration(fmt.Sprintf("%d nanoseconds", duration.Nanoseconds()), optDiffConv, optDiffBrief, optDiffDecimals)
	}
	if optRootNoNewline {
		fmt.Print(result)
	} else {
		fmt.Println(result)
	}
}
