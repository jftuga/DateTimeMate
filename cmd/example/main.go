package main

import (
	"fmt"
	"github.com/jftuga/DateTimeMate"
	"os"
)

func main() {
	fmt.Println()

	start := "2024-06-01"
	end := "2024-08-05 00:01:02"
	brief := true
	diff := DateTimeMate.NewDiff(
		DateTimeMate.DiffWithStart(start),
		DateTimeMate.DiffWithEnd(end),
		DateTimeMate.DiffWithBrief(brief))
	fmt.Println(diff)
	fmt.Println("===================================================")

	result, duration, err := diff.CalculateDiff()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("diff:", diff)
	fmt.Println("duration:", result, "=>", duration)
	fmt.Println("===================================================")

	from := "2024-06-01"
	d := "1 year 7 days 6 hours 5 minutes"
	until := "2027-06-22 18:15:11"
	ofmt := "%Y%m%d.%H%M%S"
	dur := DateTimeMate.NewDur(
		DateTimeMate.DurWithFrom(from),
		DateTimeMate.DurWithDur(d),
		DateTimeMate.DurWithRepeat(0),
		DateTimeMate.DurWithUntil(until),
		DateTimeMate.DurWithOutputFormat(ofmt))
	fmt.Println("duration:", dur)
	fmt.Println("===================================================")

	add1, err := dur.Add()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Add1: ", add1)
	fmt.Println("===================================================")
	dur.Until = ""
	dur.Repeat = 3

	add2, err := dur.Add()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Add2: ", add2)
	fmt.Println("===================================================")

	until = "2020-05-02 23:41:00"
	ofmt = "%v %T"
	dur = DateTimeMate.NewDur(
		DateTimeMate.DurWithFrom(from),
		DateTimeMate.DurWithDur(d),
		DateTimeMate.DurWithUntil(until),
		DateTimeMate.DurWithOutputFormat(ofmt))
	sub1, err := dur.Sub()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("dur:", dur)
	fmt.Println("Sub1: ", sub1)

	fmt.Println("===================================================")
	source := "1367h29m13s"
	target := "Dhms" // days, hours, minutes, seconds
	conv := DateTimeMate.NewConv(
		DateTimeMate.ConvWithSource(source),
		DateTimeMate.ConvWithTarget(target))
	newDuration, err := conv.ConvertDuration()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("conv:", conv)
	fmt.Println("new duration:", newDuration)

	fmt.Println("===================================================")
	source = "Mon Jul 22 08:40:33 EDT 2024"
	outputFormat := "%F %T"

	newFormat, err := DateTimeMate.Reformat(source, outputFormat)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("new format:", newFormat)
}
