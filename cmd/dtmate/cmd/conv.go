package cmd

import (
	"fmt"
	"github.com/jftuga/DateTimeMate"
	"os"

	"github.com/spf13/cobra"
)

var optConvBrief bool

var convCmd = &cobra.Command{
	Use:   "conv [source duration] [target duration]",
	Short: "Convert a duration from group of units to another",
	Args:  cobra.MatchAll(cobra.ExactArgs(2)),
	Run: func(cmd *cobra.Command, args []string) {
		outputConvDuration(args[0], args[1], optConvBrief)
	},
}

func init() {
	rootCmd.AddCommand(convCmd)
	convCmd.Flags().BoolVarP(&optConvBrief, "brief", "b", false, "output in brief format, such as: 1Y2M3D4h5m6s7ms8us9ns")
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
