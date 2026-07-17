# Dependency Replacement Plan

This document is a complete, self-contained implementation plan for removing
three of the four third-party date/time dependencies from this repository
(durafmt, carbon, parsetime) and replacing them with code under `internal/`.
The fourth, `lestrrat-go/strftime`, is deliberately kept as an external
dependency (see "strftime stays" below). It is written so that a fresh
session with no other context can execute the work. All investigation of the
dependency sources and this repo's usage has already been done; the findings
are recorded here.

Line numbers refer to the tree as of commit `a19cd87` (branch `remove-deps`).
If they have drifted, search for the named function instead.

## Status

Update this table as PRs land (add the merge commit or PR number to the
Notes column). A quick cross-check of the current state: `grep` the four
module paths in `go.mod` — a module still present means its PR has not
landed.

| Step | State | Notes |
|---|---|---|
| PR 1: durafmt -> internal/humandur | merged | PR #49, commit 425cc95 |
| PR 2: carbon arithmetic -> internal/datecalc | merged | PR #50, commit 18eda1b |
| PR 3: unified fallback parser (internal/dtparse) | implemented | on branch `remove-deps-pr3`, awaiting commit/PR |
| strftime | no work planned | kept as external dependency, see locked decision 3 |

## Locked decisions

These were decided with the maintainer and are not open questions:

1. **parsetime and carbon.Parse are replaced together** by a single unified
   internal fallback parser with one ordered layout list. They are not
   replaced separately.
2. **Compatibility bar for parsing: curated layouts, tightening OK.** The
   new parser supports an explicit, documented format list derived from the
   README, the test suite, and the reachable format families listed below.
   Obscure fuzzy inputs that only parsetime's permissive regexes accepted
   may stop parsing; document them as unsupported rather than chasing
   bug-for-bug compatibility.
3. **strftime stays in `go.mod` as an external dependency.** It is not
   ported and not rewritten. Rationale: it is the only healthy dependency of
   the four (pure functions, feature-complete, zero bugs caused in this
   repo), the CLI documents all ~38 of its specifiers as public surface, and
   the failure mode of a port or rewrite is silently wrong formatted output.
   Removing it would trade a frozen, battle-tested dependency for permanent
   ownership of week-numbering code (`%U`/`%V`/`%W`) with no concrete
   payoff. Its archived-but-stable transitive dep `pkg/errors` is accepted.
   Do not "improve" this decision by porting it later without the
   maintainer asking.
4. **Delivery is 3 sequential PRs**, each independently shippable and
   revertable:
   - PR 1: replace durafmt
   - PR 2: replace carbon's arithmetic and simple wrappers
   - PR 3: unified fallback parser (removes carbon, parsetime, go-timezone)

## Repository orientation

- Module: `github.com/jftuga/DateTimeMate`. Root package name is
  `DateTimeMate` (a library); the CLI is `cmd/dtmate` using cobra.
- Go version in `go.mod`: 1.23.3.
- The library's exported API (`Reformat`, `NewDiff`, `NewDur`, `NewConv`,
  `NewDurMath`, `NewTimeZoneConverter`, `LoadZoneDefinitions`,
  `ConvertRelativeDateToActual`, option funcs, etc.) **must not change**.
  The README "Library Usage" section documents it for consumers.
- Already dependency-free (pure stdlib): `conv.go`, `durmath.go`,
  `timezone.go`, `zone_definitions.go`, `zone_names.go`. The three target
  dependencies (plus strftime, which stays) are confined to
  `DateTimeMate.go`, `dur.go`, `diff.go`, and one helper call in
  `dur_test.go`.
- Tests: ~3,000 lines across `*_test.go` in the root package plus CLI tests
  under `cmd/dtmate/cmd/`. The bug-hunt suites (`parsing_fix_test.go`,
  `timezone_fix_test.go`) pin exact output strings and exact error behavior.
  They are the safety net for this entire effort.
- Verification for every PR:
  `go build ./... && go vet ./... && go test ./...`
  then `go mod tidy` and confirm the expected modules left `go.mod`/`go.sum`.
