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

	//  Execution Mode 
	if c.Execution.Mode != "sandbox" && c.Execution.Mode != "bare_metal" {
		return errors.New("execution.mode must be 'sandbox' or 'bare_metal'")
	}

	if c.Execution.MaxConcurrent <= 0 {
		return errors.New("execution.max_concurrent_sessions must be > 0")
	}

	if c.Execution.SessionTTLMinutes <= 0 {
		return errors.New("execution.session_ttl_minutes must be > 0")
	}

	if c.Execution.WorkingDirectoryBase == "" {
		return errors.New("execution.working_directory_base cannot be empty")
	}

	// Ensure working directory exists or is creatable
	if err := ensureDir(c.Execution.WorkingDirectoryBase); err != nil {
		return fmt.Errorf("invalid working_directory_base: %w", err)
	}

	//  Bare Metal 
	if c.Execution.Mode == "bare_metal" {
		if !c.BareMetal.Enabled {
			return errors.New("bare_metal mode selected but bare_metal.enabled is false")
		}

		if len(c.BareMetal.AllowedCommands) == 0 {
			return errors.New("bare_metal.allowed_commands cannot be empty in bare_metal mode")
		}
	}

	//  Sandbox 
	if c.Execution.Mode == "sandbox" {
		if !c.Sandbox.Enabled {
			return errors.New("sandbox mode selected but sandbox.enabled is false")
		}

		if c.Sandbox.Image == "" {
			return errors.New("sandbox.image cannot be empty")
		}

		if c.Sandbox.Memory == "" {
			return errors.New("sandbox.memory cannot be empty")
		}

		if c.Sandbox.CPUs == "" {
			return errors.New("sandbox.cpus cannot be empty")
		}

		if c.Sandbox.PIDs <= 0 {
			return errors.New("sandbox.pids must be > 0")
		}
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
