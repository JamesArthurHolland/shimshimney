package config

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	OperatorURL string `yaml:"operator_url"`
	Build       string `yaml:"build"`
	Run         string `yaml:"run"`
	BuildMode   string `yaml:"build_mode"`
}

func Load(path string) (Config, error) {
	cfg := Config{}
	if path == "" {
		paths := []string{"example/config.yaml", "/workspace/example/config.yaml", "/etc/shimshimney/config.yaml"}
		var err error
		for _, candidate := range paths {
			cfg, err = loadFile(candidate)
			if err == nil {
				return ApplyEnv(cfg), nil
			}
		}
		return cfg, err
	}
	cfg, err := loadFile(path)
	if err != nil {
		return cfg, err
	}
	return ApplyEnv(cfg), nil
}

func loadFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	cfg := Config{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return normalize(cfg), nil
}

func ApplyEnv(cfg Config) Config {
	if value := os.Getenv("SHIMNEY_OPERATOR_URL"); value != "" {
		cfg.OperatorURL = value
	}
	if value := os.Getenv("SHIMNEY_BUILD"); value != "" {
		cfg.Build = value
	}
	if value := os.Getenv("SHIMNEY_RUN"); value != "" {
		cfg.Run = value
	}
	if value := os.Getenv("SHIMNEY_BUILD_MODE"); value != "" {
		cfg.BuildMode = value
	} else if value := os.Getenv("SHIMNEY_MODE"); value != "" {
		cfg.BuildMode = value
	}
	return normalize(cfg)
}

func normalize(cfg Config) Config {
	cfg.OperatorURL = strings.TrimSpace(cfg.OperatorURL)
	cfg.Build = strings.TrimSpace(cfg.Build)
	cfg.Run = strings.TrimSpace(cfg.Run)
	cfg.BuildMode = strings.TrimSpace(cfg.BuildMode)
	if cfg.BuildMode == "" {
		cfg.BuildMode = "hot"
	}
	return cfg
}

func (c Config) IsHot() bool {
	return strings.EqualFold(c.BuildMode, "hot") || c.BuildMode == "SHIMNEY_MODE=hot"
}

func (c Config) IsCold() bool {
	return strings.EqualFold(c.BuildMode, "cold") || c.BuildMode == "SHIMNEY_MODE=cold"
}
