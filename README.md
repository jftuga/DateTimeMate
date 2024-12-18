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
* * `1 year 2 months 3 days 4 hours 5 minutes 6 second 7 milliseconds 8 microseconds 9 nanoseconds or 1Y2M3D4h5m6s7ms8us9ns`
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
<summary>5. Reformat a date/time</summary>

* convert the output of the `date` utility: `dtmate fmt "$(date)" "%F %T"`
* * where `($date)` equals `Mon Jul 22 22:49:18 EDT 2024`
* * output: 2024-07-22 22:49:18
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


See also the [example](cmd/example/main.go) program.


## Command Line Usage
<details>

<summary>Show</summary>

```
dtmate: output the difference between date, time or duration

Usage:
  dtmate [flags]
  dtmate [command]

Available Commands:
  conv        Convert a duration from group of units to another
  diff        Output the difference between two date/times
  dur         Output a date/time when given a starting date/time and duration
  fmt         reformat a date/time
  help        Help about any command

Flags:
  -e, --examples    show command-line examples
  -h, --help        help for dtmate
  -n, --nonewline   do not output a newline character
  -v, --version     version for dtmate

Use "dtmate [command] --help" for more information about a command.

---

Durations:
years months weeks days
hours minutes seconds milliseconds microseconds nanoseconds
example: '1 year 2 months 3 days 4 hours 1 minute 6 seconds'

---

Brief Durations:
(dates are always uppercase, times are always lowercase)
Y    M    W    D
h    m    s    ms    us    ns
examples: 1Y2M3W4D5h6m7s8ms9us1ns, '1Y 2M 3W 4D 5h 6m 7s 8ms 9us 1ns'

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
```

</details>

**Note:** The `-i` switch can accept two different types of input:

1. one line with start and end separated by a comma
2. two lines with start on the first line and end on the second line

**Note:** The `-n` switch along with `-r` will emit a comma-delimited output
* * Example: `dtmate dur now 1h -a -n -r 3`

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
$ dtmate diff 2024-06-07T08:00:00Z 2024-06-07T08:05:05-05:00 -c ms
305 minutes 5 seconds

# differentiate sub-second durations with a dot
# note the "ms" on both sides of the dot: minutes & seconds vs milliseconds
$ diff now "2020-01-01 11:12:13.123456789" -c ms.msusns
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
# can also use "years", "months", "weeks", "days"
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

########################### "dtmate conv" examples ###########################

# convert from one group of date/time units to another
$ dtmate conv 25771401s WDhms
42 weeks 4 days 6 hours 43 minutes 21 seconds

# another conversion, in the opposite direction, brief output
$ dtmate conv 42W4D6h43m21s seconds -b
25771401s

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
```
</details>

## LICENSE

[MIT LICENSE](LICENSE)

## Acknowledgements

<details>
<summary>Imported Modules</summary>

* carbon - https://github.com/golang-module/carbon
* cobra - https://github.com/spf13/cobra
* durafmt - https://github.com/hako/durafmt
* parsetime - https://github.com/tkuchiki/parsetime
* strftime - https://github.com/lestrrat-go/strftime

</details>

## Disclosure Notification

This program is my own original idea and was completely developed
on my own personal time, for my own personal benefit, and on my
personally owned equipment.

