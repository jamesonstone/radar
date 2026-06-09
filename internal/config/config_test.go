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
