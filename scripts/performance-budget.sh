#!/bin/sh
set -eu

temporary=$(mktemp "${TMPDIR:-/tmp}/go-code-scanner-benchmarks.XXXXXX")
cleanup() {
  rm -f -- "$temporary"
}
trap cleanup EXIT HUP INT TERM

go test -run '^$' -bench '^(BenchmarkDiscovery|BenchmarkPatternScanning|BenchmarkBaselineComparison|BenchmarkCacheHit|BenchmarkFastPreCommit)$' -benchtime=10x -count=1 . | tee "$temporary"

awk '
BEGIN {
  budget["BenchmarkDiscovery"] = 100000000
  budget["BenchmarkPatternScanning"] = 100000000
  budget["BenchmarkBaselineComparison"] = 50000000
  budget["BenchmarkCacheHit"] = 20000000
  budget["BenchmarkFastPreCommit"] = 1000000000
}
/^Benchmark/ {
  name = $1
  sub(/-[0-9]+$/, "", name)
  if (name in budget) {
    seen++
    ns = $(NF-1) + 0
    if (ns > budget[name]) {
      printf "%s exceeded budget: %.0f ns/op > %.0f ns/op\n", name, ns, budget[name] > "/dev/stderr"
      failed = 1
    }
  }
}
END {
  if (seen != 5) {
    printf "expected 5 performance benchmarks, found %d\n", seen > "/dev/stderr"
    exit 2
  }
  if (failed) exit 1
}
' "$temporary"
