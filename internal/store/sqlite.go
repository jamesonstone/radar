package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jamesonstone/radar/internal/domain"
	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func Open(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	s := &SQLiteStore{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *SQLiteStore) Close() error { return s.db.Close() }

func (s *SQLiteStore) migrate() error {
	if _, err := s.db.Exec(sessionsMigration); err != nil {
		return err
	}
	if _, err := s.db.Exec(eventsMigration); err != nil {
		return err
	}
	return nil
}

func (s *SQLiteStore) UpsertRunning(p domain.ProcessSnapshot) (domain.Session, bool, error) {
	id := domain.SessionID(p.Label, p.PID, p.StartedAt)
	now := p.SeenAt
	if now.IsZero() {
		now = time.Now()
	}
	created := false
	row := s.db.QueryRow(`select id from sessions where id=?`, id)
	var existingID string
	err := row.Scan(&existingID)
	if errors.Is(err, sql.ErrNoRows) {
		created = true
		_, err = s.db.Exec(`insert into sessions (id,pid,label,name,exe,cmdline,cwd,started_at,first_seen_at,last_seen_at,ended_at,created_at,updated_at)
values (?,?,?,?,?,?,?,?,?,?,NULL,?,?)`,
			id, p.PID, p.Label, p.Name, p.Exe, p.Cmdline, p.CWD,
			p.StartedAt.UTC().Format(time.RFC3339Nano), now.UTC().Format(time.RFC3339Nano), now.UTC().Format(time.RFC3339Nano), now.UTC().Format(time.RFC3339Nano), now.UTC().Format(time.RFC3339Nano))
		if err != nil {
			return domain.Session{}, false, err
		}
	} else if err != nil {
		return domain.Session{}, false, err
	} else {
		_, err = s.db.Exec(`update sessions set label=?,name=?,exe=?,cmdline=?,cwd=?,last_seen_at=?,updated_at=?,ended_at=NULL where id=?`,
			p.Label, p.Name, p.Exe, p.Cmdline, p.CWD, now.UTC().Format(time.RFC3339Nano), now.UTC().Format(time.RFC3339Nano), id)
		if err != nil {
			return domain.Session{}, false, err
		}
	}
	sess, err := s.GetSession(id)
	return sess, created, err
}

func (s *SQLiteStore) GetSession(id string) (domain.Session, error) {
	row := s.db.QueryRow(`select id,pid,label,name,exe,cmdline,cwd,started_at,first_seen_at,last_seen_at,ended_at from sessions where id=?`, id)
	var sess domain.Session
	var started, firstSeen, lastSeen string
	var ended sql.NullString
	if err := row.Scan(&sess.ID, &sess.PID, &sess.Label, &sess.Name, &sess.Exe, &sess.Cmdline, &sess.CWD, &started, &firstSeen, &lastSeen, &ended); err != nil {
		return domain.Session{}, err
	}
	var err error
	sess.StartedAt, err = time.Parse(time.RFC3339Nano, started)
	if err != nil {
		return domain.Session{}, err
	}
	sess.FirstSeenAt, err = time.Parse(time.RFC3339Nano, firstSeen)
	if err != nil {
		return domain.Session{}, err
	}
	sess.LastSeenAt, err = time.Parse(time.RFC3339Nano, lastSeen)
	if err != nil {
		return domain.Session{}, err
	}
	if ended.Valid {
		t, err := time.Parse(time.RFC3339Nano, ended.String)
		if err != nil {
			return domain.Session{}, err
		}
		sess.EndedAt = &t
		sess.Runtime = t.Sub(sess.StartedAt)
	} else {
		sess.Runtime = sess.LastSeenAt.Sub(sess.StartedAt)
	}
	return sess, nil
}

func (s *SQLiteStore) ActiveSessions() (map[string]domain.Session, error) {
	rows, err := s.db.Query(`select id,pid,label,name,exe,cmdline,cwd,started_at,first_seen_at,last_seen_at from sessions where ended_at is null`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]domain.Session{}
	for rows.Next() {
		var sess domain.Session
		var started, firstSeen, lastSeen string
		if err := rows.Scan(&sess.ID, &sess.PID, &sess.Label, &sess.Name, &sess.Exe, &sess.Cmdline, &sess.CWD, &started, &firstSeen, &lastSeen); err != nil {
			return nil, err
		}
		sess.StartedAt, _ = time.Parse(time.RFC3339Nano, started)
		sess.FirstSeenAt, _ = time.Parse(time.RFC3339Nano, firstSeen)
		sess.LastSeenAt, _ = time.Parse(time.RFC3339Nano, lastSeen)
		sess.Runtime = sess.LastSeenAt.Sub(sess.StartedAt)
		out[sess.ID] = sess
	}
	return out, rows.Err()
}

func (s *SQLiteStore) MarkExited(activeIDs map[string]struct{}, now time.Time) ([]domain.Session, error) {
	rows, err := s.db.Query(`select id from sessions where ended_at is null`)
	if err != nil {
		return nil, err
	}
	var pending []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return nil, err
		}
		if _, ok := activeIDs[id]; ok {
			continue
		}
		pending = append(pending, id)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	var exited []domain.Session
	for _, id := range pending {
		if _, err := s.db.Exec(`update sessions set ended_at=?,updated_at=? where id=? and ended_at is null`, now.UTC().Format(time.RFC3339Nano), now.UTC().Format(time.RFC3339Nano), id); err != nil {
			return nil, err
		}
		sess, err := s.GetSession(id)
		if err != nil {
			return nil, err
		}
		exited = append(exited, sess)
	}
	return exited, nil
}

