#!/bin/bash
# Clean test runner that filters out GORM hook warnings and log output

# Run tests and filter out:
# - GORM hook mismatch warnings
# - Standard log package output (timestamp + file path lines)
# - JSON-formatted INFO logs from services
# - ANSI color codes
# - Blank lines (multiple consecutive)

# If first argument is "go" and second is "test", shift them and use the rest as test arguments
if [ "$1" = "go" ] && [ "$2" = "test" ]; then
    shift 2
fi

go test "$@" 2>&1 | \
    tr -d '\r' | \
    grep -vE '^20[0-9]{2}/[0-9]{2}/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}' | \
    grep -vE 'don'\''t match Before(Create|Update|Delete)Interface' | \
    grep -vE '"level":"(info|debug|INFO|DEBUG)"' | \
    grep -vE '^\[[0-9;]+m' | \
    awk 'NF > 0 {print}'
