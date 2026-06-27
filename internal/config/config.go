package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Steps struct {
	PermissionSchemes    bool `yaml:"permission_schemes"`
	SchemeProjects       bool `yaml:"scheme_projects"`
	SchemePermissions    bool `yaml:"scheme_permissions"`
	ProjectRoleActors    bool `yaml:"project_role_actors"`
	FinalFlatExport      bool `yaml:"final_flat_export"`
	SplitFinalFlatExport bool `yaml:"split_final_flat_export"`
}

type Config struct {
	BaseURL   string `yaml:"base_url"`
	Email     string `yaml:"email"`
	APIToken  string `yaml:"api_token"`
	OutputDir string `yaml:"output_dir"`
	Steps     Steps  `yaml:"steps"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")

	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("config.base_url is required")
	}
	if cfg.Email == "" {
		return nil, fmt.Errorf("config.email is required")
	}
	if cfg.APIToken == "" {
		return nil, fmt.Errorf("config.api_token is required")
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "./output"
	}

	return &cfg, nil
}
