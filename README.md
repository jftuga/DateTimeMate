# DateTimeMate
Golang package and CLI to compute the difference between date, time or duration

The command-line program, `dtmate` *(along with the golang package)* allows you to answer these inquiries:

<details open>
<summary>1. What is the duration between two different dates and/or times?</summary>

`dtmate diff "2024-06-01 11:22:33" "2024-07-19 21:07:19"`
* answer: `6 weeks 6 days 9 hours 44 minutes 46 seconds`
* answer with the `-b` option: `6W6D9h44m46s`
* start and end can be in various formats, such as:
* * `11:22:33`, `2024-06-01`, `"2024-06-01 11:22:33"`, `2024-06-01T11:22:33.456Z`
</details>

<details>
<summary>2. What is the datetime when adding or subtracting a duration?</summary>

`dtmate dur "2024-06-01 11:22:33" 6W6D9h44m46s -a`
* answer: `2024-04-14 01:37:47 -0400 EDT`
* answer with the `-f "%Y-%m-%d %H:%M:%S"` option: `2024-04-14 01:37:47`
* Duration examples include:
* * `5 minutes 5 seconds or 5m5s`
* * `3 weeks 4 days 5 hours or 3W4D5h`
* * `1 year 3 days 4 hours 5 minutes 6 seconds 7 milliseconds 8 microseconds 9 nanoseconds or 1Y3D4h5m6s7ms8us9ns`
</details>

<details>
<summary>3. Similar to previous question, but repeats a period multiple times or until a certain date/time is encountered.</summary>

* adding dates, repeat twice: `dtmate dur "2024-06-01 12:00:00" 1h5m10s -r 2 -a`
* subtracting until a date is exceeded: `dtmate dur "12:00:00" 1h5m10s -u "09:48" -s`
</details>

<details>
<summary>4. Convert from one group of date/time units to another</summary>

* convert from seconds to weeks, days, hours, minutes, seconds: `dtmate conv 25771401s WDhms`
* * 42 weeks 4 days 6 hours 43 minutes 21 seconds
* convert weeks, days, hours, minutes, seconds to just seconds, with brief output format: `dtmate conv "42 weeks 4 days 6 hours 43 minutes 21 seconds" seconds -b`
* * 25771401s
</details>

<details>
<summary>5. Add or subtract two durations, even when expressed in different units?</summary>

* add: `dtmate durmath "1 hour 30 minutes" "45 minutes" -a`
* * `2 hours 15 minutes`
* subtract, with a signed result when the second duration is larger: `dtmate durmath "45 minutes" "1 hour" -s`
* * `-15 minutes`
* same input, always absolute with the `-A` option: `dtmate durmath "45 minutes" "1 hour" -s -A`
* * `15 minutes`
</details>

<details>
<summary>6. Reformat a date/time</summary>

* convert the output of the `date` utility: `dtmate fmt "$(date)" "%F %T"`
* * where `($date)` equals `Mon Jul 22 22:49:18 EDT 2024`
* * output: 2024-07-22 22:49:18
</details>

<details>
<summary>7. Convert a date/time from one time zone to another?</summary>

`dtmate tz "2024-01-15 12:00:00 UTC" America/New_York`
* answer: `2024-01-15 07:00:00 -0500 EST`
* zones can be given in multiple styles:
* * IANA names such as `America/New_York`, `Asia/Kolkata`, `Australia/Eucla` *(preferred; these are DST aware and case-insensitive)*
* * abbreviations such as `EST`, `JST`, `pst` *(case-insensitive, fixed offsets)*
* * UTC offsets in seconds, such as `19800` for UTC+5:30
* the source may also be a unix timestamp in seconds or milliseconds, such as `1700265600`
* pin ambiguous abbreviations with an environment variable: `DTMATE_TZ_ALIASES="IST=Asia/Jerusalem|CST=Asia/Shanghai"`
* list all supported abbreviations: `dtmate tz --list-zones`
* list all IANA zone names with their current offsets: `dtmate tz --list-iana`
</details>

## Installation

