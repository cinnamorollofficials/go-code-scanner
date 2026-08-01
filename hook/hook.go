package hook

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cinnamorollofficials/go-code-scanner/gitrepo"
)

const (
	PreCommit = "pre-commit"
	marker    = "# Managed by go-code-scanner; DO NOT EDIT."
)

type State string

const (
	Missing   State = "missing"
	Installed State = "installed"
	Conflict  State = "conflict"
)

type Manager struct {
	repository *gitrepo.Repository
	binary     string
}

func NewManager(repository *gitrepo.Repository, binary string) (*Manager, error) {
	if repository == nil {
		return nil, fmt.Errorf("repository is required")
	}
	if strings.TrimSpace(binary) == "" {
		return nil, fmt.Errorf("binary path is required")
	}
	absBinary, err := filepath.Abs(binary)
	if err != nil {
		return nil, fmt.Errorf("resolve binary path: %w", err)
	}
	return &Manager{repository: repository, binary: filepath.Clean(absBinary)}, nil
}

func (m *Manager) Install(ctx context.Context, event string) error {
	target, content, err := m.target(ctx, event)
	if err != nil {
		return err
	}
	state, err := inspect(target, content)
	if err != nil {
		return err
	}
	switch state {
	case Installed:
		return os.Chmod(target, 0o755)
	case Conflict:
		return fmt.Errorf("refusing to replace existing hook %s", target)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create hooks directory: %w", err)
	}
	file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o755)
	if err != nil {
		return fmt.Errorf("create hook %s: %w", target, err)
	}
	if _, err := file.Write(content); err != nil {
		file.Close()
		_ = os.Remove(target)
		return fmt.Errorf("write hook %s: %w", target, err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(target)
		return fmt.Errorf("close hook %s: %w", target, err)
	}
	if err := os.Chmod(target, 0o755); err != nil {
		return fmt.Errorf("make hook executable: %w", err)
	}
	return nil
}

func (m *Manager) Uninstall(ctx context.Context, event string) error {
	target, content, err := m.target(ctx, event)
	if err != nil {
		return err
	}
	state, err := inspect(target, content)
	if err != nil {
		return err
	}
	switch state {
	case Missing:
		return nil
	case Conflict:
		return fmt.Errorf("refusing to remove unmanaged hook %s", target)
	default:
		if err := os.Remove(target); err != nil {
			return fmt.Errorf("remove hook %s: %w", target, err)
		}
		return nil
	}
}

func (m *Manager) Status(ctx context.Context, event string) (State, error) {
	target, content, err := m.target(ctx, event)
	if err != nil {
		return "", err
	}
	return inspect(target, content)
}

func (m *Manager) target(ctx context.Context, event string) (string, []byte, error) {
	if event != PreCommit {
		return "", nil, fmt.Errorf("unsupported hook %q", event)
	}
	hooksDir, err := m.repository.HooksDir(ctx)
	if err != nil {
		return "", nil, err
	}
	target := filepath.Join(hooksDir, event)
	content := []byte(fmt.Sprintf("#!/bin/sh\n%s\nexec %s hook run pre-commit --root %s\n",
		marker, shellQuote(m.binary), shellQuote(m.repository.Root())))
	return target, content, nil
}

func inspect(path string, expected []byte) (State, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return Missing, nil
	}
	if err != nil {
		return "", fmt.Errorf("inspect hook %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return Conflict, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read hook %s: %w", path, err)
	}
	if bytes.Equal(content, expected) {
		return Installed, nil
	}
	return Conflict, nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
