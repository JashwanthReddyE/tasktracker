# Task Tracker

A small, self-hosted kanban task tracker. Go backend, no build step, no framework — one static binary and one HTML file.

Accounts are built in, so your board follows you across browsers and devices. Each account's tasks, categories and people are private to it.

## Features

- Kanban board with todo / working / done columns, drag to reorder
- Categories, priorities, due dates and an archive
- People you can assign to tasks, with a per-person filter
- A per-task event log — comments and an automatic audit trail of moves
- Search, sort, CSV export, light and dark themes
- Email + password accounts with server-side sessions

## Quickstart

Requires Go 1.25+.

```bash
git clone https://github.com/Tapegossip/tasktracker.git
cd tasktracker
go run .
```

Open <http://127.0.0.1:8899>, create an account, and you're in.

To build a standalone binary:

```bash
CGO_ENABLED=0 go build -o tasktracker .
./tasktracker
```

`CGO_ENABLED=0` works because the SQLite driver ([`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite)) is pure Go — there's no C toolchain requirement and the result is a single self-contained executable.

> The server serves `static/` from the working directory, so run the binary from the project root (or copy `static/` next to it).

## Configuration

All configuration is via environment variables.

| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `8899` | Port to listen on. A bare number or `:8899` both work. |
| `TASKTRACKER_DB` | see below | Path to the SQLite database file. Parent directories are created automatically. |
| `SESSION_SECURE` | auto | Force the `Secure` flag on session cookies. Normally unnecessary — see below. |

### Where your data lives

By default the database is created in your OS application-data directory, **never** inside the repository:

| OS | Path |
|---|---|
| Windows | `%AppData%\Roaming\tasktracker\data.db` |
| macOS | `~/Library/Application Support/tasktracker/data.db` |
| Linux | `$XDG_CONFIG_HOME/tasktracker/data.db` (or `~/.config/...`) |

The exact path is printed at startup. Override it with `TASKTRACKER_DB=/var/lib/tasktracker/data.db`.

Keeping the database outside the source tree is deliberate: task content is personal, and it should not be possible to commit it by accident. Back up by copying that one file — it holds everything.

### Cookies and HTTPS

Session cookies are marked `Secure` automatically when the request arrives over HTTPS, including via a reverse proxy that sets `X-Forwarded-Proto`. So localhost over plain HTTP works with no configuration, and a normal production deployment behind Caddy, nginx or Traefik also works with no configuration.

Set `SESSION_SECURE=true` only if you terminate TLS somewhere that does not set `X-Forwarded-Proto`.

## Deploying

The app speaks plain HTTP and has no built-in TLS. **Put it behind a reverse proxy that terminates HTTPS.** Sessions are cookie-based, so serving it over plain HTTP across a network exposes them.

A minimal Caddy config:

```
tasks.example.com {
    reverse_proxy 127.0.0.1:8899
}
```

Then run the binary with `PORT=8899` and a `TASKTRACKER_DB` on persistent storage. On container hosts with ephemeral disks (Fly.io, Render, Railway), point `TASKTRACKER_DB` at a mounted volume or the database is lost on redeploy.

Because SQLite is a single file owned by one process, run **one** instance — this app does not support multiple replicas behind a load balancer. For the workload it's built for, one instance is plenty.

## Upgrading from the single-user version

Earlier versions stored everything in a `tasks.json` file next to the binary. To bring that data into an account:

```bash
# 1. Start the app and sign up in the browser first.
# 2. Then, with the app stopped or running:
./tasktracker -import ./tasks.json -user you@example.com
```

It refuses to run if the account already has tasks, so it can't be applied twice by accident.

Categories and people used to live in browser `localStorage`. The import recreates categories from the tasks themselves; **people names sync automatically the first time you sign in from the browser you used before**, because their ids match. Sign in there once and assignments resolve on their own.

## API

Every endpoint returns JSON and is scoped to the signed-in account. All routes except signup and login require a valid session cookie and return `401` without one. A task belonging to another account returns `404`, not `403`, so ids can't be probed.

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/api/signup` | Create an account, returns a session cookie |
| `POST` | `/api/login` | Sign in |
| `POST` | `/api/logout` | Revoke the current session |
| `GET` | `/api/me` | Current account |
| `GET` `POST` | `/api/tasks` | List / create tasks |
| `PUT` `DELETE` | `/api/tasks/{id}` | Update / delete a task |
| `POST` | `/api/tasks/{id}/events` | Append to a task's event log |
| `GET` `PUT` | `/api/cats` | Read / replace categories |
| `GET` `PUT` | `/api/people` | Read / replace people |
| `POST` | `/api/reset` | Clear this account's board |

## Security notes

- Passwords are hashed with bcrypt (cost 12). Passwords over 72 bytes are rejected rather than silently truncated, since bcrypt ignores the remainder.
- Sessions are opaque 256-bit random tokens stored server-side, so signing out genuinely revokes access. They are not JWTs.
- Cookies are `HttpOnly` and `SameSite=Lax`.
- Login is rate-limited per IP, and a wrong password and an unknown email return the same message after the same work, so the endpoint doesn't reveal which accounts exist.
- There is no email verification or password reset yet. Anyone who can reach the server can create an account — put it behind a proxy with access control if that isn't what you want.

## Project layout

```
main.go          HTTP handlers and routing
auth.go          signup/login/sessions, rate limiting
store.go         all SQL, every query scoped by user_id
db.go            connection setup, pragmas, transactions
schema.sql       tables and indexes (embedded at build time)
config.go        environment configuration
import.go        one-shot legacy tasks.json import
static/index.html   the entire frontend
```

## License

MIT — see [LICENSE](LICENSE).