- Version: `ModVersion` in `DateTimeMate.go:26` (currently `1.16.0`).
  **The completed effort ships as `1.17.0`** — set `ModVersion` to `1.17.0`
  in the final PR (PR 3). If PRs 1 or 2 are released individually
  beforehand, they follow the repo's existing minor-bump-per-PR convention
  and PR 3 takes the next minor version instead. This is deliberately NOT a
  v2: the exported API does not change, the parse tightening in PR 3 is the
  same category of change the v1.13-v1.16 releases shipped as minor bumps,
  and a v2 would force the `/v2` module path change on every consumer for
  a change they cannot otherwise observe.
- README maintenance per PR: the "Acknowledgements > Imported Modules" list
  (README.md, near line 599) must be updated as modules are removed. Where
  code was derived from a removed module (carbon's layout list, the strftime
  port), keep an attribution line there and a copyright header in the ported
  file. PR 3 must also update the "Date and Duration Parsing Notes" section
  (README.md, near line 239) with the supported-format statement.
- New code goes under `internal/`. Proposed packages (names are suggestions,
  keep them if nothing better appears):
  - `internal/humandur` — durafmt replacement (PR 1)
  - `internal/datecalc` — carbon arithmetic replacement (PR 2)
  - `internal/dtparse` — unified fallback layout table + parse loop (PR 3)
  None of these may import the root package (import cycle); they must be
  stdlib-only. Where root-package data is needed (the zero-offset zone name
  set in PR 3), pass it in as a parameter.

## Current dependency usage inventory

### hako/durafmt (v0.0.0-20210608085754, unmaintained 2021 pseudo-version)

Exactly one path, in `diff.go:81-82`:

```go
parsed := durafmt.Parse(duration)
difference := fmt.Sprintf("%v", parsed)
```

This resolves to `Durafmt.String()`, which renders a `time.Duration` as
`"1 year 2 weeks 3 days 4 hours 5 minutes 6 seconds 7 milliseconds
8 microseconds"`.

### golang-module/carbon/v2 (v2.3.12)

Upstream has since migrated to `dromara/carbon` with breaking changes, so
this pin is a dead end regardless. Call sites:

| Call site | Carbon API | What it actually does (verified in carbon source) |
|---|---|---|
| `DateTimeMate.go:69,71` | `carbon.Now().String()` | `time.Now()` formatted `"2006-01-02 15:04:05"` in local zone |
| `DateTimeMate.go:76,78` | `carbon.Now().SubHours(24)/.AddHours(24).String()` | `time.Now().Add(∓24 * time.Hour)`, same formatting |
| `DateTimeMate.go:331` | `carbon.Parse(source)`, `.Error`, `.StdTime()` | loops `time.ParseInLocation` over ~80 hardcoded layouts (carbon `helper.go:52`), local zone |
| `dur.go:44-54` | `carbonFuncs` table: Add/Sub for 9 units | years/weeks/days -> `time.Time.AddDate`; hours..nanoseconds -> `time.Time.Add` |
| `dur.go:139-141` | `carbon.CreateFromStdTime(f)`, `.Error` | constructor wrapper; cannot error for a valid `time.Time` |
| `dur.go:187,192,196` | `.StdTime()` comparisons | unwrap to `time.Time` |
| `dur.go:212` | `.ToString()` | identical to `time.Time.String()` (`"2006-01-02 15:04:05.999999999 -0700 MST"`) |
| `dur.go:278,281` | function-table invocation with double type assertions | see PR 2 |
| `dur_test.go:474` | `carbon.Now().StartOfDay()` | `time.Date(y, m, d, 0, 0, 0, 0, time.Local)` for today |

### jftuga/parsetime (v0.4.0, fork of tkuchiki/parsetime)

One call site, `DateTimeMate.go:320-324`
(`parsetime.NewParseTime()` + `p.Parse(source)`), but it is a load-bearing
layer of the parse pipeline (see PR 3). It is a regex-based parser: four
regex families (ISO8601, RFC822/850/1123, ANSIC, US), all tried, best
priority wins. It drags in `tkuchiki/go-timezone` (transitive, in `go.mod`
as indirect) for zone-abbreviation lookups.

The repo already contains four defense layers against known parsetime
misbehavior; they are the reason this replacement has the highest payoff:

- `fixLocalZone` (`DateTimeMate.go:103`): parsetime stamps zone-less input
  with a fixed snapshot of *today's* zone, so a January date gets an EDT
  offset in July; this reinterprets the wall clock in `time.Local`.
