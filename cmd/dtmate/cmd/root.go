package cmd

import (
	"fmt"
	DateTimeMate "github.com/jftuga/DateTimeMate"
	"github.com/spf13/cobra"
	"os"
	"regexp"
)

const extendedHelp string = `
---

Durations:
years weeks days
hours minutes seconds milliseconds microseconds nanoseconds
example: '1 year 3 days 4 hours 1 minute 6 seconds'

---

Brief Durations:
(dates are always uppercase, times are always lowercase)
Y    W    D
h    m    s    ms    us    ns
examples: 1Y3W4D5h6m7s8ms9us1ns, '1Y 3W 4D 5h 6m 7s 8ms 9us 1ns'

---

Relative Date Shortcuts:
now
today (returns same value as now)
yesterday (exactly 24 hours behind of the current time)
tomorrow (exactly 24 hours ahead the current time)
example: dtmate dur today 7h10m -a -u tomorrow

---

Conversions:
1 year is equal to 365.25 days
Months are not a unit since their lengths vary between 28 and 31 days
Separate sub-second brief units with a dot
example: dtmate conv 4321s123456789ns hms.msusns
`

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "dtmate",
	Short:   "dtmate: output the difference between date, time or duration",
	Version: DateTimeMate.ModVersion,
	Run: func(cmd *cobra.Command, args []string) {
		if optRootShowExamples {
			ShowExamples()
			os.Exit(0)
		}

		// if no arguments are provided and no flags are set, show help
		if len(args) == 0 && !cmd.Flags().Changed("examples") {
			cmd.Help() //nolint:errcheck
		}
	},
}

var optRootNoNewline bool
var optRootShowExamples bool
var readmeExamplesRegex = regexp.MustCompile(`(?ms)## Command Line Examples.*?shell\n(.*?)` + "```")

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&optRootNoNewline, "nonewline", "n", false, "do not output a newline character")
	rootCmd.Flags().BoolVarP(&optRootShowExamples, "examples", "e", false, "show command-line examples")

	versionTemplate := fmt.Sprintf("dtmate version %s\n%s\n", DateTimeMate.ModVersion, DateTimeMate.ModUrl)
	rootCmd.SetVersionTemplate(versionTemplate)
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetUsageTemplate(rootCmd.UsageTemplate() + extendedHelp)
}

func extractReadmeExamples(markdown string) string {
	matches := readmeExamplesRegex.FindStringSubmatch(markdown)
	if len(matches) == 2 {
		return matches[1]
	}
	return "[Internal Error] Unable to output examples. Check the extractShellCode() function."
}

func ShowExamples() {
	fmt.Println(extractReadmeExamples(DateTimeMate.ReadmeMd))
}
