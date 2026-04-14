package config

import (
	"errors"
	"fmt"
	"os"
)

func (c *Config) Validate() error {

	//  Server
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return errors.New("server.port must be between 1 and 65535")
	}

	if c.Execution.MaxConcurrent <= 0 {
		return errors.New("execution.max_concurrent_sessions must be > 0")
	}

	if c.Execution.SessionTTLMinutes <= 0 {
		return errors.New("execution.session_ttl_minutes must be > 0")
	}

	if c.Execution.SessionCleanupIntervalSec <= 0 {
		return errors.New("execution.session_cleanup_interval_sec must be > 0")
	}

	if c.Execution.WorkingDirectoryBase == "" {
		return errors.New("execution.working_directory_base cannot be empty")
	}

	// Ensure working directory exists or is creatable
	if err := ensureDir(c.Execution.WorkingDirectoryBase); err != nil {
		return fmt.Errorf("invalid working_directory_base: %w", err)
	}

	// sandbox settings
	if !c.Sandbox.Enabled {
		return errors.New("sandbox mode selected but sandbox.enabled is false")
	}

	if c.Sandbox.Image == "" {
		return errors.New("sandbox.image cannot be empty")
	}

	if c.Sandbox.LoadEnv && c.Execution.EnvFilePath == "" {
		return errors.New("execution.envfilepath cannot be empty if sandbox.loadenv is true")
	}

	if c.Sandbox.Memory == 0 {
		return errors.New("sandbox.memory cannot be empty")
	}

	if c.Sandbox.CPUs == 0 {
		return errors.New("sandbox.cpus cannot be empty")
	}

	if c.Sandbox.PIDs <= 0 {
		return errors.New("sandbox.pids must be > 0")
	}

	if c.Sandbox.JobTTL <= 0 {
		return errors.New("sandbox.job_ttl must be > 0")
	}

	// File System
	if c.FileSystem.MaxUploadSize <= 0 {
		return errors.New("filesystem.max_upload_size_mbs must be > 0")
	}

	//  Logging
	switch c.Logging.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("invalid logging.level: %s", c.Logging.Level)
	}

	switch c.Logging.Format {
	case "json", "text":
	default:
		return fmt.Errorf("invalid logging.format: %s", c.Logging.Level)
	}

	return nil
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}
