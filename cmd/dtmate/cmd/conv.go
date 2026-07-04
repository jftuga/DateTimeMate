package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jftuga/DateTimeMate"
	"os"

	"github.com/spf13/cobra"
)

var optConvBrief bool
var optConvDecimals int

var convCmd = &cobra.Command{
	Use:   "conv [source duration] [target duration]",
	Short: "Convert a duration from group of units to another",
	Example: `  dtmate conv 90m hm
  dtmate conv 4321s123456789ns hms.msusns`,
	Args:               cobra.ArbitraryArgs,
	DisableFlagParsing: true, // this allows for negative durations; flags are parsed manually in RunE
	RunE: func(cmd *cobra.Command, args []string) error {
		positional, brief, noNewline, help, decimals, err := parseConvArgs(args)
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
		optConvDecimals = decimals
		if noNewline {
			optRootNoNewline = true
		}
		outputConvDuration(positional[0], positional[1], optConvBrief, optConvDecimals)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(convCmd)
	convCmd.Flags().BoolVarP(&optConvBrief, "brief", "b", false, "output in brief format, such as: 1Y3W4D5h6m7s")
	convCmd.Flags().IntVarP(&optConvDecimals, "decimals", "d", 0, "show the smallest unit with this many decimal places, rounded")
}

// parseConvArgs manually separates flags from positional args because convCmd
// disables cobra flag parsing to support negative durations; an arg starting
// with "-" followed by a digit is a negative duration, not a flag
func parseConvArgs(args []string) (positional []string, brief, noNewline, help bool, decimals int, err error) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--":
			positional = append(positional, args[i+1:]...)
			return positional, brief, noNewline, help, decimals, nil
		case arg == "--brief":
			brief = true
		case arg == "--nonewline":
			noNewline = true
		case arg == "--help":
			help = true
		case arg == "--decimals":
			if i+1 >= len(args) {
				return nil, false, false, false, 0, fmt.Errorf("flag needs an argument: --decimals")
			}
			i++
			decimals, err = strconv.Atoi(args[i])
			if err != nil {
				return nil, false, false, false, 0, fmt.Errorf("invalid argument %q for --decimals", args[i])
			}
		case strings.HasPrefix(arg, "--decimals="):
			value := strings.TrimPrefix(arg, "--decimals=")
			decimals, err = strconv.Atoi(value)
			if err != nil {
				return nil, false, false, false, 0, fmt.Errorf("invalid argument %q for --decimals", value)
			}
		case strings.HasPrefix(arg, "--"):
			return nil, false, false, false, 0, fmt.Errorf("unknown flag: %s", arg)
		case len(arg) > 1 && arg[0] == '-' && (arg[1] < '0' || arg[1] > '9'):
			for j := 1; j < len(arg); j++ {
				switch arg[j] {
				case 'b':
					brief = true
				case 'n':
					noNewline = true
				case 'h':
					help = true
				case 'd':
					// 'd' takes a value: the rest of this arg (-d2) or the next arg (-d 2)
					value := arg[j+1:]
					if value == "" {
						if i+1 >= len(args) {
							return nil, false, false, false, 0, fmt.Errorf("flag needs an argument: 'd' in %s", arg)
						}
						i++
						value = args[i]
					}
					decimals, err = strconv.Atoi(value)
					if err != nil {
						return nil, false, false, false, 0, fmt.Errorf("invalid argument %q for -d", value)
					}
					j = len(arg)
				default:
					return nil, false, false, false, 0, fmt.Errorf("unknown shorthand flag: %q in %s", arg[j], arg)
				}
			}
		default:
			positional = append(positional, arg)
		}
	}
	return positional, brief, noNewline, help, decimals, nil
}

func outputConvDuration(source, target string, brief bool, decimals int) {
	conv := DateTimeMate.NewConv(DateTimeMate.ConvWithSource(source), DateTimeMate.ConvWithTarget(target), DateTimeMate.ConvWithBrief(brief), DateTimeMate.ConvWithDecimals(decimals))
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
