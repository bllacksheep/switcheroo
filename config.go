package main

import (
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const configYaml string = "~/.switcheroo.yaml"

// default switcheroo AppDir can be overwritten in ~/.switcheroo.yaml
const gitConfigs = "~/.local/share/switcheroo"

type Credential struct {
	Name   string `yaml:"name"`
	Helper string `yaml:"helper"`
}

type Alias struct {
	Name        string       `yaml:"name"`
	Account     string       `yaml:"account"`
	Email       string       `yaml:"email"`
	SigningKey  string       `yaml:"signingkey,omitempty"`
	Credentials []Credential `yaml:"credentials,omitempty"`
	Deafult     bool         `yaml:"default"`
}

type Config struct {
	Switcheroo struct {
		GitConfigs string  `yaml:"gitconfigs,omitempty"`
		Aliases    []Alias `yaml:"aliases"`
	} `yaml:"switcheroo"`
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

func validateConfig() (*Config, error) {
	slog.Debug("main.go", "main", "validating config")
	config := &Config{}

	switcherooConfigFilePath, err := expandPath(configYaml)

	cf, err := os.ReadFile(switcherooConfigFilePath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(cf, &config); err != nil {
		return nil, err
	}

	if config.Switcheroo.GitConfigs == "" {
		config.Switcheroo.GitConfigs, err = expandPath(gitConfigs)
	} else {
		config.Switcheroo.GitConfigs, err = expandPath(config.Switcheroo.GitConfigs)
	}

	if err != nil {
		return nil, err
	}

	return config, nil
}
