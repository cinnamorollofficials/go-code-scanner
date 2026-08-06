package scripts

import (
	"os/exec"
	"testing"
)

func requireSh(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("skipping shell script test: sh executable not found on PATH")
	}
}