- `parsedYearMismatch` (`DateTimeMate.go:344`): parsetime silently corrupts
  pre-1970 dates instead of erroring; this detects the corruption.
- `wallClockLayouts` (`DateTimeMate.go:141`) exists partly to intercept
  inputs *before* parsetime can corrupt them (see its comment).
- `sourceHasExplicitZone` (`DateTimeMate.go:123`) decides which parsetime
  results need the `fixLocalZone` repair. **Caution: this function is also
  called from `timezone.go:225` (`parseWallClockIn`) and must survive the
  parsetime removal** (its job — "does the input text itself name the zone
  found on the parsed time" — remains meaningful for any parser).

### lestrrat-go/strftime (v1.0.6) — KEPT, not replaced

The healthiest of the four (self-contained, stable, pure functions).
Production transitive dep: `pkg/errors` (archived but frozen-stable;
accepted). Call sites, both identical in shape:

- `DateTimeMate.go:391` (in `Reformat`):
  `strftime.New(outputFormat, strftime.WithUnixSeconds('s'))` then
  `f.FormatString(t)`.
- `dur.go:216` (in `Dur.renderResults`): same two calls.

The CLI publicly documents the full specifier table (~38 specifiers) in
`cmd/dtmate/cmd/fmt.go:46-...`, shown by `dtmate fmt -l`. All of it is API
surface, which is exactly why it stays external (see locked decision 3).
No work is planned for this dependency; it is inventoried here only so a
reader knows it was considered and kept deliberately.

---

## PR 1: Replace durafmt with `internal/humandur`

**Goal:** delete the `hako/durafmt` dependency; `diff.go` gets simpler, not
just different.

**Deliverable:** `internal/humandur/humandur.go` exporting one function,
e.g. `func Format(d time.Duration) string`, plus its own unit tests.

**Exact semantics to replicate** (verified against durafmt's source and
pinned by `diff_test.go`):

- Units, largest first: years, weeks, days, hours, minutes, seconds,
  milliseconds, microseconds. Flat conversion by integer division:
  1 year = 365 days exactly, 1 week = 7 days exactly. No calendar awareness.
- Zero-valued components are omitted.
- Singular vs. plural unit names (`1 microsecond` vs. `2 microseconds`).
- Components joined by single spaces, no Oxford anything:
  `"4 weeks 3 days 1 hour 2 minutes 3 seconds"` (see `diff_test.go:99`).
- Negative durations: leading `-` then the positive rendering.
- Zero duration renders as `"0 seconds"` (pinned by `diff_test.go:63` and
  `:215`).
- durafmt truncates below the microsecond. **Do not replicate that.**
  Instead make nanoseconds a ninth unit, and then delete the workaround
  block in `diff.go:83-107` that hand-appends the sub-microsecond remainder
  (`"1 microsecond 500 nanoseconds"`, lone-`ns` handling, the `""`/`"-"`
  patch-up). The pinned outputs in `diff_test.go` (around line 53) define
  the expected combined behavior — they must pass unchanged.

**Changes:**

- Add `internal/humandur` with tests covering: zero, exactly one unit, all
  units at once, singular forms, negative, sub-microsecond values,
  `duration % unit` boundaries.
- `diff.go`: replace the durafmt call and delete the remainder workaround;
  `CalculateDiff`'s formatted output must be byte-identical for every case
  in `diff_test.go` (including brief mode, which goes through
  `shrinkPeriod`).
- `go.mod`: `go mod tidy` removes `hako/durafmt`.
- README: remove durafmt from Acknowledgements.

**Risk:** minimal; every relevant output is pinned by existing tests.

---

## PR 2: Replace carbon arithmetic and trivial wrappers with `internal/datecalc`

**Goal:** remove every carbon usage *except* `carbon.Parse` (which PR 3
removes). After this PR, carbon remains in `go.mod` solely for the one
`carbon.Parse` call at `DateTimeMate.go:331`.

**Deliverable:** `internal/datecalc` with a single entry point, e.g.:

```go
// Apply adds (sign=+1) or subtracts (sign=-1) n units to t.
// unit is one of: year, week, day, hour, minute, second,
// millisecond, microsecond, nanosecond (singular, lowercase —
// exactly the keys of the current carbonFuncs map in dur.go).
func Apply(t time.Time, unit string, n int, sign int) (time.Time, error)
```

**Exact semantics to replicate** (this split is what carbon does and what
the DST-related tests in `dur_test.go` depend on):

- `year` -> `t.AddDate(sign*n, 0, 0)`
- `week` -> `t.AddDate(0, 0, sign*n*7)`
- `day` -> `t.AddDate(0, 0, sign*n)`
- `hour`..`nanosecond` -> `t.Add(time.Duration(sign*n) * unitDuration)`

`AddDate` is calendar-aware and overflow-normalizing (Feb 29 + 1 year =
Mar 1); `Add` is absolute. Do not "fix" either behavior.

**Changes in `dur.go`:**

- Delete the `carbonFuncs` `map[string]interface{}` table (lines 44-54) and
  rewrite `applyPeriod` (line 267) to call `datecalc.Apply`, eliminating the
  double type assertions at lines 278 and 281. The fractional-amount logic
  (integer part via the unit, fractional part via nanoseconds, using
  `unitNanoseconds`) stays exactly as is.
- `addOrSub` (line 124): work directly on `time.Time`; delete
  `carbon.CreateFromStdTime` and the dead `from.Error` check; the
  `StdTime()` unwrapping at lines 187, 192, 196 disappears naturally.
- `renderResults` (line 208): default output becomes `t.String()`
  (identical to carbon's `ToString()`).

**Changes in `DateTimeMate.go`:**

- `ConvertRelativeDateToActual` (line 66): replace
  `carbon.Now().String()` with
  `time.Now().Format("2006-01-02 15:04:05")` and the yesterday/tomorrow
  variants with `time.Now().Add(∓24 * time.Hour).Format(...)`. Keep the
  "exactly 24 hours across DST" comment and behavior. (The returned string
  is re-parsed by `wallClockLayouts` in local time, same as today.)

**Changes in `dur_test.go`:**

- Line 474: replace `carbon.Now().StartOfDay()` with
  `now := time.Now(); time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)`.

**Risk:** low; every call maps 1:1 to a stdlib call. The full `dur` test
suite (497 lines) plus `durmath`/`conv` suites must pass unchanged.

---

## PR 3: Unified fallback parser (`internal/dtparse`) — removes carbon, parsetime, go-timezone

This is the largest PR and the reason this effort exists.

### The current parse pipeline

`parseDateTime` (`DateTimeMate.go:284`) runs five layers in order. Callers
reach it via `parseDateTimeOrUnix` (`DateTimeMate.go:487`), which first
handles empty input, pure integers (unix timestamps by digit count: 10 =
seconds, 13 = ms; 4/8/14 digits = year/date/datetime; others rejected), and
relative words via `ConvertRelativeDateToActual`.

1. `wallClockLayouts` (line 141): zone-less layouts in `time.Local` —
   `"2006-01-02 15:04:05"`, `"2006-01-02T15:04:05"`, `"2006-01-02 15:04"`,
   `"2006-01-02"`, `"20060102150405"`. Any "out of range" parse error
   rejects immediately (`outOfRangeParseError`, line 274) so invalid
   components like `2024-02-30` can never fall through to a lenient layer.
2. `zonedLayouts` (line 145): layouts carrying zone/offset —
   RFC3339Nano, `"2006-01-02 15:04:05 -0700 MST"`,
   `"2006-01-02 15:04:05 -0700"`, `"2006-01-02 15:04:05 MST"`, UnixDate,
   RFC1123Z, RFC1123, RubyDate. A successful parse is then validated: Go's
   `time.Parse` fabricates a zero-offset zone for any abbreviation it cannot
   resolve, so a zero offset whose name is not in `zeroOffsetZoneNames`
   (line 153, built from `LoadZoneDefinitions()`) is either repaired via
   `parseOffsetSuffix` (defined in `timezone.go`; handles `+08`-style
   tokens) or **rejected with an error** ("zone abbreviation %q is not
   resolvable in this input position"). This validation must survive.
3. `parseSlashDate` (line 222): slash dates (`01/02/2024`, `1/2/24`,
   optional time suffixes). It *claims* anything shaped like
   `d{1,2}/d{1,2}/d{2 or 4}` — when claimed, no later layer may run, so
   `DTMATE_DATE_ORDER` (MDY/DMY) semantics can never be bypassed.
4. parsetime (line 320), with `fixLocalZone` repair for zone-less results
   and `parsedYearMismatch` corruption detection.
5. `carbon.Parse` (line 331) as the final fallback.

### The change

Layers 4 and 5 are replaced by **one** new layer backed by
`internal/dtparse`: a single ordered layout list tried with
`time.ParseInLocation`, exactly the mechanism layers 1-3 already use. This
collapses the two fallbacks whose disagreements caused past bugs
(`DateTimeMate_test.go:130`: "they used to try carbon and parsetime in
opposite orders") and makes ordering bugs structurally impossible.

### Design of `internal/dtparse`

Stdlib-only package. Core data structure: an ordered table of layout
entries with metadata, because three post-parse behaviors depend on the
layout's shape:

```go
type layoutKind int

const (
    wallClock layoutKind = iota // zone-less date+time or date: parse in a supplied location
    timeOnly                    // zone-less time of day: parse, then stamp with today's date
    zoned                       // carries zone or offset: parse with time.Parse, then caller validates the zone
)

type layoutEntry struct {
    layout string
    kind   layoutKind
}
```

Suggested API (adjust freely, but keep the package root-import-free):

```go
// Parse tries the ordered layout table. loc is the location for
// zone-less layouts (the root package passes time.Local).
// For kind==zoned results, zone validation is the CALLER's job.
func Parse(source string, loc *time.Location) (t time.Time, kind layoutKind, err error)
```

Rules the implementation must follow:

- **Ordering:** all entries live in one slice, tried top to bottom, first
  success wins. Order within the table is part of the spec — commit it with
  a comment explaining any entry whose position matters.
- **Out-of-range rejection:** the same `outOfRangeParseError` treatment as
  layers 1-3 — if any layout matches the shape but a component is out of
  range, return an error immediately instead of trying further layouts.
  (Either export a small helper from dtparse or keep using the root one on
  dtparse's returned error; the current root implementation checks
  `strings.Contains(err.Error(), "out of range")`.)
- **timeOnly stamping:** parsetime today gives bare times (`"08:30"`,
  `"11:00AM"`, `"12:34:56.1234"`) *today's* date in the local zone; Go's
  `time.Parse` would give year 0. After a timeOnly layout succeeds,
  transplant today's y/m/d in `loc`:
  `now := time.Now().In(loc); time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)`.
  This behavior is exercised by `diff_test.go:75-93`
  (`"12:00:00"`/`"15:30:45"`, `"11:00AM"`/`"11:00PM"`), `dur_test.go:297`,
  the CLI stdin tests in `cmd/dtmate/cmd` (`"15:16:15"` etc.), and the tz
  wall-clock path (`"08:30 CET"`, see `parsing_fix_test.go:114-138` and
  `timezone.go:parseSourceTime`).
- **Fractional seconds:** Go's `time.Parse` accepts a decimal fraction
  after seconds even when the layout has none, so `"12:34:56.1234"` needs
  no dedicated layout.
- **am/pm case:** Go's `"PM"` layout element accepts both `"PM"` and
  `"pm"`, so one layout per shape suffices.

### Integration in `DateTimeMate.go`

`parseDateTime` layers 1-3 stay untouched. Layers 4+5 become:

```go
t, kind, err := dtparse.Parse(source, time.Local)
// on success:
//   kind == zoned  -> run the SAME zone validation as layer 2
//                     (zeroOffsetZoneNames / parseOffsetSuffix / reject),
//                     extracted into a shared helper so layers 2 and 4
//                     cannot drift apart
//   otherwise      -> return t as is
// on failure       -> return the error
```

Extract layer 2's post-parse zone validation (currently inline at
`DateTimeMate.go:299-315`) into a helper (e.g.
`validateParsedZone(source string, t time.Time) (time.Time, error)`) used by
both layers. Note it needs `parseOffsetSuffix` from `timezone.go` — both
live in the root package, so no cycle.

Import-cycle note: `zeroOffsetZoneNames` is built from
`LoadZoneDefinitions()` (root). Since zone validation stays in the root
package, `internal/dtparse` needs no zone data at all. Keep it that way.

### The layout table

Two sources feed the table; dedupe against layers 1-3 and order carefully.

**Source A — formats that today only parsetime handles** (each item lists
the evidence that it must keep working):

| Format family | Example input | Evidence | Suggested layout(s) |
|---|---|---|---|
| Month-name date with comma | `Jan 2, 2024` | `DateTimeMate_test.go:131`, CLI tests | `Jan 2, 2006` (+ time suffix variants) |
| Full month name | `January 2, 2024 08:30:00` | README examples | `January 2, 2006 15:04:05`, `January 2, 2006` |
| Day-first month name | `2-Jan-2024`, `22-Jul-2024 08:21:44` | strftime `%v` output round-trips | `2-Jan-2006`, `2-Jan-2006 15:04:05` |
| ANSIC with zone + trailing year | `Jan 15 12:00:00 EDT 2026` | `TestExplicitZonePreservedAcrossDST` (`DateTimeMate_test.go:151-180`), deliberately "not covered by zonedLayouts" | `Jan 2 15:04:05 MST 2006` (kind=zoned) |
| ANSIC without zone | `Jan 2 15:04:05 2024` | parsetime ANSIC family | `Jan 2 15:04:05 2006` |
| Bare time 24h | `08:30`, `15:16:15` | diff/dur/CLI/tz tests above | `15:04:05`, `15:04` (kind=timeOnly) |
| Bare time am/pm | `11:00AM`, `3:04pm` | `diff_test.go:85`, `dur_test.go:297` | `3:04PM`, `3:04:05PM` (kind=timeOnly) |
| Year-month | `2024-01` | `DateTimeMate_test.go:131` | `2006-1` |
| Dotted date | `2024.01.15` | parsetime ISO8601 separator class | `2006.1.2` + time variants (carbon also had these) |

**Source B — carbon's layout list** (carbon `helper.go:52`, MIT; keep an
attribution comment). Reproduced here with the named constants resolved, so
the module cache is not needed. Before adopting each entry, apply the
exclusion rules below the list.

```
Mon, Jan 2, 2006 3:04 PM
2006-01-02 15:04:05            (dup: wallClockLayouts)
2006-01-02 15:04:05.999999999
20060102150405                 (dup: wallClockLayouts)
20060102150405.999999999
2006-01-02                     (dup: wallClockLayouts)
2006-01-02.999999999
20060102                       (handled by parseIntegerDateTime)
20060102.999999999
2006-01-02T15:04:05-07:00
2006-01-02T15:04:05.999999999-07:00
02 Jan 06 15:04 MST            (RFC822; the Go constant — no weekday, no seconds)
02 Jan 06 15:04 -0700          (RFC822Z; the Go constant)
Monday, 02-Jan-06 15:04:05 MST (RFC850)
Mon, 02 Jan 2006 15:04:05 MST  (RFC1123; dup: zonedLayouts)
Mon, 02 Jan 2006 15:04:05 -0700 (RFC1123Z; dup: zonedLayouts)
2006-01-02T15:04:05Z07:00      (RFC3339)
2006-01-02T15:04:05.999999999Z07:00 (RFC3339Nano; dup: zonedLayouts)
Mon, 02 Jan 06 15:04:05 -0700  (RFC1036)
Mon, 02 Jan 2006 15:04:05 MST  (RFC7231; same as RFC1123)
3:04PM                         (Kitchen; kind=timeOnly)
Monday, 02-Jan-2006 15:04:05 MST (Cookie)
Mon Jan _2 15:04:05 2006       (ANSIC)
Mon Jan _2 15:04:05 MST 2006   (UnixDate; dup: zonedLayouts)
Mon Jan 02 15:04:05 -0700 2006 (RubyDate; dup: zonedLayouts)
2006                           (handled by parseIntegerDateTime)
2006-1 / 2006-1-2 / 2006-1-2 15 / 2006-1-2 15:4 / 2006-1-2 15:4:5 / 2006-1-2 15:4:5.999999999
2006.1 / 2006.1.2 / 2006.1.2 15 / 2006.1.2 15:4 / 2006.1.2 15:4:5 / 2006.1.2 15:4:5.999999999
2006/1 / 2006/1/2 / 2006/1/2 15 / 2006/1/2 15:4 / 2006/1/2 15:4:5 / 2006/1/2 15:4:5.999999999
2006-01-02 15:04:05 -0700 MST  (dup: zonedLayouts)
2006-01-02 15:04:05PM MST / 2006-01-02 15:04:05.999999999PM MST
2006-1-2 15:4:5PM MST / 2006-1-2 15:4:5.999999999PM MST
2006-01-02 15:04:05 PM MST / 2006-01-02 15:04:05.999999999 PM MST
2006-1-2 15:4:5 PM MST / 2006-1-2 15:4:5.999999999 PM MST
1/2/2006 / 1/2/2006 15 / 1/2/2006 15:4 / 1/2/2006 15:4:5 / 1/2/2006 15:4:5.999999999   (EXCLUDE, see below)
2006-1-2 15:4:5 -0700 MST / 2006-1-2 15:4:5.999999999 -0700 MST
2006-1-2 15:04:05 -0700 MST / 2006-1-2 15:04:05.999999999 -0700 MST
2006-01-02T15:04:05            (dup: wallClockLayouts)
2006-01-02T15:04:05.999999999
2006-1-2T3:4:5 / 2006-1-2T3:4:5.999999999
2006-01-02T15:04:05Z07 / 2006-01-02T15:04:05.999999999Z07
2006-1-2T15:4:5Z07 / 2006-1-2T15:4:5.999999999Z07
2006-01-02T15:04:05Z07:00 / 2006-01-02T15:04:05.999999999Z07:00
2006-1-2T15:4:5Z07:00 / 2006-1-2T15:4:5.999999999Z07:00
2006-01-02T15:04:05-07:00 / 2006-01-02T15:04:05.999999999-07:00
2006-1-2T15:4:5-07:00 / 2006-1-2T15:4:5.999999999-07:00
2006-01-02T15:04:05-0700 / 2006-01-02T15:04:05.999999999-0700
2006-1-2T3:4:5-0700 / 2006-1-2T3:4:5.999999999-0700
20060102150405-07:00 / 20060102150405.999999999-07:00
20060102150405Z07 / 20060102150405.999999999Z07
20060102150405Z07:00 / 20060102150405.999999999Z07:00
```

**Exclusion rules:**

- Drop exact duplicates of `wallClockLayouts` and `zonedLayouts` entries
  (marked "dup" above) — they can never be reached.
- Drop `1/2/2006*`: `parseSlashDate` claims every `d{1,2}/d{1,2}/d{2,4}`
  shape before this layer, so these are unreachable — and if they were ever
  reachable they would bypass `DTMATE_DATE_ORDER`. Excluding them makes the
  invariant visible. (`2006/1/2` with a 4-digit first field is NOT claimed
  by the slash parser and stays.)
- Drop `2006` and `20060102`: pure integers never reach `parseDateTime`
  (`parseDateTimeOrUnix` routes them to `parseIntegerDateTime`). Fractional
  variants like `20060102.999` are also pure-integer-ish only when they have
  no fraction; with a fraction they are not integers — decide whether to
  keep them (harmless) and note the decision.
- Every layout containing `MST` or a numeric offset gets `kind = zoned` and
  therefore runs through the shared zone validation. Layouts ending in
  `Z07`/`Z07:00` also count as zoned (a literal trailing `Z` denotes UTC;
  the existing `zeroOffsetZoneNames` check already accepts it via the
  empty-name/`UTC` entries — verify against
  `TestZuluWithContradictoryZoneRejected` in `parsing_fix_test.go:92`).

**Zone abbreviation semantics (unchanged, important):** Go's `time.Parse`
resolves an abbreviation only if the *local* location defines it (e.g. EDT
when local is America/New_York), producing the correct instant even on a
date in the opposite DST regime — that exact behavior is what
`TestExplicitZonePreservedAcrossDST` requires. Unresolvable abbreviations
come back as fabricated zero-offset zones and must be rejected by the shared
validation helper, same as layer 2 today. Do NOT add go-timezone-style
abbreviation resolution to the parse path; abbreviation-to-offset resolution
via the in-repo zone table is the `tz` sub-command's job
(`timezone.go:parseSourceTime` splits a trailing zone token off *before* the
wall clock reaches `parseDateTime`).

**Two-digit years in month-name formats:** parsetime pivots 69-99 -> 19xx,
00-68 -> 20xx, which is identical to Go's `06` layout convention (documented
in README parsing notes). If any adopted layout uses `06`, Go handles it;
no custom code needed.

### Code and tests to delete or update

After the switch, in the root package:

- Delete `fixLocalZone` (`DateTimeMate.go:103`) — its only caller was the
  parsetime branch. The memory/comment about "parsetime stamps today's DST
  offset" becomes historical.
- Delete `parsedYearMismatch` (`DateTimeMate.go:344`) and its direct unit
  tests (`DateTimeMate_test.go:303-320` area, e.g. the
  `"Mon Feb 28 23:59:59 EST 1900"` and `"12:34:56.1234"` probes). The
  *behavioral* pre-1970 tests (`DateTimeMate_test.go:265` area,
  `dur_test.go:446` area) must keep passing — pre-1970 support now comes
  from layers 1-3 and dtparse, none of which have parsetime's 1970 floor
  (parsetime's year regex literally required `19[7-9][0-9]|2[0-9]{3}`).
- **Keep** `sourceHasExplicitZone` and `zoneOffsetRegexp` — still used by
  `timezone.go:225`. Re-check its doc comment (it currently explains itself
  in terms of parsetime).
- Delete the parsetime and carbon imports; drop the "refusing unreliable
  parse" error path (line 335) — dtparse errors are ordinary parse errors.
- Update the comment block above `parseDateTime` (lines 278-283) to describe
  the new 4-layer pipeline.
- `go mod tidy` must remove `golang-module/carbon/v2`, `jftuga/parsetime`,
  and `tkuchiki/go-timezone`.
- README: update Acknowledgements (carbon attribution stays as "layout list
  derived from"; parsetime entry removed) and extend "Date and Duration
  Parsing Notes" with a sentence stating the supported input formats are now
  a fixed, documented list.

### Test strategy (do this before removing the old code)

1. **Differential corpus test (temporary):** while parsetime and carbon are
   still importable, add a throwaway test that feeds a corpus through the
   old layer-4/5 pair and through `dtparse.Parse` and diffs the resulting
   instants. Corpus: every date/time literal in the existing test files,
   the README examples (lines 279-560), plus generated variants (each
   layout family x {with/without time, with/without zone, 2/4-digit year}).
   Investigate every mismatch: either add a layout, or record the input as
   deliberately unsupported in the PR description. Delete this test before
   merge (or keep the corpus as a fixture against dtparse alone).
2. **New permanent tests** in `internal/dtparse`: one table-driven test per
   format family in the Source A table, including the timeOnly stamping
   behavior and zoned-kind classification.
3. The full existing suite must pass unchanged, most critically:
   `TestParserAgreement`, `TestExplicitZonePreservedAcrossDST`,
   `TestPre1970*`, everything in `parsing_fix_test.go` and
   `timezone_fix_test.go`, `TestDiffAmPm`, `TestDiffTwoTimesSameDay`, and
   the CLI stdin tests under `cmd/dtmate/cmd/`.

**Risk:** moderate — this is the one PR where silent behavior change is
possible. The differential corpus test is the mitigation; do not skip it.

---

## Per-PR checklist

For each of the three PRs:

1. `go build ./... && go vet ./... && go test ./...` — all green.
2. `go mod tidy` — confirm exactly the expected modules disappeared:
   - PR 1: `hako/durafmt`
   - PR 2: none (carbon stays until PR 3)
   - PR 3: `golang-module/carbon/v2`, `jftuga/parsetime`,
     `tkuchiki/go-timezone`
   The end state of `go.mod`: `spf13/cobra`, `stretchr/testify`, and
   `lestrrat-go/strftime` (kept deliberately) as direct requirements, plus
   their transitive deps (`spf13/pflag`, `inconshreveable/mousetrap`,
   `davecgh/go-spew`, `pmezard/go-difflib`, `gopkg.in/yaml.v3`,
   `pkg/errors`).
3. Exercise the real CLI, not just tests: build `cmd/dtmate` and run the
   README "Command Line Examples" for the affected sub-commands
   (`diff` for PR 1; `dur`/`durmath` for PR 2; everything for PR 3),
   comparing against the outputs shown in the README.
4. Update README Acknowledgements (and for PR 3, the parsing notes).
   strftime stays listed in Acknowledgements as an imported module.
5. Version: PR 3 sets `ModVersion` to `1.17.0` (or the next minor version
   if PRs 1 or 2 were released individually first); PRs 1 and 2 only bump
   the minor version if released individually. No major version bump — see
   "Repository orientation".
6. MIT attribution header present in any derived file (PR 3 layout list).
