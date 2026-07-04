package cmd

import (
	"fmt"
	"strings"

	"github.com/jftuga/DateTimeMate"
	"os"

	"github.com/spf13/cobra"
)

var optConvBrief bool

var convCmd = &cobra.Command{
	Use:                "conv [source duration] [target duration]",
	Short:              "Convert a duration from group of units to another",
	Args:               cobra.ArbitraryArgs,
	DisableFlagParsing: true, // this allows for negative durations; flags are parsed manually in RunE
	RunE: func(cmd *cobra.Command, args []string) error {
		positional, brief, noNewline, help, err := parseConvArgs(args)
		if err != nil {
			return err
		}
		if help {
			return cmd.Help()
		}
		if len(positional) != 2 {
			return fmt.Errorf("accepts 2 arg(s), received %d", len(positional))
		}
		optConvBrief = brief
		if noNewline {
			optRootNoNewline = true
		}
		outputConvDuration(positional[0], positional[1], optConvBrief)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(convCmd)
	convCmd.Flags().BoolVarP(&optConvBrief, "brief", "b", false, "output in brief format, such as: 1Y2M3D4h5m6s7ms8us9ns")
}

// parseConvArgs manually separates flags from positional args because convCmd
// disables cobra flag parsing to support negative durations; an arg starting
// with "-" followed by a digit is a negative duration, not a flag
func parseConvArgs(args []string) (positional []string, brief, noNewline, help bool, err error) {
	for i, arg := range args {
		switch {
		case arg == "--":
			positional = append(positional, args[i+1:]...)
			return positional, brief, noNewline, help, nil
		case arg == "--brief":
			brief = true
		case arg == "--nonewline":
			noNewline = true
		case arg == "--help":
			help = true
		case strings.HasPrefix(arg, "--"):
			return nil, false, false, false, fmt.Errorf("unknown flag: %s", arg)
		case len(arg) > 1 && arg[0] == '-' && (arg[1] < '0' || arg[1] > '9'):
			for _, shorthand := range arg[1:] {
				switch shorthand {
				case 'b':
					brief = true
				case 'n':
					noNewline = true
				case 'h':
					help = true
				default:
					return nil, false, false, false, fmt.Errorf("unknown shorthand flag: %q in %s", shorthand, arg)
				}
			}
		default:
			positional = append(positional, arg)
		}
	}
	return positional, brief, noNewline, help, nil
}

func outputConvDuration(source, target string, brief bool) {
	conv := DateTimeMate.NewConv(DateTimeMate.ConvWithSource(source), DateTimeMate.ConvWithTarget(target), DateTimeMate.ConvWithBrief(brief))
	result, err := conv.ConvertDuration()
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
