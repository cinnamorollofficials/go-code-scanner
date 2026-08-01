#!/bin/sh
set -eu

report_path=".self-scan-report.json"
cleanup() {
  rm -f -- "$report_path"
}
trap cleanup EXIT INT TERM

go test ./...
go test -race ./...
go vet ./...
go test -run '^$' -bench . ./cache
go run ./cmd/security-review scan --root . --quiet --output "$report_path"

echo "verification and self-scan completed"
