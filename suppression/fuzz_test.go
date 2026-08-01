package suppression

import (
	"os"
	"path/filepath"
	"testing"
)

func FuzzLoadNeverPanics(f *testing.F) {
	f.Add([]byte(`{"version":1,"suppressions":[]}`))
	f.Add([]byte(`{"version":1,"suppressions":[{"file":"a.go","reason":"x","expires":"2030-01-01"}]}`))
	f.Add([]byte(`{"suppressions":`))
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
