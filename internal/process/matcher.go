package process

import (
	"strings"

	"github.com/jamesonstone/radar/internal/config"
	"github.com/jamesonstone/radar/internal/domain"
)

func MatchTargets(snapshots []domain.ProcessSnapshot, targets []config.TargetConfig) []domain.ProcessSnapshot {
	matches := make([]domain.ProcessSnapshot, 0)
	for _, p := range snapshots {
		for _, t := range targets {
			if MatchProcess(p, t.Match) {
				p.Label = t.Label
				matches = append(matches, p)
				break
			}
		}
	}
	return matches
}

func MatchProcess(p domain.ProcessSnapshot, m config.MatchConfig) bool {
	for _, v := range m.Names {
		if p.Name == v {
			return true
		}
	}
	for _, v := range m.NameContains {
		if strings.Contains(strings.ToLower(p.Name), strings.ToLower(v)) {
			return true
		}
	}
	for _, v := range m.CommandContains {
		if strings.Contains(strings.ToLower(p.Cmdline), strings.ToLower(v)) {
			return true
		}
	}
	for _, v := range m.ExeContains {
		if strings.Contains(strings.ToLower(p.Exe), strings.ToLower(v)) {
			return true
		}
	}
	return false
}
