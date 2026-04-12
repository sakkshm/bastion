package config

type Config struct {
	Server     ServerConfig
	Execution  ExecutionConfig
	Sandbox    SandboxConfig
	FileSystem FileSystemConfig
	Logging    LoggingConfig
}

type ServerConfig struct {
	Host string
	Port int
}

type ExecutionConfig struct {
	MaxConcurrent             int    `toml:"max_concurrent_sessions"`
	SessionTTLMinutes         int    `toml:"session_ttl_minutes"`
	SessionCleanupIntervalSec int    `toml:"session_cleanup_interval_sec"`
	WorkingDirectoryBase      string `toml:"working_directory_base"`
}

type SandboxConfig struct {
	Enabled        bool    `toml:"enabled"`
	Image          string  `toml:"image"` // image of containers for sandbox
	NetworkEnabled bool    `toml:"network_enabled"`
	Memory         int     `toml:"memory_mbs"`
	CPUs           float32 `toml:"cpus"`
	PIDs           int     `toml:"pids"`
	JobTTL         int     `toml:"job_ttl"`
}

type FileSystemConfig struct {
	MaxUploadSize        int    `toml:"max_upload_size_mbs"`
}


type LoggingConfig struct {
	Level  string `toml:"level"`  // debug | info | warn | error
	Format string `toml:"format"` // json | text
	File   string `toml:"file"`
}

func (c *Config) applyDefaults() {
	if c.Execution.SessionCleanupIntervalSec == 0 {
		c.Execution.SessionCleanupIntervalSec = 30
	}

	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}

	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}

	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
}
