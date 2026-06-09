package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingReturnsDefault(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(filepath.Join(dir, "missing.yaml"))
	if err != nil {
		t.Fatalf("load returned error: %v", err)
	}
	if cfg.PollInterval.Duration.String() != "2s" {
		t.Fatalf("unexpected poll interval: %s", cfg.PollInterval.Duration)
	}
}

func TestWriteDefaultAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := WriteDefault(path); err != nil {
		t.Fatalf("write default: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat config: %v", err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if len(cfg.Targets) < 4 {
		t.Fatalf("expected default targets, got %d", len(cfg.Targets))
	}
}

func TestConfigAndDBPathsDefaultToCWD(t *testing.T) {
	dir := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldwd); err != nil {
			t.Logf("restore cwd: %v", err)
		}
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	configPath, err := ConfigPath()
	if err != nil {
		t.Fatalf("config path: %v", err)
	}
	if got, want := configPath, filepath.Join(dir, ".radar", "config.yaml"); got != want {
		t.Fatalf("config path = %q, want %q", got, want)
	}

	dbPath, err := DBPath()
	if err != nil {
		t.Fatalf("db path: %v", err)
	}
	if got, want := dbPath, filepath.Join(dir, ".radar", "radar.sqlite3"); got != want {
		t.Fatalf("db path = %q, want %q", got, want)
	}
}

func TestConfigAndDBPathsSupportEnvOverride(t *testing.T) {
	t.Setenv("RADAR_CONFIG_PATH", "/tmp/custom-config.yaml")
	t.Setenv("RADAR_DB_PATH", "/tmp/custom-radar.sqlite3")

	configPath, err := ConfigPath()
	if err != nil {
		t.Fatalf("config path: %v", err)
	}
	if got, want := configPath, "/tmp/custom-config.yaml"; got != want {
		t.Fatalf("config path = %q, want %q", got, want)
	}

	dbPath, err := DBPath()
	if err != nil {
		t.Fatalf("db path: %v", err)
	}
	if got, want := dbPath, "/tmp/custom-radar.sqlite3"; got != want {
		t.Fatalf("db path = %q, want %q", got, want)
	}
}
