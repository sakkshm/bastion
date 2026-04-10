package filesystem

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sakkshm/bastion/internal/config"
)

type FSWorkspace struct {
	Mount string
	base  string
}

var validSessionID = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func NewFSWorkspace(cfg config.Config, sessionID string) (*FSWorkspace, error) {

	if sessionID == "" {
		return nil, errors.New("sessionID cannot be empty")
	}
	if !validSessionID.MatchString(sessionID) {
		return nil, fmt.Errorf("invalid sessionID format")
	}

	// resolve absolute base path
	base, err := filepath.Abs(cfg.Execution.WorkingDirectoryBase)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base path: %w", err)
	}

	base = filepath.Clean(base)

	// construct workspace path
	workspacePath := filepath.Join(base, sessionID)
	workspacePath = filepath.Clean(workspacePath)

	// prevent path traversal
	if !strings.HasPrefix(workspacePath, base+string(os.PathSeparator)) {
		return nil, fmt.Errorf("path traversal detected")
	}

	// create directory with restrictive permissions
	err = os.MkdirAll(workspacePath, 0700)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	// add permisions
	os.Chown(workspacePath, 1000, 1000)

	return &FSWorkspace{
		Mount: workspacePath,
		base:  base,
	}, nil
}

func (fs *FSWorkspace) DeleteWorkspace() error {
	if fs.Mount == "" {
		return errors.New("invalid workspace path")
	}

	cleanBase := filepath.Clean(fs.base)
	cleanMount := filepath.Clean(fs.Mount)

	// ensure deletion is inside base directory
	if !strings.HasPrefix(cleanMount, cleanBase+string(os.PathSeparator)) {
		return fmt.Errorf("refusing to delete outside base directory")
	}

	// prevent symlink attacks
	info, err := os.Lstat(cleanMount)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // already gone
		}
		return err
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to delete symlink")
	}

	return os.RemoveAll(cleanMount)
}

func (fs *FSWorkspace) SafePath(rel string) (string, error) {
	full := filepath.Join(fs.Mount, rel)
	full = filepath.Clean(full)

	// Basic prefix check (existing)
	if !strings.HasPrefix(full, fs.Mount+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid path")
	}

	// symlink resolution
	resolved, err := filepath.EvalSymlinks(full)
	if err != nil {
		return "", fmt.Errorf("invalid path")
	}

	// Resolve mount as well
	mountResolved, err := filepath.EvalSymlinks(fs.Mount)
	if err != nil {
		return "", fmt.Errorf("invalid mount")
	}

	// Ensure resolved path is still inside mount
	if !strings.HasPrefix(resolved, mountResolved+string(os.PathSeparator)) &&
		resolved != mountResolved {
		return "", fmt.Errorf("path escapes sandbox")
	}

	return resolved, nil
}

func (fs FSWorkspace) FSExists() bool {
	if _, err := os.Stat(fs.Mount); os.IsNotExist(err) {
		return false
	}

	return true
}
