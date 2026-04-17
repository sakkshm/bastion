package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/pelletier/go-toml/v2"
)

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	cfg.applyDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func LoadEnvToDocker(envPath string) ([]string, error) {
	envMap, err := godotenv.Read(envPath)
	if err != nil {
		return nil, err
	}

	var envList []string
	for k, v := range envMap {
		envList = append(envList, fmt.Sprintf("%s=%s", k, v))
	}

	return envList, nil
}
