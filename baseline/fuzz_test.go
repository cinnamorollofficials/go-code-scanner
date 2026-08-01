package baseline

import (
	"os"
	"path/filepath"
	"testing"
)

func FuzzLoadNeverPanics(f *testing.F) {
	f.Add([]byte(`{"version":1,"fingerprint_version":"2","generated_at":"1970-01-01T00:00:00Z","entries":[]}`))
	f.Add([]byte(`{"version":999}`))
	f.Add([]byte(`{"entries":`))
	root := f.TempDir()
	f.Fuzz(func(t *testing.T, data []byte) {
		path := filepath.Join(root, t.Name()+".json")
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Skip()
		}
		_, _ = Load(path)
		_ = os.Remove(path)
	})
}
