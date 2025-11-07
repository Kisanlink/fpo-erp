@echo off
REM Clean test runner that filters out GORM hook warnings for Windows

REM Run tests and use PowerShell to filter warnings
go test %* 2>&1 | powershell -Command "$input | Where-Object { $_ -notmatch \"don't match Before.*Interface\" -and $_ -notmatch \"\\[3[0-9];1m\" -and $_ -notmatch \"\\[0m\\[35m\" }"
