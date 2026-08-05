package config

import (
	"os"
	"path/filepath"
	"testing"
)

func FuzzLoadNeverPanics(f *testing.F) {
	f.Add([]byte(`{"version":1,"project":"fixture","root":".","output":"report.json","fail_on":"HIGH","workers":1,"pattern_max_file_bytes":1,"pattern_max_line_bytes":1}`))
	f.Add([]byte(`{"unknown":true}`))
	f.Add([]byte(`{"version":`))
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
