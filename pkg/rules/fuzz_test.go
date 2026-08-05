package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func FuzzLoadNeverPanics(f *testing.F) {
	f.Add([]byte(`{"version":1,"rules":[]}`))
	f.Add([]byte(`{"version":1,"rules":[{"id":"custom","pattern":"secret","severity":"HIGH","category":"security","description":"test"}]}`))
	f.Add([]byte(`{"version":`))
	root := f.TempDir()
	f.Fuzz(func(t *testing.T, data []byte) {
		path := filepath.Join(root, t.Name()+".json")
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Skip()
		}
		_, _ = Load([]string{path})
		_ = os.Remove(path)
	})
}
