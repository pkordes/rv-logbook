#!/usr/bin/env python3
"""
Check text files for UTF-8 BOM and cp1252 mojibake patterns.

Mojibake happens when UTF-8 bytes are decoded as Windows-1252 and then
re-saved as UTF-8.  The canonical symptom: em-dashes (U+2014) appear as
the three-character sequence U+00E2 U+20AC U+201D, and box-drawing
characters appear as sequences starting with U+00E2 U+2514, etc.

Usage:
    python scripts/check-encoding.py [root_dir]   # defaults to repo root

Exit code is 0 if all files are clean, 1 if any issues are found.
"""

import pathlib
import sys


# ---------------------------------------------------------------------------
# Files to check.
# ---------------------------------------------------------------------------
PATTERNS = ["**/*.md", "**/*.txt"]

# Directories that are never relevant.
EXCLUDE_DIRS = {".git", "node_modules", "dist", "build", "__pycache__"}

# ---------------------------------------------------------------------------
# Mojibake markers.
#
# Each entry is (marker_unicode_string, human_readable_explanation).
#
# These two-character sequences are produced when UTF-8 bytes are decoded as
# Windows-1252 and then stored as UTF-8.  They cannot appear legitimately in
# English technical prose.
#
# All strings use \u escapes so that this source file stays pure ASCII and
# avoids any bootstrapping irony.
#
# How the mapping works (example: em-dash U+2014):
#   UTF-8 bytes of em-dash: E2 80 94
#   Decoded as cp1252:  E2 -> U+00E2 (a-circumflex)
#                       80 -> U+20AC (euro sign)
#                       94 -> U+201D (right double quotation mark)
#   Those three chars re-encoded as UTF-8 produce what you see in the file.
#   The first two chars (U+00E2 U+20AC) are our detection marker.
#
# Box-drawing example (vertical bar U+2502):
#   UTF-8 bytes: E2 94 82
#   Decoded as cp1252: E2 -> U+00E2, 94 -> U+201D, 82 -> U+201A
#   Two-char prefix: U+00E2 U+201D
# ---------------------------------------------------------------------------
MOJIBAKE_MARKERS = [
    # em-dash (U+2014), en-dash (U+2013), curly quotes
    ("\u00e2\u20ac", "em/en-dash or curly quote decoded as cp1252"),
    # box-drawing chars: U+2500..U+257F (prefix U+00E2 U+201D from byte 0x94)
    ("\u00e2\u201d", "box-drawing character decoded as cp1252"),
    # greater-than-or-equal U+2265, less-than-or-equal U+2264
    ("\u00e2\u2030", "comparison operator (>= <=) decoded as cp1252"),
    # arrows U+2190..U+21FF
    ("\u00e2\u2020", "arrow character decoded as cp1252"),
    # block elements U+2580..U+259F
    ("\u00e2\u2013", "block element decoded as cp1252"),
]


def check_file(path: pathlib.Path) -> list:
    """Return a list of human-readable error strings for *path*."""
    errors = []

    try:
        raw = path.read_bytes()
    except OSError as exc:
        return ["Cannot read {}: {}".format(path, exc)]

    # BOM check.
    if raw.startswith(b"\xef\xbb\xbf"):
        errors.append("{}: contains UTF-8 BOM -- save as UTF-8 without BOM".format(path))
        raw = raw[3:]

    # Encoding check.
    try:
        text = raw.decode("utf-8")
    except UnicodeDecodeError as exc:
        errors.append("{}: not valid UTF-8: {}".format(path, exc))
        return errors

    # Mojibake check.
    for marker, explanation in MOJIBAKE_MARKERS:
        if marker not in text:
            continue
        for lineno, line in enumerate(text.splitlines(), start=1):
            if marker in line:
                errors.append(
                    "{}:{}: mojibake -- {} "
                    "(marker U+{:04X} U+{:04X})".format(
                        path,
                        lineno,
                        explanation,
                        ord(marker[0]),
                        ord(marker[1]),
                    )
                )
                break  # one report per marker per file

    return errors


def iter_files(root: pathlib.Path):
    """Yield files matching PATTERNS under *root*, skipping EXCLUDE_DIRS."""
    for pattern in PATTERNS:
        for candidate in root.glob(pattern):
            if any(part in EXCLUDE_DIRS for part in candidate.parts):
                continue
            yield candidate


def main():
    root = pathlib.Path(sys.argv[1]) if len(sys.argv) > 1 else pathlib.Path(".")
    root = root.resolve()

    files = sorted(set(iter_files(root)))
    if not files:
        print("check-encoding: no files found to check.")
        return 0

    all_errors = []
    for f in files:
        all_errors.extend(check_file(f))

    if all_errors:
        print("check-encoding: encoding issues found\n")
        for err in all_errors:
            print("  {}".format(err))
        print(
            "\n{} issue(s) across {} file(s) checked.".format(
                len(all_errors), len(files)
            )
        )
        print(
            "\nTo fix a mojibake file:\n"
            "  python -c \""
            "import pathlib; p=pathlib.Path('FILE'); "
            "t=p.read_bytes().lstrip(b'\\xef\\xbb\\xbf').decode('utf-8'); "
            "p.write_text(t.encode('cp1252').decode('utf-8'), encoding='utf-8')"
            "\""
        )
        return 1

    print("check-encoding: {} file(s) checked -- all OK.".format(len(files)))
    return 0


if __name__ == "__main__":
    sys.exit(main())
