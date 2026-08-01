//go:build unix

package command

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/finding"
	"github.com/cinnamorollofficials/go-code-scanner/scanner"
)

func TestCancellationTerminatesCommandProcessGroup(t *testing.T) {
	pidFile := t.TempDir() + "/child.pid"
	source, err := New(Spec{
		ID: "process-group", Domain: finding.Reliability,
		Command:  []string{os.Args[0], "-test.run=TestCommandHelperProcess", "--", "spawn-child", pidFile},
		Severity: finding.High, Category: "process", Description: "fixture",
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	result := source.Scan(ctx, scanner.Request{Root: t.TempDir(), Mode: "full"})
	if result.State != finding.ScannerFailed || result.Failure != scanner.FailureCanceled {
		t.Fatalf("unexpected canceled result: %+v", result)
	}
	data, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		err = syscall.Kill(pid, 0)
		if errors.Is(err, syscall.ESRCH) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("child process %d survived scanner cancellation", pid)
}
