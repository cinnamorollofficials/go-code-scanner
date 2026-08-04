#!/bin/sh
set -eu

fuzz_time=${FUZZ_TIME:-1s}

go test -run '^$' -fuzz '^FuzzLoadNeverPanics$' -fuzztime "$fuzz_time" ./config
go test -run '^$' -fuzz '^FuzzLoadNeverPanics$' -fuzztime "$fuzz_time" ./rules
go test -run '^$' -fuzz '^FuzzLoadNeverPanics$' -fuzztime "$fuzz_time" ./suppression
go test -run '^$' -fuzz '^FuzzLoadNeverPanics$' -fuzztime "$fuzz_time" ./baseline
go test -run '^$' -fuzz '^FuzzExternalAdapterParsersNeverPanic$' -fuzztime "$fuzz_time" ./scanner/adapters
go test -run '^$' -fuzz '^FuzzFrontendParsersNeverPanic$' -fuzztime "$fuzz_time" ./scanner/frontend
