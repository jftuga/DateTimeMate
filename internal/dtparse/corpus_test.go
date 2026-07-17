// corpus_test.go pins Parse's behavior over a large input corpus against a
// golden file, so any change to the layout table that alters an existing
// outcome (a new acceptance, a new rejection, or a different resolved value)
// shows up as a reviewable golden diff instead of slipping through. The
// corpus is the union of a generated cross-product of date shapes, time
// forms, and zone suffixes with the date/time-looking string literals
// harvested from this repo's tests and README (testdata/corpus.txt). It was
// introduced with the v1.19.0 library bug hunt, whose differential run
// against the pre-hunt table showed 0 regressions and 0 instant mismatches.
//
// After an intentional table change, regenerate with:
//
//	go test ./internal/dtparse -run TestCorpusGolden -update
//
// and review the golden diff line by line.
package dtparse

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"
)

var updateGolden = flag.Bool("update", false, "regenerate testdata/golden.txt from current Parse behavior")

// corpusDates, corpusTimes, and corpusZones form the generated part of the
// corpus: every date with every time form, and every date+time with every
// zone suffix; invalid shapes are included deliberately so their rejection
// is pinned too
var (
	corpusDates = []string{
		"2024-1-2", "2024-01-02", "2024.1.2", "2024.01.02", "2024/1/2", "2024/01/02",
		"Jan 2, 2024", "January 2, 2024", "Mon, Jan 2, 2024", "Tue, Jan 2, 2024",
		"2-Jan-2024", "22-Jul-2024", "1950-1-2", "Feb 29, 2024", "Feb 29, 2023",
		"2024-02-30", "2024-13-01", "2024-00-15", "Jan 32, 2024",
	}
	corpusTimes = []string{
		"", " 08:30", " 8:30", " 08:30:45", " 15:04", " 15:04:05", " 23:59:59",
		" 3:04PM", " 3:04 PM", " 3:04pm", " 3:04 pm",
		" 3:04:05PM", " 3:04:05 PM", " 3:04:05pm", " 3:04:05 pm",
		" 11:00AM", " 11:00 am", " 12:00:00", " 13:04PM", " 13:04 PM",
		" 25:00", " 08:61", " 08:30:61", " 08:30:45.123", " 08:30:45.123456789",
		"T08:30", "T08:30:45", " 0:30", " 12:30AM", " 12:30 PM",
	}
	corpusZones = []string{" EST", " EDT", " MST", " XYZ", " UTC", " -0500", " +0545", " +05", " Z"}
	corpusExtra = []string{
		"2024-01", "2024-1", "2024.1", "2024/1", "2024-10", "2024-13",
		"08:30", "8:30", "15:16:15", "3:04PM", "3:04 PM", "3:04pm", "3:04 pm",
		"3:04:05PM", "3:04:05 PM", "11:00AM", "11:00 AM", "11:00PM", "12:34:56.1234",
		"13:04PM", "13:04 PM", "25:00", "08:61", "0:30",
		"Mon Jan  2 15:04:05 2006", "Jan 2 15:04:05 2024", "Jan 15 12:00:00 EDT 2026",
		"Jan 15 12:00:00 EST 2026", "Jan 2 15:04:05 XYZ 2024",
		"02 Jan 06 15:04 MST", "02 Jan 06 15:04 -0700", "02 Jan 24 15:04 EST",
		"Monday, 02-Jan-06 15:04:05 MST", "Monday, 02-Jan-2024 15:04:05 GMT",
		"Mon, 02 Jan 06 15:04:05 -0700", "Mon, 02 Jan 24 15:04:05 -0700",
		"2024-01-15T08:30:00-0500", "2024-01-15T08:30:00+08", "2024-01-15T08:30:00Z",
		"2024-01-02T15:04:05+05", "2024-01-15 17:45:00 +0545 +0545",
		"2024-01-15 12:00:00 -0700 -0700", "Jan 2 2024", "2 Jan 2024",
		"not a date", "hello world", "1/2", "12/31/24", "1/2/2024", "",
		"2024", "20240101", "20240101080102", "1700265600",
	}
)

// corpusInputs returns the sorted, deduplicated corpus: the generated
// cross-product plus the harvested literals in testdata/corpus.txt
func corpusInputs(t *testing.T) []string {
	t.Helper()
	var inputs []string
	for _, d := range corpusDates {
		for _, tm := range corpusTimes {
			inputs = append(inputs, d+tm)
			if tm != "" {
				for _, z := range corpusZones {
					inputs = append(inputs, d+tm+z)
				}
			}
		}
	}
	inputs = append(inputs, corpusExtra...)
	harvested, err := os.ReadFile("testdata/corpus.txt")
	if err != nil {
		t.Fatalf("read testdata/corpus.txt: %v", err)
	}
	for _, line := range strings.Split(string(harvested), "\n") {
		if line != "" && !strings.Contains(line, "\t") {
			inputs = append(inputs, line)
		}
	}
	seen := make(map[string]bool, len(inputs))
	var corpus []string
	for _, s := range inputs {
		if !seen[s] {
			seen[s] = true
			corpus = append(corpus, s)
		}
	}
	sort.Strings(corpus)
	return corpus
}

// corpusOutcome encodes one Parse result deterministically for any test
// environment: clock only for time-only results (the stamped date is
// always today), and bare "ZONED" for zoned results, because time.Parse
// resolves a zone abbreviation only when the machine's local zone defines
// it, shifting both the offset and the rendered wall clock (e.g.
// "Jan 15 12:00:00 EDT 2026" renders as 11:00 EST where EDT is local);
// the zoned instant semantics are validated at the root-package layer
func corpusOutcome(source string) string {
	t, kind, err := Parse(source, time.UTC)
	if err != nil {
		return "ERR"
	}
	switch kind {
	case KindWallClock:
		return "WALL " + t.Format("2006-01-02 15:04:05.999999999")
	case KindTimeOnly:
		return "TIME " + t.Format("15:04:05.999999999")
	default:
		return "ZONED"
	}
}

// TestCorpusGolden compares every corpus input's outcome to
// testdata/golden.txt; run with -update to regenerate the golden file after
// an intentional layout-table change
func TestCorpusGolden(t *testing.T) {
	t.Parallel()
	corpus := corpusInputs(t)

	if *updateGolden {
		var b strings.Builder
		for _, source := range corpus {
			fmt.Fprintf(&b, "%s\t%s\n", corpusOutcome(source), source)
		}
		if err := os.WriteFile("testdata/golden.txt", []byte(b.String()), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("regenerated testdata/golden.txt with %d entries", len(corpus))
		return
	}

	raw, err := os.ReadFile("testdata/golden.txt")
	if err != nil {
		t.Fatalf("read testdata/golden.txt (run with -update to create it): %v", err)
	}
	golden := make(map[string]string, len(corpus))
	for _, line := range strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n") {
		expected, source, found := strings.Cut(line, "\t")
		if !found {
			t.Fatalf("malformed golden line: %q", line)
		}
		golden[source] = expected
	}

	if len(golden) != len(corpus) {
		t.Errorf("golden has %d entries but the corpus has %d; run with -update and review the diff", len(golden), len(corpus))
	}
	for _, source := range corpus {
		expected, ok := golden[source]
		if !ok {
			t.Errorf("corpus input %q missing from golden; run with -update and review the diff", source)
			continue
		}
		if got := corpusOutcome(source); got != expected {
			t.Errorf("Parse(%q) = %q, golden pins %q", source, got, expected)
		}
	}
}
