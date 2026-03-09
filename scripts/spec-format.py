#!/usr/bin/env python3
"""Format Go test names as a human-readable specification.

Two modes:

  Stdin mode (used by backend/spec):
    Reads JSON from `go test -json` piped through gotestdox, or plain test
    names from `go test -list`, and pretty-prints them.
    Usage: go test -json ./... | gotestdox | python scripts/spec-format.py

  Directory mode (used by backend/spec/integration):
    Scans Go test files in the given directories and prints all test names.
    No database or build tags required — reads source files directly.
    Usage: python scripts/spec-format.py backend/internal/repo backend/internal/apitest
"""
import sys
import re
import os


def camel_to_words(s: str) -> str:
    # Split before an uppercase letter that follows a lowercase letter or digit.
    # e.g. "TripRepo" -> "Trip Repo", "GetByID" -> "Get By ID"
    s = re.sub(r'([a-z0-9])([A-Z])', r'\1 \2', s)
    # Split before an uppercase+lowercase sequence preceded by uppercase run.
    # e.g. "IDFoo" -> "ID Foo"
    s = re.sub(r'([A-Z]+)([A-Z][a-z])', r'\1 \2', s)
    return s.lower()


def format_name(raw: str) -> str:
    """Convert TestFoo_Bar_Baz to 'foo bar baz'."""
    parts = raw.split('_')
    words = [camel_to_words(p) for p in parts if p]
    return ' '.join(words)


if len(sys.argv) > 1:
    # Directory mode: scan Go *_test.go files for test function declarations.
    for dirpath in sys.argv[1:]:
        pkg = os.path.basename(dirpath.rstrip('/\\'))
        tests = []
        for root, _dirs, files in os.walk(dirpath):
            for fname in sorted(files):
                if not fname.endswith('_test.go'):
                    continue
                with open(os.path.join(root, fname), encoding='utf-8') as f:
                    for line in f:
                        m = re.match(r'^func (Test[A-Z]\w*)\(', line)
                        if m and m.group(1) != 'TestMain':
                            tests.append(format_name(m.group(1)[4:]))
        if tests:
            print(f'\n{pkg}:')
            for t in tests:
                print(f' \u2714 {t}')
else:
    # Stdin mode: format plain test names from `go test -list` output.
    for line in sys.stdin:
        line = line.rstrip()
        if line.startswith('Test'):
            print(f' \u2714 {format_name(line[4:])}')
