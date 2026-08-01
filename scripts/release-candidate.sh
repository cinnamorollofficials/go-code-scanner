#!/bin/sh
set -eu

report_path=".self-scan-report.json"
cleanup() {
  rm -f -- "$report_path"
}
trap cleanup EXIT HUP INT TERM

go test ./...
go test -race ./...
go vet ./...
git diff --check

go test . -run '^TestRuntimeCachePreservesResultsAndInvalidatesContent$' -count=1
go test ./scripts -run '^(TestReleaseBuildIsReproducible|TestReleasePipelineEndToEnd|TestReleaseBinaryHookLifecycle)$' -count=1
go test ./reporter ./compatibility ./release ./scripts -run 'Golden|Contract|Structure|Checksums' -count=1

./scripts/fuzz-smoke.sh
./scripts/vulnerability-scan.sh --if-available
./scripts/performance-budget.sh
go run ./cmd/security-review scan --root . --quiet --output "$report_path"

echo "release candidate verification completed"
