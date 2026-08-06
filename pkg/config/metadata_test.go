package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cinnamorollofficials/go-code-scanner/pkg/config"
)

func TestConfigurationMetadataGolden(t *testing.T) {
	metaJSON, err := config.MetadataJSON()
	if err != nil {
		t.Fatalf("failed to generate configuration metadata JSON: %v", err)
	}

	goldenPath := filepath.Join("testdata", "metadata_golden.json")
	if _, err := os.Stat(goldenPath); os.IsNotExist(err) {
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatalf("failed to create testdata dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, metaJSON, 0644); err != nil {
			t.Fatalf("failed to write golden metadata file: %v", err)
		}
	}

	goldenData, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden metadata file: %v", err)
	}

	if string(metaJSON) != string(goldenData) {
		t.Errorf("configuration metadata does not match golden file %s", goldenPath)
	}
}
