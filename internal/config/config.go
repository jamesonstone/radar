package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode {
		return errors.New("duration must be a string")
	}
	parsed, err := time.ParseDuration(value.Value)
	if err != nil {
		return err
	}
	d.Duration = parsed
	return nil
}

func (d Duration) MarshalYAML() (interface{}, error) {
	return d.String(), nil
}

type MatchConfig struct {
	Names           []string `yaml:"names"`
	NameContains    []string `yaml:"name_contains,omitempty"`
	CommandContains []string `yaml:"command_contains,omitempty"`
	ExeContains     []string `yaml:"exe_contains,omitempty"`
}

type TargetConfig struct {
	Label string      `yaml:"label"`
	Match MatchConfig `yaml:"match"`
}

type UIConfig struct {
	ShowEvents         bool     `yaml:"show_events"`
	ShowExitedRecent   bool     `yaml:"show_exited_recent"`
	ExitedRecentWindow Duration `yaml:"exited_recent_window"`
}

type NotificationsConfig struct {
	Enabled              bool     `yaml:"enabled"`
	OnStart              bool     `yaml:"on_start"`
	OnStop               bool     `yaml:"on_stop"`
	OnLongRunning        bool     `yaml:"on_long_running"`
	LongRunningThreshold Duration `yaml:"long_running_threshold"`
}

type Config struct {
	PollInterval  Duration            `yaml:"poll_interval"`
	UI            UIConfig            `yaml:"ui"`
	Notifications NotificationsConfig `yaml:"notifications"`
	Targets       []TargetConfig      `yaml:"targets"`
}

func Default() Config {
	return Config{
		PollInterval: Duration{Duration: 2 * time.Second},
		UI: UIConfig{
			ShowEvents:         true,
			ShowExitedRecent:   true,
			ExitedRecentWindow: Duration{Duration: 15 * time.Minute},
		},
		Notifications: NotificationsConfig{
			Enabled:              true,
			OnStart:              true,
			OnStop:               true,
			OnLongRunning:        true,
			LongRunningThreshold: Duration{Duration: 2 * time.Hour},
		},
		Targets: []TargetConfig{
			{Label: "Codex", Match: MatchConfig{Names: []string{"Codex"}, CommandContains: []string{"Codex.app"}}},
			{Label: "GitHub Copilot", Match: MatchConfig{Names: []string{"GitHub Copilot", "Copilot"}}},
			{Label: "Warp", Match: MatchConfig{Names: []string{"Warp"}}},
			{Label: "Claude Code", Match: MatchConfig{Names: []string{"claude"}, CommandContains: []string{"claude"}}},
		},
	}
}

func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "radar", "config.yaml"), nil
}

func DBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "radar", "radar.sqlite3"), nil
}

func EnsureParentDir(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0o755)
}

func Load(path string) (Config, error) {
	cfg := Default()
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return Config{}, err
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func WriteDefault(path string) error {
	if err := EnsureParentDir(path); err != nil {
		return err
	}
	cfg := Default()
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("write default config: %w", err)
	}
	return nil
}