* Library: `go get -u github.com/jftuga/DateTimeMate`
* Command line tool: `go install -ldflags="-s -w" github.com/jftuga/DateTimeMate/cmd/dtmate@latest`
* * Binaries for all platforms are provided in the [releases](https://github.com/jftuga/DateTimeMate/releases) section.
* Homebrew (MacOS / Linux):
* * `brew tap jftuga/homebrew-tap; brew update; brew install jftuga/tap/dtmate`

## Library Usage
<details open>
<summary>Example 1 - duration between two dates</summary>

Supported date time formats are listed in: https://go.dev/src/time/format.go

```golang
import "github.com/jftuga/DateTimeMate"

// example 1 - duration between two dates
start := "2024-06-01"
end := "2024-08-05 00:01:02"
brief := true
diff := DateTimeMate.NewDiff(DateTimeMate.DiffWithStart(start), DateTimeMate.DiffWithEnd(end),
	DateTimeMate.DiffWithBrief(brief))
result, duration, err := diff.CalculateDiff()
if err != nil { ... }
fmt.Println(result, duration)  // 9W2D1m2s 1560h1m2s
```
</details>

<details>
<summary>Example 2 - add a duration</summary>

```go
// example 2 - add a duration and repeat it until the "until" date is exceeded
from := "2024-06-01"
d := "1 year 7 days 6 hours 5 minutes"
until := "2027-06-22 18:15:11"
ofmt := "%Y%m%d.%H%M%S"
dur := DateTimeMate.NewDur(DateTimeMate.DurWithFrom(from), DateTimeMate.DurWithDur(d),
	DateTimeMate.DurWithRepeat(0), DateTimeMate.DurWithUntil(until),
	DateTimeMate.DurWithOutputFormat(ofmt))
add, err := dur.Add()
if err != nil { ... }
fmt.Println(add) // [20250608.060500 20260615.121000 20270622.181500]
```
</details>

<details>
<summary>Example 3 - convert date/time units</summary>

```go
source := "1367h29m13s"
target := "Dhms" // days, hours, minutes, seconds
conv := DateTimeMate.NewConv(
DateTimeMate.ConvWithSource(source),
DateTimeMate.ConvWithTarget(target))
newDuration, err := conv.ConvertDuration()
if err != nil { ... }
fmt.Println("new duration:", newDuration) // 56 days 23 hours 29 minutes 13 seconds
```
</details>

<details>
<summary>Example 4 - reformat a date/time</summary>

```go
source := "Mon Jul 22 08:40:33 EDT 2024"
outputFormat := "%F %T"
newFormat, err := DateTimeMate.Reformat(source, outputFormat)
if err != nil { ... }
fmt.Println("new format:", newFormat) // 2024-07-22 08:40:33
```
</details>

<details>
<summary>Example 5 - duration arithmetic</summary>

```go
first := "1 hour 30 minutes"
second := "45 minutes"
dm := DateTimeMate.NewDurMath(
	DateTimeMate.DurMathWithFirst(first),
	DateTimeMate.DurMathWithSecond(second))
sum, err := dm.Add()
if err != nil { ... }
fmt.Println(sum) // 2 hours 15 minutes
difference, err := dm.Sub()
if err != nil { ... }
fmt.Println(difference) // 45 minutes
```
</details>

<details>
<summary>Example 6 - convert between time zones</summary>

```go
conv := DateTimeMate.NewTimeZoneConverter(
	DateTimeMate.TimeZoneConverterWithZoneAbbrevs(DateTimeMate.LoadZoneDefinitions()))
result, err := conv.ConvertTimeZone("2024-01-15 12:00:00 UTC", "America/New_York")
if err != nil { ... }
fmt.Println(result.Format("2006-01-02 15:04:05 MST")) // 2024-01-15 07:00:00 EST

// pin ambiguous abbreviations to a specific IANA zone
aliases, err := DateTimeMate.ParseZoneAliases("IST=Asia/Jerusalem")
if err != nil { ... }
conv = DateTimeMate.NewTimeZoneConverter(
	DateTimeMate.TimeZoneConverterWithZoneAbbrevs(DateTimeMate.LoadZoneDefinitions()),
	DateTimeMate.TimeZoneConverterWithAliases(aliases))
result, err = conv.ConvertTimeZone("2024-07-15 12:00:00 UTC", "IST")
if err != nil { ... }
fmt.Println(result.Format("2006-01-02 15:04:05 MST")) // 2024-07-15 15:00:00 IDT
```
</details>


See also the [example](cmd/example/main.go) program.


## Command Line Usage
<details>

<summary>Show</summary>

```
Compute date/time differences, durations, conversions, and reformatting

Usage:
  dtmate [flags]
  dtmate [command]

Available Commands:
  conv        Convert a duration from group of units to another
  diff        Output the difference between two date/times
  dur         Output a date/time when given a starting date/time and duration
  durmath     Add or subtract two durations
  fmt         Reformat a date/time
  help        Help about any command
  tz          Convert a date/time from one time zone to another

Flags:
  -e, --examples    show command-line examples
  -h, --help        help for dtmate
      --help-all    show help plus duration syntax, brief units, and conversion notes
  -n, --nonewline   do not output a newline character
  -v, --version     version for dtmate

Use "dtmate [command] --help" for more information about a command.

Use "dtmate --help-all" for duration syntax, brief units, and conversion notes.
```

</details>

**Note:** The `-i` switch can accept two different types of input:

1. one line with start and end separated by a comma
2. two lines with start on the first line and end on the second line

**Note:** The `-n` switch along with `-r` will emit a comma-delimited output
* * Example: `dtmate dur now 1h -a -n -r 3`

## Date and Duration Parsing Notes

* **Supported input formats** are a fixed, documented list rather than
  fuzzy matching: ISO-style dates and date/times (padded or unpadded, `-`,
  `.`, or `/` separated, `T` or space before the time, optional fractional
  seconds, optional zone or offset), year and year-month forms (`2024`,
  `2024-01`), month-name dates (`Jan 2, 2024`, `January 2, 2024 08:30:00`,
  `2-Jan-2024 08:21:44`, ANSIC forms such as `Jan 2 15:04:05 2024` with
  optional weekday and zone), RFC822/850/1036/1123, Unix and Ruby date
  formats, slash dates, bare times of day (`08:30`, `3:04pm`, `11:00 AM`,
  `12:34:56.1234`, interpreted as today; am/pm may be joined to the time or
  separated by a space), Unix timestamps, and the relative
  words `now`, `today`, `yesterday`, and `tomorrow`. Inputs outside this
  list are rejected with an error instead of being guessed at.
* **Zone abbreviations inside date/times** (such as `EDT` in
  `Jan 15 12:00:00 EDT 2026`) are honored when the local time zone defines
  them; an unrecognized abbreviation is rejected rather than silently read
  as UTC. For arbitrary zone conversions, use `dtmate tz`, which resolves
  abbreviations through its own zone table.
* **Slash dates** default to US order, month first: `01/02/2024` is January 2.
* * Set `DTMATE_DATE_ORDER=DMY` for day/month/year, or `MDY` to silence the
    ambiguity warning; a field greater than 12 (such as `25/12/2024`)
    disambiguates on its own.
* * Two-digit years such as `1/2/24` follow the same order rules; years
    69-99 are 19xx and 00-68 are 20xx.
* **Out-of-range date/times** such as `2024-02-30` or `08:61:00` are
  rejected instead of being silently normalized, and empty input is
  rejected instead of being read as the current time.
* **Pure integers** parse by digit count: 10 digits are Unix seconds, 13 are
  Unix milliseconds, while 4, 8, and 14 digits are a year (`2024`), a compact
  date (`20240101`), and a compact date/time (`20240101080102`); 11, 12, and
  other digit counts are ambiguous and rejected.
* **Negative timestamps** are rejected everywhere; pre-1970 date/times are
  fully supported through normal date strings such as `1950-01-01`.
* **Relative dates**: `yesterday` and `tomorrow` are exactly 24 hours from
  now, even across daylight saving transitions.
* **Duration amounts** must be plain decimals (`90`, `1.5`, and mid-string
  negatives such as `1 year -30 days` in `conv`); `NaN`, `Inf`, exponent
  (`1e2`), and hex (`0x1p4`) forms are rejected.
* **Long-form unit names** are case-insensitive (`1 Hour` equals `1 hour`);
  brief units stay case-sensitive because `D` means days while `m` means
  minutes.
* **Zone abbreviations** such as `CET` or `EST` always mean their fixed UTC
  offsets, on any date; use an IANA name such as `Europe/Paris` (or a
  `DTMATE_TZ_ALIASES` alias) for DST-aware conversion.
* **Duration range and precision**: durations are computed in integer
  nanoseconds, so integral amounts are exact; fractional amounts carry
  float64 precision (about 15-16 significant digits); totals are limited to
  about +/-292 years.
* **Brief sub-second targets**: a lone `us` or `ns` target means that
  sub-second unit; a lone `ms` keeps its historical minutes+seconds meaning
  and warns on stderr (use `.ms` or `milliseconds` for milliseconds);
  combine larger and sub-second units with a dot, such as `ms.msusns`.
* **Repeat and until**: `-r` is capped at 1,000,000 results, and `-u` must
  lie in the direction of travel (after the start when adding, before it
  when subtracting).

## Command Line Examples

<details>
<summary>Show</summary>

```shell

########################### "dtmate diff" examples ###########################

# difference between two times on the same day
$ dtmate diff 12:00:00 15:30:45
3 hours 30 minutes 45 seconds

# same input, using brief output
$ dtmate diff 12:00:00 15:30:45 -b
3h30m45s

# using AM/PM and not 24-hour times
$ dtmate diff "11:00AM" "11:00PM"
12 hours

# using ISO-8601 dates
$ dtmate diff 2024-06-07T08:00:00Z 2024-06-08T09:02:03Z
1 day 1 hour 2 minutes 3 seconds

# same input, also convert to seconds only, brief format
$ dtmate diff 2024-06-07T08:00:00Z 2024-06-08T09:02:03Z --conv s -b
90123s

# using timezone offset
$ dtmate diff 2024-06-07T08:00:00Z 2024-06-07T08:05:05-05:00
5 hours 5 minutes 5 seconds

# same input, also convert duration to minutes and seconds
# a bare "ms" target warns on stderr because ms means milliseconds elsewhere
$ dtmate diff 2024-06-07T08:00:00Z 2024-06-07T08:05:05-05:00 -c ms
warning: target "ms" is ambiguous: interpreting as minutes+seconds; use ".ms" or "milliseconds" for milliseconds
305 minutes 5 seconds

# a dot selects sub-second units: .ms is milliseconds, no warning
$ dtmate diff 2024-06-07T08:00:00Z 2024-06-07T08:05:05-05:00 -c .ms
18305000 milliseconds

# convert to a single unit, showing 2 decimal places
# without -d, this would truncate to just: 2 years
$ dtmate diff 2023-10-17 2026-07-04 -c Y -d 2
2.71 years

# differentiate sub-second durations with a dot
# note the "ms" on both sides of the dot: minutes & seconds vs milliseconds
$ dtmate diff now "2020-01-01 11:12:13.123456789" -c ms.msusns
-2566445 minutes 40 seconds 876 milliseconds 542 microseconds 985 nanoseconds

# using a format which includes spaces
$ dtmate diff "2024-06-07 08:01:02" "2024-06-07 08:02"
58 seconds

# using the built-in MacOS date program and do not include a newline character
$ dtmate diff "$(date -R)" "$(date -v+1M -v+30S)" -n
1 minute 30 seconds%

# using the cross-platform date program, ending time starting first
$ dtmate diff "$(date)" 2020
-4 years 24 weeks 1 day 7 hours 21 minutes 53 seconds

# same input, using brief output
$ dtmate diff "$(date)" 2020 -b
-4Y24W1D7h21m53s

# ending time first yields a signed result
$ dtmate diff 15:30:45 12:00:00
-3 hours 30 minutes 45 seconds

# same input, always output an absolute (positive) duration
$ dtmate diff 15:30:45 12:00:00 -A
3 hours 30 minutes 45 seconds

# using microsecond formatting
$ dtmate diff 2024-06-07T08:00:00Z 2024-06-07T08:00:00.000123Z
123 microseconds

# using millisecond formatting, adding -b returns: 1m2s345ms
$ dtmate diff 2024-06-07T08:00:00Z 2024-06-07T08:01:02.345Z
1 minute 2 seconds 345 milliseconds

# read from STDIN in CSV format and do not include a newline character
$ dtmate diff -i -n
15:16:15,15:17
45 seconds%

# same as above, include newline character
$ echo 15:16:15,15:17 | dtmate diff -i
45 seconds

# read from STDIN with start on first line and end on second line
$ printf "15:16:15\n15:17:20" | dtmate diff -i
1 minute 5 seconds

# use relative start date with brief output
$ dtmate diff today 2024-07-07 -b
3D16h38m47s

########################### "dtmate dur" examples ###########################

# add time
# can also use "years", "weeks", "days"
$ dtmate dur 2024-01-01 "1 hour 30 minutes 45 seconds" -a
2024-01-01 01:30:45 -0500 EST

# subtract time
# can also use "milliseconds", "microseconds"
$ dtmate dur "2024-01-02 01:02:03" "1 day 1 hour 2 minutes 3 seconds" -s
2024-01-01 00:00:00 -0500 EST

# output multiple occurrences: add 5 weeks, for 3 intervals
$ dtmate dur "2024-01-02" "5W" -r 3 -a
2024-02-06 00:00:00 -0500 EST
2024-03-12 00:00:00 -0400 EDT
2024-04-16 00:00:00 -0400 EDT

# repeat until a certain datetime is encountered: subtract 5 minutes until 15:00
$ dtmate dur 15:20 5m -u 15:00 -s
2024-06-30 15:15:00 -0400 EDT
2024-06-30 15:10:00 -0400 EDT
2024-06-30 15:05:00 -0400 EDT
2024-06-30 15:00:00 -0400 EDT

# use relative date until tomorrow
$ dtmate dur today 7h10m -u tomorrow -a
2024-07-03 14:29:28 -0400 EDT
2024-07-03 21:39:28 -0400 EDT
2024-07-04 04:49:28 -0400 EDT

# set the output format
$ dtmate dur "2024-07-01 12:00:00" 1W2D3h4m5s -a -f "%Y%m%d.%H%M%S"
20240710.150405

# unix (epoch) timestamps are accepted: 10 digits for seconds, 13 for milliseconds
$ dtmate dur 1700265600 "1 day" -a
2023-11-18 19:00:00 -0500 EST

# combine with -f "%s" to also output unix time
$ dtmate dur 1700265600 "1 day" -a -f "%s"
1700352000

########################### "dtmate durmath" examples ###########################

# add two durations expressed in different units
$ dtmate durmath "1 hour 30 minutes" "45 minutes" -a
2 hours 15 minutes

# subtract the second duration from the first
$ dtmate durmath "1 hour 30 minutes" "45 minutes" -s
45 minutes

# brief input and output
$ dtmate durmath 1h30m 45m -a -b
2h15m

# results are signed when the second duration is larger
$ dtmate durmath "45 minutes" "1 hour" -s
-15 minutes

# same input, always output an absolute (positive) duration
$ dtmate durmath "45 minutes" "1 hour" -s -A
15 minutes

# mixed units between the two durations
$ dtmate durmath "1 week" "3 days 12 hours" -s
3 days 12 hours

# convert the result to specific target units
$ dtmate durmath "1 day" "90 minutes" -s -c minutes
1350 minutes

# show the smallest unit with decimal places, rounded
$ dtmate durmath "1 hour" "30 minutes" -s -c hours -d 1
0.5 hours

# sub-second units appear only when the result needs them
$ dtmate durmath "1.5 seconds" "250 milliseconds" -s
1 second 250 milliseconds

########################### "dtmate conv" examples ###########################

# convert from one group of date/time units to another
$ dtmate conv 25771401s WDhms
42 weeks 4 days 6 hours 43 minutes 21 seconds

# another conversion, in the opposite direction, brief output
$ dtmate conv 42W4D6h43m21s seconds -b
25771401s

# show the smallest unit with decimal places, rounded
$ dtmate conv "1 hour 30 minutes" hours -d 1
1.5 hours

########################### "dtmate fmt" examples ###########################

# reformat date/times
$ dtmate fmt "2024-07-22 08:21:44" "%T %D"
08:21:44 07/22/24

$ dtmate fmt "2024-07-22 08:21:44" "%v %r"
22-Jul-2024 08:21:44 AM

$ dtmate fmt "2024-07-22 08:21:44" "%Y%m%d.%H%M%S"
20240722.082144

$ dtmate fmt "2024-02-29T23:59:59Z" "%Y%m%d.%H%M%S"
20240229.235959

$ dtmate fmt "2024-02-29T23:59:59Z" "%Z"
UTC

$ dtmate fmt "Mon Jul 22 08:40:33 EDT 2024" "%Z %z"
EDT -0400

# convert to unix (epoch) time seconds
$ dtmate fmt "2024-11-16 14:01:02" "%s"
1731783662

# from unix (epoch) time seconds
$ dtmate fmt 1704085262 "%F %T"
2024-01-01 00:01:02

# also from milliseconds
$ dtmate fmt 1704085262999 "%F %T"
2024-01-01 00:01:02

# compact integer date/times: 4, 8, or 14 digits
$ dtmate fmt 20240101080102 "%F %T"
2024-01-01 08:01:02

# ambiguous slash dates default to month/day/year and warn on stderr
$ dtmate fmt 01/02/2024 "%F"
warning: "01/02/2024" is ambiguous: interpreting as month/day/year; set DTMATE_DATE_ORDER=DMY to override
2024-01-02

# pin the order with an environment variable
$ DTMATE_DATE_ORDER=DMY dtmate fmt 01/02/2024 "%F"
2024-02-01

########################### "dtmate tz" examples ###########################

# convert using IANA zone names (preferred; these are DST aware)
$ dtmate tz "2024-01-15 12:00:00 UTC" America/New_York
2024-01-15 07:00:00 -0500 EST

# the same source in July automatically yields daylight time
$ dtmate tz "2024-07-04 08:00:00 EDT" Europe/Paris
2024-07-04 14:00:00 +0200 CEST

# abbreviations work for both the source and the target
$ dtmate tz "2024-01-15 09:00:00 PST" JST
2024-01-16 02:00:00 +0900 JST

# abbreviations and IANA names are case-insensitive
$ dtmate tz "2024-01-15 12:00:00 UTC" jst
2024-01-15 21:00:00 +0900 JST

# a zone-less source is interpreted as local time
$ dtmate tz "2024-01-15 12:00:00" UTC
2024-01-15 17:00:00 +0000 UTC

# a unix timestamp in seconds or milliseconds also works as the source
$ dtmate tz "1700265600" UTC
2023-11-18 00:00:00 +0000 UTC

# a UTC offset in seconds is also accepted (19800 = UTC+5:30)
$ dtmate tz "2024-01-15 12:00:00 UTC" 19800
2024-01-15 17:30:00 +0530 UTC+05:30

# ambiguous abbreviations warn on stderr and use their primary meaning
$ dtmate tz "2024-01-15 12:00:00 UTC" IST
warning: IST is ambiguous: using India Standard Time (UTC+05:30), not Israel Standard Time (UTC+2), Irish Standard Time (UTC+1); set DTMATE_TZ_ALIASES="IST=<IANA zone>" to override
2024-01-15 17:30:00 +0530 IST

# pin an ambiguous abbreviation to an IANA zone; aliases stay DST aware
$ DTMATE_TZ_ALIASES="IST=Asia/Jerusalem" dtmate tz "2024-07-15 12:00:00 UTC" IST
2024-07-15 15:00:00 +0300 IDT

# multiple aliases are pipe-delimited
$ DTMATE_TZ_ALIASES="IST=Asia/Jerusalem|CST=Asia/Shanghai" dtmate tz "2024-01-15 12:00:00 UTC" CST
2024-01-15 20:00:00 +0800 CST

# list the supported abbreviations
$ dtmate tz --list-zones
ACDT   UTC+10:30  Australian Central Daylight Time
ACST   UTC+09:30  Australian Central Standard Time
ACWST  UTC+08:45  Australian Central Western Standard Time
...

# list the IANA zone names with the offset currently in effect there
$ dtmate tz --list-iana
offsets and abbreviations are those currently in effect (2026-07-07)
Africa/Abidjan                   UTC+00:00 (GMT)
Africa/Accra                     UTC+00:00 (GMT)
...
America/New_York                 UTC-04:00 (EDT)
...
Europe/London                    UTC+01:00 (BST)
Europe/Paris                     UTC+02:00 (CEST)
...

# combine with grep to find a zone
$ dtmate tz --list-iana | grep -i sydney
Australia/Sydney                 UTC+10:00 (AEST)

# date/times before 1970 are rejected by default because time zone
# data is unreliable before then; use --force to convert anyway
$ dtmate tz --force "1900-02-28 23:59:59 UTC" Europe/London
1900-02-28 23:59:59 +0000 GMT
```
</details>

## LICENSE

[MIT LICENSE](LICENSE)

## Acknowledgements

<details>
<summary>Imported Modules</summary>

* cobra - https://github.com/spf13/cobra
* strftime - https://github.com/lestrrat-go/strftime

The fallback parser's layout table (`internal/dtparse`) is partly derived
from the layout list in carbon - https://github.com/golang-module/carbon
(MIT License).

</details>

## Disclosure Notification

This program is my own original idea and was completely developed
on my own personal time, for my own personal benefit, and on my
personally owned equipment.

