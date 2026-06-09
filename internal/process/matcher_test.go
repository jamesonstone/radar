package process

import (
	"testing"

	"github.com/jamesonstone/radar/internal/config"
	"github.com/jamesonstone/radar/internal/domain"
)

func TestMatchProcess(t *testing.T) {
	p := domain.ProcessSnapshot{Name: "claude", Cmdline: "/usr/local/bin/claude --version", Exe: "/usr/local/bin/claude"}
	m := config.MatchConfig{Names: []string{"Claude"}, NameContains: []string{"claud"}, CommandContains: []string{"--version"}}
	if !MatchProcess(p, m) {
		t.Fatal("expected process to match")
	}
}

func TestMatchTargetsAssignsLabel(t *testing.T) {
	snapshots := []domain.ProcessSnapshot{{PID: 1, Name: "Warp"}}
	targets := []config.TargetConfig{{Label: "Warp", Match: config.MatchConfig{Names: []string{"Warp"}}}}
	matches := MatchTargets(snapshots, targets)
	if len(matches) != 1 || matches[0].Label != "Warp" {
		t.Fatalf("unexpected match result: %+v", matches)
	}
}
