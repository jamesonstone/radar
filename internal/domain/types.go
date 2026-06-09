package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type ProcessStatus string

const (
	StatusRunning ProcessStatus = "running"
	StatusExited  ProcessStatus = "exited"
)

type ProcessSnapshot struct {
	PID        int
	PPID       int
	Label      string
	Name       string
	Exe        string
	Cmdline    string
	CWD        string
	StartedAt  time.Time
	SeenAt     time.Time
	CPUPercent float64
	MemoryRSS  uint64
}

type Session struct {
	ID          string
	PID         int
	Label       string
	Name        string
	Exe         string
	Cmdline     string
	CWD         string
	StartedAt   time.Time
	FirstSeenAt time.Time
	LastSeenAt  time.Time
	EndedAt     *time.Time
	Runtime     time.Duration
}

type Event struct {
	At      time.Time
	Type    string
	Message string
}

type PollResult struct {
	Running []Session
	Exited  []Session
	Events  []Event
}

func SessionID(label string, pid int, startedAt time.Time) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%d:%s", label, pid, startedAt.UTC().Format(time.RFC3339Nano))))
	return hex.EncodeToString(h[:])
}
