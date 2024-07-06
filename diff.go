package DateTimeMate

import (
	"fmt"
	"github.com/golang-module/carbon/v2"
	"github.com/hako/durafmt"
	"github.com/tkuchiki/parsetime"
	"time"
)

type Diff struct {
	Start string
	End   string
	Brief bool
}

type OptionsDiff func(*Diff)

func NewDiff(options ...OptionsDiff) *Diff {
	diff := &Diff{}
	for _, opt := range options {
		opt(diff)
	}
	return diff
}

func DiffWithStart(start string) OptionsDiff {
	return func(opt *Diff) {
		opt.Start = start
	}
}

func DiffWithEnd(end string) OptionsDiff {
	return func(opt *Diff) {
		opt.End = end
	}
}

func DiffWithBrief(brief bool) OptionsDiff {
	return func(opt *Diff) {
		opt.Brief = brief
	}
}

func (diff *Diff) String() string {
	return fmt.Sprintf("start:%v end:%v brief:%v", diff.Start, diff.End, diff.Brief)
}

// CalculateDiff return the time difference and also set dt.Diff
// first try to parse with carbon, fallback to parsing with now if carbon fails to parse
func (diff *Diff) CalculateDiff() (string, time.Duration, error) {
	var start, end time.Time

	alpha := carbon.Parse(convertRelativeDateToActual(diff.Start))
	if alpha.Error != nil {
		// fmt.Println("alpha:", alpha.Error)
		p, err := parsetime.NewParseTime()
		if err != nil {
			return "", 0, err
		}
		start, err = p.Parse(diff.Start)
		if err != nil {
			return "", 0, err
		}
	} else {
		start = alpha.StdTime()
	}

	omega := carbon.Parse(convertRelativeDateToActual(diff.End))
	if omega.Error != nil {
		// fmt.Println("omega:", omega.Error)
		p, err := parsetime.NewParseTime()
		if err != nil {
			return "", 0, err
		}
		end, err = p.Parse(diff.End)
		if err != nil {
			return "", 0, err
		}
	} else {
		end = omega.StdTime()
	}

	duration := end.Sub(start)
	parsed := durafmt.Parse(duration)
	difference := fmt.Sprintf("%v", parsed)
	if diff.Brief {
		difference = shrinkPeriod(difference)
	}
	return difference, duration, nil
}
