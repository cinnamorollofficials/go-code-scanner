#!/bin/sh
set -eu

fuzz_time=${FUZZ_TIME:-1s}

go test -run '^$' -fuzz '^FuzzLoadNeverPanics$' -fuzztime "$fuzz_time" ./pkg/config
go test -run '^$' -fuzz '^FuzzLoadNeverPanics$' -fuzztime "$fuzz_time" ./pkg/rules
go test -run '^$' -fuzz '^FuzzLoadNeverPanics$' -fuzztime "$fuzz_time" ./pkg/suppression
go test -run '^$' -fuzz '^FuzzLoadNeverPanics$' -fuzztime "$fuzz_time" ./pkg/baseline
go test -run '^$' -fuzz '^FuzzExternalAdapterParsersNeverPanic$' -fuzztime "$fuzz_time" ./pkg/scanner/adapters
go test -run '^$' -fuzz '^FuzzFrontendParsersNeverPanic$' -fuzztime "$fuzz_time" ./pkg/scanner/frontend
