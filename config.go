package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Config holds every runtime knob. Everything is env-driven so the binary can
// be dropped onto a host without a config file.
type Config struct {
	Addr        string // listen address, e.g. ":8080"
	DBPath      string // absolute path to the SQLite file
	ForceSecure bool   // always mark session cookies Secure, even if the request looks plaintext
}

// defaultDBPath resolves to a per-user application data directory, never the
// working directory. Keeping the database out of the source tree means task
// content cannot be captured by a stray `git add`.
func defaultDBPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		// UserConfigDir only fails when the environment is missing HOME or
		// APPDATA. Fall back to a dotdir rather than dropping the DB in cwd.
		home, herr := os.UserHomeDir()
		if herr != nil {
			log.Fatalf("cannot determine a data directory: %v", err)
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "tasktracker", "data.db")
}

func loadConfig() Config {
	c := Config{
		Addr:   ":8080",
		DBPath: defaultDBPath(),
	}

	if p := os.Getenv("PORT"); p != "" {
		if !strings.HasPrefix(p, ":") {
			p = ":" + p
		}
		c.Addr = p
	}
	if d := os.Getenv("TASKTRACKER_DB"); d != "" {
		abs, err := filepath.Abs(d)
		if err != nil {
			log.Fatalf("invalid TASKTRACKER_DB %q: %v", d, err)
		}
		c.DBPath = abs
	}

	// Whether a session cookie gets the Secure flag is decided per request (see
	// requestIsTLS), so localhost-over-HTTP and production-behind-a-proxy both
	// work with no configuration. This forces it on regardless, for operators
	// who terminate TLS somewhere that does not set X-Forwarded-Proto.
	c.ForceSecure = isTruthy(os.Getenv("SESSION_SECURE"))

	return c
}

func isTruthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

// requestIsTLS reports whether the client reached us over HTTPS, accounting for
// the common case of TLS being terminated by a reverse proxy in front of us.
func requestIsTLS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	// X-Forwarded-Proto may be a comma-separated chain; the first entry is the
	// protocol the original client used.
	if xf := r.Header.Get("X-Forwarded-Proto"); xf != "" {
		first, _, _ := strings.Cut(xf, ",")
		return strings.EqualFold(strings.TrimSpace(first), "https")
	}
	return false
}
