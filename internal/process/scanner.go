package process

import (
	"time"

	"github.com/jamesonstone/radar/internal/domain"
	"github.com/shirou/gopsutil/v4/process"
)

type Scanner struct{}

func NewScanner() *Scanner { return &Scanner{} }

func (s *Scanner) Scan() ([]domain.ProcessSnapshot, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	out := make([]domain.ProcessSnapshot, 0, len(procs))
	for _, p := range procs {
		name, _ := p.Name()
		exe, _ := p.Exe()
		cmdline, _ := p.Cmdline()
		cwd, _ := p.Cwd()
		ppid, _ := p.Ppid()
		startedMS, _ := p.CreateTime()
		startedAt := time.UnixMilli(startedMS)
		if startedMS == 0 {
			startedAt = now
		}
		cpuPercent, _ := p.CPUPercent()
		memInfo, _ := p.MemoryInfo()
		var rss uint64
		if memInfo != nil {
			rss = memInfo.RSS
		}
		out = append(out, domain.ProcessSnapshot{
			PID:        int(p.Pid),
			PPID:       int(ppid),
			Name:       name,
			Exe:        exe,
			Cmdline:    cmdline,
			CWD:        cwd,
			StartedAt:  startedAt,
			SeenAt:     now,
			CPUPercent: cpuPercent,
			MemoryRSS:  rss,
		})
	}
	return out, nil
}
