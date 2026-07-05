# Known bugs in the Go module programming interface (as of v1.8.0)

**Status update (v1.11.0):** Both bugs are resolved. Bug 1's brief-expansion overwrite had already been fixed by the `parseDurationSeconds()` refactor, and the remaining sign-strip mutation in `ConvertDuration()` now operates on a local copy of `Source` (regression test: `TestConvReuse` in `conv_test.go`). Bug 2 was resolved by removing the `Dur.Op` field and unexporting the `Add`/`Sub` constants (now internal `opAdd`/`opSub`); `String()` no longer prints an op. The removal is technically API-breaking, but the field was a silent no-op and nothing external is known to set it, so it shipped as a minor bump (1.11.0) rather than a /v2 major.

These two issues affect only consumers of the `github.com/jftuga/DateTimeMate` Go module. The `dtmate` CLI never hits them because it constructs each object once, uses it once, and exits. Both were found during the v1.8.0 bug hunt (PR #36) by exercising the library as an external module consumer; neither is covered by the existing test suite. Work on branch `bug-hunt-2`.

## Bug 1: Conv.ConvertDuration() mutates its receiver, silently losing the negative sign on reuse

### Severity

High for library users: silent wrong result, same class as the P1 bugs fixed in v1.8.0.

### Reproduction

```go
package main

import (
	"fmt"
	DateTimeMate "github.com/jftuga/DateTimeMate"
)

func main() {
	conv := DateTimeMate.NewConv(
		DateTimeMate.ConvWithSource("-90m"),
		DateTimeMate.ConvWithTarget("hours"))
	r1, _ := conv.ConvertDuration()
	r2, _ := conv.ConvertDuration()
	fmt.Println(r1) // "-1 hour"  correct
	fmt.Println(r2) // "1 hour"   WRONG: the negative sign is gone
}
```

Any reuse of a `*Conv` is affected. The second and subsequent calls on a negative source return a positive result. A brief source (e.g. "90m") is also rewritten in place to its expanded long form ("90 minutes"), which happens to still parse correctly, so the sign loss is the observable wrong result; the in-place rewriting is the root defect.

### Root cause

`ConvertDuration()` in `conv.go` writes back to the struct instead of working on a local copy:

- `conv.go:273-274` — `strings.CutPrefix(conv.Source, "-")` strips the leading minus and assigns the stripped string back to `conv.Source`. The `isNegativeDuration` flag that remembers the sign is a local variable, so the sign is permanently lost from the struct after the first call.
- `conv.go:283` — after brief expansion, `conv.Source = expandedSource` overwrites the original source with the expanded long form.
- `parseSource()` (`conv.go:104`) reads `conv.Source` from the struct, which is why the mutations were introduced in the first place.

### Suggested fix

Make `ConvertDuration()` operate on a local `source` string. Two mechanical options; option A is less invasive:

- Option A: change `parseSource()` to accept the source as a parameter, `func (conv *Conv) parseSource(source string) (float64, error)`, and thread a local variable through `ConvertDuration()` (local `source := conv.Source`, strip the minus into `isNegativeDuration`, expand brief format into the local, never assign to `conv.Source`).
- Option B: keep signatures and save/restore `conv.Source` with a `defer`. Rejected style-wise: it keeps the mutation window open and is fragile.

Audit the sibling types for the same pattern while there: `Diff.CalculateDiff()` (diff.go) and `Dur.addOrSub()` (dur.go) do NOT mutate their receivers — verified during the hunt, and `dur_test.go` already reuses one `Dur` for both `Add()` and `Sub()`.

### Regression test

Add to `conv_test.go`, following the existing `testConv` helper style: construct one `Conv` with source "-90m" target "hours", call `ConvertDuration()` twice, assert both calls return "-1 hour". Also assert a second call with a brief positive source ("90m" -> "1 hour") still works, to cover the brief-expansion overwrite.

## Bug 2: Dur.Op is an exported field that is silently ignored

### Severity

Low-to-medium: no wrong arithmetic by itself, but it is a misleading API contract that invites wrong code.

### Reproduction

```go
dur := DateTimeMate.Dur{From: "2024-01-01", Op: DateTimeMate.Sub, Period: "1 day"}
res, _ := dur.Add()
fmt.Println(res) // [2024-01-02 ...] — it ADDED, even though Op says Sub
```

### Root cause

- `dur.go:22` — `Dur` has an exported `Op int` field, and `dur.go:16-18` exports the `Add`/`Sub` constants that look like its intended values.
- `Dur.Add()` and `Dur.Sub()` pass their own hardcoded op constant to `addOrSub()` and never read `dur.Op`. Nothing else reads it either, except `String()` (`dur.go:108`) which prints it, reinforcing the impression that it matters.
- There is also no `DurWithOp()` functional option, unlike every other `Dur` field.

### Decision needed before fixing

Pick one, in descending order of preference:

1. Wire it up without breaking anyone: add a `Compute() ([]string, error)` method that dispatches on `dur.Op` (`Add` -> add, `Sub` -> subtract, anything else -> error), add a `DurWithOp(op int)` option, and have `Add()`/`Sub()` set `dur.Op` before delegating so `String()` output stays truthful. Existing callers are unaffected.
2. Remove/unexport the field. Honest but a compile-time breaking change for any consumer who sets `Op` (harmlessly today); would warrant a version bump discussion.
3. Document-only: comment the field as informational. Weakest option since it leaves the trap in place.

If option 1 is chosen, note that the `Add`/`Sub` constants at `dur.go:16-18` are untyped iota ints; consider giving them a named type (e.g. `type Op int`) for the `DurWithOp` signature, but weigh that against breaking `Dur{Op: 1}` literals.

### Regression test

Whichever option is chosen: a test that constructs a `Dur` with `Op` set and asserts the chosen contract (e.g. `Compute()` subtracts when `Op == Sub`, and `Add()` still adds regardless).

## Verification for both fixes

1. `go test ./...` passes (74 existing tests plus the new ones) — run under `TZ=UTC` and `TZ=America/New_York` at minimum.
2. `go vet ./...` clean.
3. Rebuild the CLI (`go build ./cmd/dtmate`) and spot-check `dtmate conv -- -90m h` still prints "-1 hour", since the CLI shares these code paths.
