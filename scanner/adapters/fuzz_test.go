package adapters

import "testing"

func FuzzExternalAdapterParsersNeverPanic(f *testing.F) {
	f.Add([]byte(`{}`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`{"results":`))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parseGosec(data)
		_, _ = parseGitleaks(data)
		_, _ = parseTrivy(data)
		_, _ = parseSemgrep(data)
		_, _ = parseGovulncheck(data)
		_, _ = parseOSVScanner(data)
	})
}
