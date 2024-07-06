package cmd

import (
	"fmt"
	DateTimeMate "github.com/jftuga/DateTimeMate"
	"github.com/spf13/cobra"
	"os"
)

const extendedHelp string = `
Durations:
years months weeks days
hours minutes seconds milliseconds microseconds nanoseconds
example: '1 year 2 months 3 days 4 hours 1 minute 6 seconds'

Brief Durations: (dates are always uppercase, times are always lowercase)
Y    M    W    D
h    m    s    ms    us    ns
examples: 1Y2M3W4D5h6m7s8ms9us1ns, '1Y 2M 3W 4D 5h 6m 7s 8ms 9us 1ns'

Relative Date Shortcuts:
now
today (returns same value as now)
yesterday (exactly 24 hours ahead of the current time)
tomorrow (exactly 24 hours behind the current time)
example: dtmate dur today 7h10m -a -u tomorrow
`

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "dtmate",
	Short:   "dtmate: output the difference between date, time or duration",
	Version: DateTimeMate.ModVersion,
}

var optRootNoNewline bool

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().BoolVarP(&optRootNoNewline, "nonewline", "n", false, "do not output a newline character")
	versionTemplate := fmt.Sprintf("dtmate version %s\n%s\n", DateTimeMate.ModVersion, DateTimeMate.ModUrl)
	rootCmd.SetVersionTemplate(versionTemplate)
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetUsageTemplate(rootCmd.UsageTemplate() + extendedHelp)
}
