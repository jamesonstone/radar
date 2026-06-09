package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jamesonstone/radar/internal/config"
	"github.com/jamesonstone/radar/internal/domain"
)

type tickMsg time.Time

type pollMsg struct {
	result domain.PollResult
	err    error
}

type Poller interface {
	Poll() (domain.PollResult, error)
}

type Model struct {
	cfg         config.Config
	service     Poller
	showEvents  bool
	showHistory bool
	showHelp    bool
	lastPoll    domain.PollResult
	err         error
}

func NewModel(cfg config.Config, service Poller) Model {
	return Model{cfg: cfg, service: service, showEvents: cfg.UI.ShowEvents, showHistory: true}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.pollCmd(), tickCmd(m.cfg.PollInterval.Duration))
}

func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m Model) pollCmd() tea.Cmd {
	return func() tea.Msg {
		res, err := m.service.Poll()
		return pollMsg{result: res, err: err}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		return m, tea.Batch(m.pollCmd(), tickCmd(m.cfg.PollInterval.Duration))
	case pollMsg:
		m.lastPoll = msg.result
		m.err = msg.err
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			return m, m.pollCmd()
		case "e":
			m.showEvents = !m.showEvents
		case "h":
			m.showHistory = !m.showHistory
		case "?":
			m.showHelp = !m.showHelp
		}
	}
	return m, nil
}

func (m Model) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true)
	var b strings.Builder
	b.WriteString(headerStyle.Render(fmt.Sprintf("radar\tUpdated %s", time.Now().Format("15:04:05"))))
	b.WriteString("\n")
	b.WriteString("STATUS   AGENT             PID     RUNTIME    CPU    MEM     PROCESS\n")
	for _, s := range m.lastPoll.Running {
		b.WriteString(fmt.Sprintf("● RUN    %-16s %-7d %-10s %-6s %-7s %s\n", s.Label, s.PID, s.Runtime.Round(time.Second), "-", "-", s.Name))
	}
	if m.cfg.UI.ShowExitedRecent && m.showHistory {
		cutoff := time.Now().Add(-m.cfg.UI.ExitedRecentWindow.Duration)
		for _, s := range m.lastPoll.Exited {
			if s.EndedAt == nil || s.EndedAt.Before(cutoff) {
				continue
			}
			b.WriteString(fmt.Sprintf("○ EXIT   %-16s %-7d %-10s %-6s %-7s %s\n", s.Label, s.PID, s.Runtime.Round(time.Second), "-", "-", s.Name))
		}
	}
	if m.showEvents {
		b.WriteString("\nEvents\n")
		for _, e := range m.lastPoll.Events {
			b.WriteString(fmt.Sprintf("%s  %-12s %s\n", e.At.Format("15:04:05"), e.Type, e.Message))
		}
	}
	if m.showHelp {
		b.WriteString("\nq quit  r refresh  h toggle history  e toggle events  ? help\n")
	}
	if m.err != nil {
		b.WriteString("\nerror: ")
		b.WriteString(m.err.Error())
		b.WriteString("\n")
	}
	return b.String()
}
