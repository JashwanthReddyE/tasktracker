-- Schema for tasktracker. Applied at startup; every statement is idempotent.
--
-- Ownership rule: every row a user can reach is reachable only through a
-- user_id, either directly or via its parent task. Handlers must never filter
-- by user in Go -- it belongs in the WHERE clause.

CREATE TABLE IF NOT EXISTS users (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    email        TEXT    NOT NULL UNIQUE COLLATE NOCASE,
    display_name TEXT    NOT NULL DEFAULT '',
    created_at   TEXT    NOT NULL,
    -- Free-form JSON blob of per-account UI state (e.g. the last active
    -- category tab), so a session resumes where it left off on any device.
    prefs        TEXT    NOT NULL DEFAULT '{}'
);

-- One row per way of signing in. Password login is stored as
-- (provider='password', subject=email, password_hash=<bcrypt>). Adding Google
-- later inserts (provider='google', subject=<sub>, password_hash=NULL) against
-- the same user_id -- no schema change, and one user may hold both.
CREATE TABLE IF NOT EXISTS identities (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id       INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider      TEXT    NOT NULL,
    subject       TEXT    NOT NULL COLLATE NOCASE,
    password_hash TEXT,
    created_at    TEXT    NOT NULL,
    UNIQUE (provider, subject)
);
CREATE INDEX IF NOT EXISTS idx_identities_user ON identities(user_id);

-- Server-side sessions, so logout genuinely revokes access.
CREATE TABLE IF NOT EXISTS sessions (
    id         TEXT    PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TEXT    NOT NULL,
    expires_at TEXT    NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sessions_user    ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

-- Category ids are client-generated slugs ('work'), unique per user, so the
-- primary key is composite rather than a surrogate.
CREATE TABLE IF NOT EXISTS categories (
    user_id  INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    id       TEXT    NOT NULL,
    label    TEXT    NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, id)
);

-- People are assignment labels, not accounts. Scoped to a category, matching
-- the allPeople[categoryId] shape the client already uses.
CREATE TABLE IF NOT EXISTS people (
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    id          TEXT    NOT NULL,
    category_id TEXT    NOT NULL,
    name        TEXT    NOT NULL,
    hue         INTEGER NOT NULL DEFAULT 220,
    position    INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, id)
);
CREATE INDEX IF NOT EXISTS idx_people_cat ON people(user_id, category_id);

-- 'status' holds the kanban column ('todo'/'working'/'done'); it is exposed to
-- the client as "row". ROW and ORDER are SQL keywords, hence status/position.
CREATE TABLE IF NOT EXISTS tasks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       TEXT    NOT NULL,
    notes       TEXT    NOT NULL DEFAULT '',
    priority    TEXT    NOT NULL DEFAULT 'medium',
    status      TEXT    NOT NULL DEFAULT 'todo',
    category_id TEXT    NOT NULL DEFAULT '',
    due_date    TEXT    NOT NULL DEFAULT '',
    created_at  TEXT    NOT NULL,
    archived    INTEGER NOT NULL DEFAULT 0,
    position    INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_tasks_user ON tasks(user_id, archived, category_id);

CREATE TABLE IF NOT EXISTS task_people (
    task_id   INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    person_id TEXT    NOT NULL,
    position  INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (task_id, person_id)
);

CREATE TABLE IF NOT EXISTS events (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    type    TEXT    NOT NULL,
    time    TEXT    NOT NULL,
    text    TEXT    NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_events_task ON events(task_id, id);
