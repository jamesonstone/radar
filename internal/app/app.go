package app

import (
	"fmt"
	"os"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/jamesonstone/radar/internal/config"
	"github.com/jamesonstone/radar/internal/domain"
	"github.com/jamesonstone/radar/internal/notify"
	"github.com/jamesonstone/radar/internal/process"
	"github.com/jamesonstone/radar/internal/store"
	"github.com/jamesonstone/radar/internal/tui"
)

type Service struct {
	cfg      config.Config
	scanner  *process.Scanner
	store    *store.SQLiteStore
	notifier *notify.Notifier
	longSent map[string]bool
}

func NewService(cfg config.Config, st *store.SQLiteStore) *Service {
	return &Service{cfg: cfg, scanner: process.NewScanner(), store: st, notifier: notify.New(), longSent: map[string]bool{}}
}

func (s *Service) Snapshot() ([]domain.ProcessSnapshot, error) {
	snapshots, err := s.scanner.Scan()
	if err != nil {
		return nil, err
	}
	matches := process.MatchTargets(snapshots, s.cfg.Targets)
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Label == matches[j].Label {
			return matches[i].PID < matches[j].PID
		}
		return matches[i].Label < matches[j].Label
	})
	return matches, nil
}

func (s *Service) Poll() (domain.PollResult, error) {
	now := time.Now()
	matches, err := s.Snapshot()
	if err != nil {
		return domain.PollResult{}, err
	}
	active := map[string]struct{}{}
	for _, p := range matches {
		sess, created, err := s.store.UpsertRunning(p)
		if err != nil {
			return domain.PollResult{}, err
		}
		active[sess.ID] = struct{}{}
		if created {
			msg := fmt.Sprintf("started %s pid=%d", sess.Label, sess.PID)
			_ = s.store.InsertEvent(sess.ID, "started", msg, now)
			if s.cfg.Notifications.Enabled && s.cfg.Notifications.OnStart {
				_ = s.notifier.Notify("radar", fmt.Sprintf("%s started", sess.Label))
			}
		}
		runtime := now.Sub(sess.StartedAt)
		if s.cfg.Notifications.Enabled && s.cfg.Notifications.OnLongRunning && runtime >= s.cfg.Notifications.LongRunningThreshold.Duration && !s.longSent[sess.ID] {
			s.longSent[sess.ID] = true
			msg := fmt.Sprintf("%s has been running for %s", sess.Label, runtime.Round(time.Second))
			_ = s.store.InsertEvent(sess.ID, "long_running", msg, now)
			_ = s.notifier.Notify("radar", msg)
		}
	}
	exited, err := s.store.MarkExited(active, now)
	if err != nil {
		return domain.PollResult{}, err
	}
	for _, sess := range exited {
		delete(s.longSent, sess.ID)
		msg := fmt.Sprintf("stopped %s pid=%d runtime=%s", sess.Label, sess.PID, sess.Runtime.Round(time.Second))
		_ = s.store.InsertEvent(sess.ID, "stopped", msg, now)
		if s.cfg.Notifications.Enabled && s.cfg.Notifications.OnStop {
			_ = s.notifier.Notify("radar", fmt.Sprintf("%s stopped after %s", sess.Label, sess.Runtime.Round(time.Second)))
		}
	}
	runningMap, err := s.store.ActiveSessions()
	if err != nil {
		return domain.PollResult{}, err
	}
	running := make([]domain.Session, 0, len(runningMap))
	for _, sess := range runningMap {
		sess.Runtime = now.Sub(sess.StartedAt)
		running = append(running, sess)
	}
	sort.Slice(running, func(i, j int) bool { return running[i].StartedAt.Before(running[j].StartedAt) })
	events, _ := s.store.ListEvents(20)
	return domain.PollResult{Running: running, Exited: exited, Events: events}, nil
}

func Execute() {
	root := &cobra.Command{
		Use:   "rdr",
		Short: "Watch coding-agent processes on macOS",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, cfg, cleanup, err := initService()
			if err != nil {
				return err
			}
			defer cleanup()
			m := tui.NewModel(cfg, svc)
			p := tea.NewProgram(m)
			_, err = p.Run()
			return err
		},
	}

	root.AddCommand(onceCommand(), listCommand(), historyCommand(), configCommand())
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initService() (*Service, config.Config, func(), error) {
	configPath, err := config.ConfigPath()
	if err != nil {
		return nil, config.Config{}, nil, err
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, config.Config{}, nil, err
	}
	dbPath, err := config.DBPath()
	if err != nil {
		return nil, config.Config{}, nil, err
	}
	if err := config.EnsureParentDir(dbPath); err != nil {
		return nil, config.Config{}, nil, err
	}
	st, err := store.Open(dbPath)
	if err != nil {
		return nil, config.Config{}, nil, err
	}
	return NewService(cfg, st), cfg, func() { _ = st.Close() }, nil
}

func onceCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "once",
		Short: "Print a process snapshot and exit",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, _, cleanup, err := initService()
			if err != nil {
				return err
			}
			defer cleanup()
			snapshots, err := svc.Snapshot()
			if err != nil {
				return err
			}
			fmt.Println("STATUS\tAGENT\tPID\tRUNTIME\tCPU\tMEM\tPROCESS")
			now := time.Now()
			for _, p := range snapshots {
				fmt.Printf("RUN\t%s\t%d\t%s\t%.1f%%\t%s\t%s\n", p.Label, p.PID, now.Sub(p.StartedAt).Round(time.Second), p.CPUPercent, formatBytes(p.MemoryRSS), p.Name)
			}
			return nil
		},
	}
}

func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List current target process matches",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, _, cleanup, err := initService()
			if err != nil {
				return err
			}
			defer cleanup()
			snapshots, err := svc.Snapshot()
			if err != nil {
				return err
			}
			for _, p := range snapshots {
				fmt.Printf("%s pid=%d name=%q exe=%q cmdline=%q\n", p.Label, p.PID, p.Name, p.Exe, p.Cmdline)
			}
			return nil
		},
	}
}

func historyCommand() *cobra.Command {
	var limit int
	var agent string
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show recent sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _, cleanup, err := initService()
			if err != nil {
				return err
			}
			defer cleanup()
			dbPath, _ := config.DBPath()
			st, err := store.Open(dbPath)
			if err != nil {
				return err
			}
			defer st.Close()
			sessions, err := st.ListSessions(limit, agent)
			if err != nil {
				return err
			}
			for _, s := range sessions {
				status := "RUN"
				if s.EndedAt != nil {
					status = "EXIT"
				}
				fmt.Printf("%s\t%s\tpid=%d\tstart=%s\truntime=%s\n", status, s.Label, s.PID, s.StartedAt.Format(time.RFC3339), s.Runtime.Round(time.Second))
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 20, "max number of sessions")
	cmd.Flags().StringVar(&agent, "agent", "", "filter by agent label")
	return cmd
}

func configCommand() *cobra.Command {
	cfgCmd := &cobra.Command{Use: "config", Short: "Configuration commands"}
	cfgCmd.AddCommand(&cobra.Command{Use: "path", Short: "Print config path", RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.ConfigPath()
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	}})
	cfgCmd.AddCommand(&cobra.Command{Use: "init", Short: "Write default config", RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.ConfigPath()
		if err != nil {
			return err
		}
		if err := config.WriteDefault(path); err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	}})
	return cfgCmd
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.0f%ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
