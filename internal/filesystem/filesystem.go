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

func SessionFSExist(cfg config.Config, sessionID string) (bool, error) {
	if sessionID == "" {
		return false, errors.New("sessionID cannot be empty")
	}
	if !validSessionID.MatchString(sessionID) {
		return false, errors.New("invalid sessionID format")
	}

	base, err := filepath.Abs(cfg.Execution.WorkingDirectoryBase)
	if err != nil {
		return false, fmt.Errorf("failed to resolve base path: %w", err)
	}

	workspacePath := filepath.Join(base, sessionID)

	rel, err := filepath.Rel(base, workspacePath)
	if err != nil {
		return false, fmt.Errorf("path resolution failed: %w", err)
	}

	if strings.HasPrefix(rel, "..") {
		return false, fmt.Errorf("path traversal detected")
	}

	info, err := os.Stat(workspacePath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("stat failed: %w", err)
	}

	if !info.IsDir() {
		return false, errors.New("not a directory")
	}

	return true, nil
}

func NewFSWorkspace(cfg config.Config, sessionID string) (*FSWorkspace, error) {

	if sessionID == "" {
		return nil, errors.New("sessionID cannot be empty")
	}
	if !validSessionID.MatchString(sessionID) {
		return nil, fmt.Errorf("invalid sessionID format")
	}

	base, err := filepath.Abs(cfg.Execution.WorkingDirectoryBase)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base path: %w", err)
	}

	base = filepath.Clean(base)

	workspacePath := filepath.Join(base, sessionID)
	workspacePath = filepath.Clean(workspacePath)

	if !strings.HasPrefix(workspacePath, base+string(os.PathSeparator)) {
		return nil, fmt.Errorf("path traversal detected")
	}

	// check if dir exist
	info, err := os.Stat(workspacePath)
	if err == nil {
		if !info.IsDir() {
			return nil, fmt.Errorf("workspace path exists but is not a directory")
		}

		return &FSWorkspace{
			Mount: workspacePath,
			base:  base,
		}, nil
	}

	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to check workspace: %w", err)
	}

	// create only if not exists
	err = os.MkdirAll(workspacePath, 0700)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	_ = os.Chown(workspacePath, 1000, 1000)

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

	if strings.Contains(rel, "..") {
		return "", fmt.Errorf("invalid path")
	}

	rel = filepath.Clean("/" + rel)
	rel = strings.TrimPrefix(rel, "/")

	full := filepath.Join(fs.Mount, rel)
	full = filepath.Clean(full)

	mountResolved, err := filepath.EvalSymlinks(fs.Mount)
	if err != nil {
		return "", fmt.Errorf("invalid mount")
	}

	if full == fs.Mount {
		return mountResolved, nil
	}

	relPath, err := filepath.Rel(mountResolved, full)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("invalid path")
	}

	dir := filepath.Dir(full)
	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", fmt.Errorf("invalid path")
	}

	resolved := filepath.Join(resolvedDir, filepath.Base(full))

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
