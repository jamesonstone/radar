package store

const sessionsMigration = `
create table if not exists sessions (
    id text primary key,
    pid integer not null,
    label text not null,
    name text not null,
    exe text,
    cmdline text,
    cwd text,
    started_at text not null,
    first_seen_at text not null,
    last_seen_at text not null,
    ended_at text,
    created_at text not null,
    updated_at text not null
);
create index if not exists idx_sessions_label on sessions(label);
create index if not exists idx_sessions_started_at on sessions(started_at);
create index if not exists idx_sessions_ended_at on sessions(ended_at);
`

const eventsMigration = `
create table if not exists events (
    id integer primary key autoincrement,
    session_id text not null,
    event_type text not null,
    message text not null,
    created_at text not null
);
`
