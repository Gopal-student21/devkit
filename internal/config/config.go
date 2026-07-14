package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const ConfigFile = "devkit.json"

type ProjectConfig struct {
	Name     string            `json:"name"`
	Type     string            `json:"type,omitempty"`
	Plugins  []string          `json:"plugins,omitempty"`
	Theme    string            `json:"theme,omitempty"`
	Port     int               `json:"port,omitempty"`
	Services []ServiceConfig   `json:"services,omitempty"`
	Scripts  map[string]string `json:"scripts,omitempty"`
}

type ServiceConfig struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Port    int    `json:"port,omitempty"`
	Version string `json:"version,omitempty"`
}

type Plugin struct {
	Name        string       `json:"name"`
	Label       string       `json:"label"`
	Description string       `json:"description"`
	Icon        string       `json:"icon"`
	Commands    []CommandDef `json:"commands"`
}

type CommandDef struct {
	Name    string     `json:"name"`
	Label   string     `json:"label"`
	Action  string     `json:"action"`
	Args    []string   `json:"args,omitempty"`
	Confirm string     `json:"confirm,omitempty"`
	Params  []ParamDef `json:"params,omitempty"`
}

type ParamDef struct {
	Name    string `json:"name"`
	Label   string `json:"label"`
	Type    string `json:"type"`
	Default string `json:"default,omitempty"`
}

func Load() (*ProjectConfig, error) {
	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		return &ProjectConfig{
			Name:     "DevKit",
			Plugins:  []string{},
			Port:     8080,
			Services: []ServiceConfig{},
			Scripts:  map[string]string{},
		}, nil
	}
	var cfg ProjectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid devkit.json: %w", err)
	}
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	return &cfg, nil
}

func Save(cfg *ProjectConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigFile, data, 0644)
}

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".devkit")
}

func PluginsDir() string {
	return filepath.Join(ConfigDir(), "plugins")
}

func Register(root *cobra.Command) {
	root.AddCommand(&cobra.Command{
		Use:   "config",
		Short: "Show current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := Load()
			data, _ := json.MarshalIndent(cfg, "", "  ")
			fmt.Println(string(data))
		},
	})
}
