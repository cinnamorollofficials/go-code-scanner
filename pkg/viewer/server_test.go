package viewer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestViewerServerEndpoints(t *testing.T) {
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "sample.go")
	if err := os.WriteFile(testFilePath, []byte("package main\n\nfunc main() {}"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	server := NewServer(ServerOptions{
		Root: tempDir,
	})
	handler := server.Handler()

	t.Run("GET /api/status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		var status StatusResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &status); err != nil {
			t.Fatalf("failed to parse status json: %v", err)
		}
		if status.Root == "" {
			t.Fatalf("expected root path in status, got empty")
		}
	})

	t.Run("POST /api/scan", func(t *testing.T) {
		body := bytes.NewBufferString(`{"mode":"full"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/scan", body)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("GET /api/file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/file?path=sample.go", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var fileResp struct {
			Path       string   `json:"path"`
			TotalLines int      `json:"total_lines"`
			Lines      []string `json:"lines"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &fileResp); err != nil {
			t.Fatalf("failed to unmarshal file response: %v", err)
		}
		if fileResp.TotalLines != 3 {
			t.Fatalf("expected 3 lines, got %d", fileResp.TotalLines)
		}
	})

	t.Run("GET / (Default UI)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		if !bytes.Contains(rec.Body.Bytes(), []byte("Go Code Scanner")) {
			t.Fatalf("expected dashboard HTML content in response")
		}
	})
}
