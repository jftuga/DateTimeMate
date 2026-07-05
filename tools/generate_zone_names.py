#!/usr/bin/env python3
"""Generate zone_names.go from the system IANA time zone database.

Walks /usr/share/zoneinfo (or a directory given with --zoneinfo), collects
the IANA zone names, and emits the DateTimeMate zone_names.go source file.
Only the Python standard library is used. The emitted list is validated at
runtime by ListIANAZones, so stale names are filtered there, not here.
"""

import argparse
import os
import sys

DEFAULT_ZONEINFO = "/usr/share/zoneinfo"

PREAMBLE = """package DateTimeMate

//go:generate python3 tools/generate_zone_names.py -o zone_names.go

// ianaZoneNames lists the IANA time zone names known at generation time;
// ListIANAZones validates each against the time zone database in use, so
// stale entries are filtered out at runtime. Regenerate on macOS/Linux
// (e.g. after a Go toolchain bump updates the embedded tzdata) with:
// go generate ./...
var ianaZoneNames = []string{
"""


def collect_zone_names(zoneinfo_dir: str) -> list[str]:
    """Collect IANA zone names from a zoneinfo directory tree.

    A zone name is the file path relative to the zoneinfo directory. Names
    must start with an ASCII uppercase letter, which excludes metadata files
    such as leapseconds, posixrules, tzdata.zi, and the .tab files. The
    placeholder zone "Factory" and any .tab file are excluded explicitly.

    Args:
        zoneinfo_dir: Path to the zoneinfo directory, e.g. /usr/share/zoneinfo.

    Returns:
        Zone names sorted in byte order, matching LC_ALL=C sort.

    Raises:
        FileNotFoundError: If zoneinfo_dir does not exist.
    """
    if not os.path.isdir(zoneinfo_dir):
        raise FileNotFoundError(f"zoneinfo directory not found: {zoneinfo_dir}")
    names: list[str] = []
    for root, _, files in os.walk(zoneinfo_dir, followlinks=True):
        for filename in files:
            path = os.path.join(root, filename)
            name = os.path.relpath(path, zoneinfo_dir)
            if not name[0].isascii() or not name[0].isupper():
                continue
            if name.endswith(".tab") or name == "Factory":
                continue
            names.append(name)
    return sorted(names)


def render_go_source(names: list[str]) -> str:
    """Render the zone_names.go source file contents.

    Args:
        names: Sorted IANA zone names to embed.

    Returns:
        The complete Go source file as a string.
    """
    lines = [PREAMBLE]
    for name in names:
        lines.append(f'\t"{name}",\n')
    lines.append("}\n")
    return "".join(lines)


def main() -> int:
    """Parse arguments, generate the Go source, and write it out.

    Returns:
        Process exit code: 0 on success, 1 on error.
    """
    parser = argparse.ArgumentParser(description="Generate zone_names.go from the IANA time zone database.")
    parser.add_argument("-z", "--zoneinfo", default=DEFAULT_ZONEINFO, help=f"zoneinfo directory (default: {DEFAULT_ZONEINFO})")
    parser.add_argument("-o", "--output", default="-", help="output file (default: stdout)")
    args = parser.parse_args()

    try:
        names = collect_zone_names(args.zoneinfo)
    except FileNotFoundError as err:
        print(err, file=sys.stderr)
        return 1
    if not names:
        print(f"no zone names found under {args.zoneinfo}", file=sys.stderr)
        return 1

    source = render_go_source(names)
    if args.output == "-":
        sys.stdout.write(source)
    else:
        with open(args.output, "w", encoding="utf-8") as f:
            f.write(source)
        print(f"wrote {len(names)} zone names to {args.output}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())
