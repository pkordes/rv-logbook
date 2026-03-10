#!/usr/bin/env python3
"""Format Go and TypeScript test names as a human-readable specification.

Modes are selected by the --lang flag (default: go):

  go (default):
    Scans Go *_test.go files for TestXxx functions.
    Usage: python scripts/spec-format.py backend/internal/repo backend/internal/apitest

  ts:
    Scans TypeScript *.test.ts(x) files for describe/it blocks.
    Usage: python scripts/spec-format.py --lang ts frontend/src

  e2e:
    Scans Playwright *.spec.ts files for test.describe/test blocks.
    Usage: python scripts/spec-format.py --lang e2e frontend/e2e
"""
import sys
import re
import os
import argparse


def camel_to_words(s: str) -> str:
    s = re.sub(r'([a-z0-9])([A-Z])', r'\1 \2', s)
    s = re.sub(r'([A-Z]+)([A-Z][a-z])', r'\1 \2', s)
    return s.lower()


def format_go_name(raw: str) -> str:
    parts = raw.split('_')
    words = [camel_to_words(p) for p in parts if p]
    return ' '.join(words)


def _extract_string(line: str) -> str:
    """Extract the first single- or double-quoted string from a line."""
    m = re.search(r'[\'"](.+?)[\'"]', line)
    return m.group(1) if m else ''


def scan_go(dirs):
    for dirpath in dirs:
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
                            tests.append(format_go_name(m.group(1)[4:]))
        if tests:
            print(f'\n{pkg}:')
            for t in tests:
                print(f' \u2714 {t}')


def scan_ts(dirs):
    for dirpath in dirs:
        for root, _dirs, files in os.walk(dirpath):
            for fname in sorted(files):
                if not re.search(r'\.test\.(ts|tsx)$', fname):
                    continue
                suite = None
                tests = []
                with open(os.path.join(root, fname), encoding='utf-8') as f:
                    for line in f:
                        if re.match(r'^describe\(', line):
                            if suite and tests:
                                print(f'\n{suite}:')
                                for t in tests:
                                    print(f' \u2714 {t}')
                            suite = _extract_string(line)
                            tests = []
                        elif re.match(r'^  (?:it|test)\(', line):
                            name = _extract_string(line)
                            if name:
                                tests.append(name)
                if suite and tests:
                    print(f'\n{suite}:')
                    for t in tests:
                        print(f' \u2714 {t}')


def scan_e2e(dirs):
    for dirpath in dirs:
        for root, _dirs, files in os.walk(dirpath):
            for fname in sorted(files):
                if not fname.endswith('.spec.ts'):
                    continue
                suite = None
                tests = []
                with open(os.path.join(root, fname), encoding='utf-8') as f:
                    for line in f:
                        if re.match(r'^test\.describe\(', line):
                            if suite and tests:
                                print(f'\n{suite}:')
                                for t in tests:
                                    print(f' \u2714 {t}')
                            suite = _extract_string(line)
                            tests = []
                        elif re.match(r'^  test\(', line):
                            name = _extract_string(line)
                            if name:
                                tests.append(name)
                if suite and tests:
                    print(f'\n{suite}:')
                    for t in tests:
                        print(f' \u2714 {t}')


parser = argparse.ArgumentParser(add_help=False)
parser.add_argument('--lang', default='go')
parser.add_argument('dirs', nargs='*')
args = parser.parse_args()

if args.lang == 'ts':
    scan_ts(args.dirs)
elif args.lang == 'e2e':
    scan_e2e(args.dirs)
else:
    scan_go(args.dirs)