func (s *SQLiteStore) InsertEvent(sessionID, eventType, message string, at time.Time) error {
	_, err := s.db.Exec(`insert into events (session_id,event_type,message,created_at) values (?,?,?,?)`, sessionID, eventType, message, at.UTC().Format(time.RFC3339Nano))
	return err
}

func (s *SQLiteStore) ListSessions(limit int, agent string) ([]domain.Session, error) {
	if limit <= 0 {
		limit = 20
	}
	query := `select id,pid,label,name,exe,cmdline,cwd,started_at,first_seen_at,last_seen_at,ended_at from sessions`
	args := []any{}
	if agent != "" {
		query += ` where label=?`
		args = append(args, agent)
	}
	query += ` order by started_at desc limit ?`
	args = append(args, limit)
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []domain.Session
	for rows.Next() {
		var sess domain.Session
		var started, firstSeen, lastSeen string
		var ended sql.NullString
		if err := rows.Scan(&sess.ID, &sess.PID, &sess.Label, &sess.Name, &sess.Exe, &sess.Cmdline, &sess.CWD, &started, &firstSeen, &lastSeen, &ended); err != nil {
			return nil, err
		}
		sess.StartedAt, _ = time.Parse(time.RFC3339Nano, started)
		sess.FirstSeenAt, _ = time.Parse(time.RFC3339Nano, firstSeen)
		sess.LastSeenAt, _ = time.Parse(time.RFC3339Nano, lastSeen)
		if ended.Valid {
			t, _ := time.Parse(time.RFC3339Nano, ended.String)
			sess.EndedAt = &t
			sess.Runtime = t.Sub(sess.StartedAt)
		} else {
			sess.Runtime = sess.LastSeenAt.Sub(sess.StartedAt)
		}
		sessions = append(sessions, sess)
	}
	return sessions, rows.Err()
}

func (s *SQLiteStore) ListEvents(limit int) ([]domain.Event, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.Query(`select event_type,message,created_at from events order by id desc limit ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := []domain.Event{}
	for rows.Next() {
		var e domain.Event
		var at string
		if err := rows.Scan(&e.Type, &e.Message, &at); err != nil {
			return nil, err
		}
		e.At, _ = time.Parse(time.RFC3339Nano, at)
		events = append(events, e)
	}
	return events, rows.Err()
}

func (s *SQLiteStore) DebugCounts() (int, error) {
	row := s.db.QueryRow(`select count(*) from sessions`)
	var c int
	if err := row.Scan(&c); err != nil {
		return 0, fmt.Errorf("count sessions: %w", err)
	}
	return c, nil
}
