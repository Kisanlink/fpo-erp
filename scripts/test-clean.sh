#!/bin/bash
# Clean test runner that filters out GORM hook warnings

# Run tests and filter out GORM hook mismatch warnings
# Filters out:
# - Warning text lines
# - Timestamp + file path lines
# - Timestamp + color code only lines (orphaned)
go test "$@" 2>&1 | sed -r '/don'\''t match Before(Create|Update|Delete)Interface/d; /^20[0-9]{2}\/[0-9]{2}\/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} \x1b\[3[0-9];1m.*database\.go:[0-9]+$/d; /^20[0-9]{2}\/[0-9]{2}\/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} \x1b\[3[0-9];1m$/d; /^\x1b\[0m\x1b\[35m/d'
