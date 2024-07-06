package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/jftuga/DateTimeMate"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff [start] [end]",
	Short: "Output the difference between two date/times",
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
			start, end := getInput()
			outputDiff(start, end, optDiffBrief)
			return
		}
		outputDiff(args[0], args[1], optDiffBrief)
	},
}

var optDiffBrief bool
var optDiffReadFromStdin bool

func init() {
	rootCmd.AddCommand(diffCmd)
	diffCmd.Flags().BoolVarP(&optDiffBrief, "brief", "b", false, "output in brief format, such as: 1Y2M3D4h5m6s7ms8us9ns")
	diffCmd.Flags().BoolVarP(&optDiffReadFromStdin, "stdin", "i", false, "read from STDIN instead of using -s/-e")
}

// either read one line containing a comma, then split start and end on this
// or read two lines with start on line one and end on line two
func getInput() (string, string) {
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	line := input.Text()
	if strings.Contains(line, ",") {
		split := strings.Split(line, ",")
		if len(split) != 2 {
			fmt.Fprintf(os.Stderr, "invalid stdin input: %s\n", line)
			os.Exit(1)
		}
		return split[0], split[1]
	}
	input.Scan()
	end := input.Text()
	return line, end
}

// outputDiff compute the duration between two dates, times, and/or date/times
func outputDiff(start, end string, brief bool) {
	diff := DateTimeMate.NewDiff(DateTimeMate.DiffWithStart(start), DateTimeMate.DiffWithEnd(end), DateTimeMate.DiffWithBrief(brief))
	result, _, err := diff.CalculateDiff()
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
