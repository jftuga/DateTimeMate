package cmd

import (
	"fmt"
	DateTimeMate "github.com/jftuga/DateTimeMate"
	"github.com/spf13/cobra"
	"os"
	"regexp"
)

const extendedHelp string = `
DURATION UNITS
  years  weeks  days  hours  minutes  seconds  milliseconds  microseconds  nanoseconds
  example: '1 year 3 days 4 hours 1 minute 6 seconds'

BRIEF UNITS  (date units uppercase, time units lowercase)
  Y  W  D          years, weeks, days
  h  m  s          hours, minutes, seconds
  ms us ns         milliseconds, microseconds, nanoseconds
  examples: 1Y3W4D5h6m7s8ms9us1ns  or  '1Y 3W 4D 5h 6m 7s'

RELATIVE DATE SHORTCUTS
  now, today       the current time
  yesterday        exactly 24 hours before now
  tomorrow         exactly 24 hours after now
  example: dtmate dur today 7h10m -a -u tomorrow

CONVERSION NOTES
  1 year equals 365.25 days
  months are not a unit; their lengths vary between 28 and 31 days
  separate sub-second brief target units with a dot:
    dtmate conv 4321s123456789ns hms.msusns
  use -d to round the smallest unit to N decimal places:
    dtmate diff 2023-10-17 2026-07-04 -c Y -d 2  =>  2.71 years
`

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "dtmate",
	Short:   "Compute date/time differences, durations, conversions, and reformatting",
	Version: DateTimeMate.ModVersion,
	Run: func(cmd *cobra.Command, args []string) {
		if optRootShowExamples {
			ShowExamples()
			os.Exit(0)
		}
		if optRootHelpAll {
			cmd.Help() //nolint:errcheck
			fmt.Print(extendedHelp)
			os.Exit(0)
		}

		// if no arguments are provided and no flags are set, show help
		if len(args) == 0 && !cmd.Flags().Changed("examples") && !cmd.Flags().Changed("help-all") {
			cmd.Help() //nolint:errcheck
		}
	},
}

var optRootNoNewline bool
var optRootShowExamples bool
var optRootHelpAll bool
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
	rootCmd.Flags().BoolVar(&optRootHelpAll, "help-all", false, "show help plus duration syntax, brief units, and conversion notes")

	versionTemplate := fmt.Sprintf("dtmate version %s\n%s\n", DateTimeMate.ModVersion, DateTimeMate.ModUrl)
	rootCmd.SetVersionTemplate(versionTemplate)
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetUsageTemplate(rootCmd.UsageTemplate() + "\nUse \"dtmate --help-all\" for duration syntax, brief units, and conversion notes.\n")
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
