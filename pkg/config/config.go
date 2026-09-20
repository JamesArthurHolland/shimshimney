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
		path = "/etc/shimshimney/config.yaml"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	cfg.OperatorURL = strings.TrimSpace(cfg.OperatorURL)
	cfg.Build = strings.TrimSpace(cfg.Build)
	cfg.Run = strings.TrimSpace(cfg.Run)
	cfg.BuildMode = strings.TrimSpace(cfg.BuildMode)
	if cfg.BuildMode == "" {
		cfg.BuildMode = "hot"
	}
	return cfg, nil
}

func (c Config) IsHot() bool {
	return strings.EqualFold(c.BuildMode, "hot") || c.BuildMode == "SHIMNEY_MODE=hot"
}

func (c Config) IsCold() bool {
	return strings.EqualFold(c.BuildMode, "cold") || c.BuildMode == "SHIMNEY_MODE=cold"
}
