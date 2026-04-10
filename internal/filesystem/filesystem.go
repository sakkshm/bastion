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

    // Normalize input
    rel = filepath.Clean("/" + rel)
    rel = strings.TrimPrefix(rel, "/")

    full := filepath.Join(fs.Mount, rel)
    full = filepath.Clean(full)

    // Resolve mount once
    mountResolved, err := filepath.EvalSymlinks(fs.Mount)
    if err != nil {
        return "", fmt.Errorf("invalid mount")
    }

    // ROOT CASE
    if full == fs.Mount {
        return mountResolved, nil
    }

    // Pre-check (cheap containment)
    relPath, err := filepath.Rel(mountResolved, full)
    if err != nil || strings.HasPrefix(relPath, "..") {
        return "", fmt.Errorf("invalid path")
    }

    // Resolve symlinks ONLY if path exists
    resolved := full
    if _, err := os.Lstat(full); err == nil {
        resolved, err = filepath.EvalSymlinks(full)
        if err != nil {
            return "", fmt.Errorf("invalid path")
        }
    }

    // Final containment check
    relPath, err = filepath.Rel(mountResolved, resolved)
    if err != nil || strings.HasPrefix(relPath, "..") {
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

func (fs *FSWorkspace) ListWorkspace(path string) ([]FileEntry, error) {

	listDirPath, err := fs.SafePath(path)
	if err != nil {
		return nil, err
	}

	// check if dir exists
	info, err := os.Stat(listDirPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("not a directory")
	}

	entries, err := os.ReadDir(listDirPath)
	if err != nil {
		return nil, err
	}

	fileInfo := make([]FileEntry, 0, len(entries))

	for _, e := range entries {

		// skip symlink
		if e.Type()&os.ModeSymlink != 0 {
			continue
		}

		var file = FileEntry{
			Name:  e.Name(),
			IsDir: e.IsDir(),
		}

		info, err := e.Info()
		if err != nil {
			continue
		}
		file.Size = info.Size()
		file.Mode = info.Mode().String()
		file.ModTime = info.ModTime()

		fileInfo = append(fileInfo, file)
	}

	return fileInfo, nil
}
