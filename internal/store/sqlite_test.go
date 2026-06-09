package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/jamesonstone/radar/internal/domain"
)

func TestUpsertAndMarkExited(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "agent-watch.sqlite3")
	st, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	now := time.Now().Add(-time.Minute)
	snapshot := domain.ProcessSnapshot{
		PID: 42, Label: "Codex", Name: "Codex", StartedAt: now,
		SeenAt: now.Add(5 * time.Second),
	}
	sess, created, err := st.UpsertRunning(snapshot)
	if err != nil {
		t.Fatalf("upsert running: %v", err)
	}
	if !created {
		t.Fatal("expected first upsert to create session")
	}
	exited, err := st.MarkExited(map[string]struct{}{}, time.Now())
	if err != nil {
		t.Fatalf("mark exited: %v", err)
	}
	if len(exited) != 1 || exited[0].ID != sess.ID || exited[0].EndedAt == nil {
		t.Fatalf("unexpected exited sessions: %+v", exited)
	}
}
