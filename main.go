package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type app struct {
	db           *sql.DB
	cfg          Config
	loginLimiter *limiter
}

func main() {
	importPath := flag.String("import", "", "path to a legacy tasks.json to import, then exit")
	importUser := flag.String("user", "", "email of the account to import into (with -import)")
	flag.Parse()

	cfg := loadConfig()
	db, err := openDB(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	a := &app{
		db:  db,
		cfg: cfg,
		// Ten attempts per IP per fifteen minutes: generous for a fumbled
		// password, useless for guessing one.
		loginLimiter: newLimiter(10, 15*time.Minute),
	}

	if *importPath != "" {
		if err := runImport(a, *importPath, *importUser); err != nil {
			log.Fatalf("import failed: %v", err)
		}
		return
	}

	go a.sessionJanitor()

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           a.routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Close the database cleanly on Ctrl-C so SQLite can checkpoint the WAL.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		log.Println("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	log.Printf("Task Tracker → http://127.0.0.1%s", cfg.Addr)
	log.Printf("database: %s", cfg.DBPath)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	log.Println("stopped")
}

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()

	// Public: the only two endpoints reachable without a session.
	mux.HandleFunc("/api/signup", jsonHeader(a.handleSignup))
	mux.HandleFunc("/api/login", jsonHeader(a.handleLogin))
	mux.HandleFunc("/api/logout", jsonHeader(a.handleLogout))

	// Everything below is per-user and requires a valid session.
	mux.HandleFunc("/api/me", jsonHeader(a.requireAuth(a.handleMe)))
	mux.HandleFunc("/api/tasks", jsonHeader(a.requireAuth(a.tasksHandler)))
	mux.HandleFunc("/api/tasks/", jsonHeader(a.requireAuth(a.taskHandler)))
	mux.HandleFunc("/api/cats", jsonHeader(a.requireAuth(a.catsHandler)))
	mux.HandleFunc("/api/people", jsonHeader(a.requireAuth(a.peopleHandler)))
	mux.HandleFunc("/api/prefs", jsonHeader(a.requireAuth(a.prefsHandler)))
	mux.HandleFunc("/api/reset", jsonHeader(a.requireAuth(a.resetHandler)))

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	return mux
}

// sessionJanitor drops expired session rows. Without it the table grows without
// bound, since lookupSession only deletes the one row it happens to touch.
func (a *app) sessionJanitor() {
	for {
		a.db.Exec(`DELETE FROM sessions WHERE expires_at < ?`, time.Now().Format(timeFmt))
		time.Sleep(time.Hour)
	}
}

// ── HTTP helpers ──────────────────────────────────────

func jsonHeader(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		h(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ── Tasks ─────────────────────────────────────────────

func (a *app) tasksHandler(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)

	switch r.Method {
	case http.MethodGet:
		tasks, err := listTasks(a.db, uid)
		if err != nil {
			log.Printf("list tasks: %v", err)
			writeErr(w, http.StatusInternalServerError, "could not load tasks")
			return
		}
		writeJSON(w, http.StatusOK, tasks)

	case http.MethodPost:
		var t Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if strings.TrimSpace(t.Title) == "" {
			writeErr(w, http.StatusBadRequest, "title is required")
			return
		}
		t.CreatedAt = time.Now()
		t.Events = []Event{{Type: "created", Time: t.CreatedAt, Text: "Task created"}}
		if t.People == nil {
			t.People = []string{}
		}
		if err := createTask(a.db, uid, &t); err != nil {
			log.Printf("create task: %v", err)
			writeErr(w, http.StatusInternalServerError, "could not create task")
			return
		}
		writeJSON(w, http.StatusCreated, t)

	default:
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *app) taskHandler(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")

	addingEvent := strings.HasSuffix(path, "/events")
	if addingEvent {
		path = strings.TrimSuffix(path, "/events")
	}

	id, err := strconv.Atoi(path)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}

	if addingEvent {
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var evt Event
		if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid request body")
			return
		}
		evt.Time = time.Now()
		ok, err := addEvent(a.db, uid, id, evt)
		if err != nil {
			log.Printf("add event: %v", err)
			writeErr(w, http.StatusInternalServerError, "could not add event")
			return
		}
		if !ok {
			// A task owned by someone else is reported as missing, not
			// forbidden, so ids cannot be probed for existence.
			writeErr(w, http.StatusNotFound, "task not found")
			return
		}
		writeJSON(w, http.StatusOK, evt)
		return
	}

	switch r.Method {
	case http.MethodPut:
		var update Task
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid request body")
			return
		}
		ok, err := updateTask(a.db, uid, id, &update)
		if err != nil {
			log.Printf("update task: %v", err)
			writeErr(w, http.StatusInternalServerError, "could not update task")
			return
		}
		if !ok {
			writeErr(w, http.StatusNotFound, "task not found")
			return
		}
		w.WriteHeader(http.StatusOK)

	case http.MethodDelete:
		ok, err := deleteTask(a.db, uid, id)
		if err != nil {
			log.Printf("delete task: %v", err)
			writeErr(w, http.StatusInternalServerError, "could not delete task")
			return
		}
		if !ok {
			writeErr(w, http.StatusNotFound, "task not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// ── Categories & people ───────────────────────────────

func (a *app) catsHandler(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)

	switch r.Method {
	case http.MethodGet:
		cats, err := listCats(a.db, uid)
		if err != nil {
			log.Printf("list cats: %v", err)
			writeErr(w, http.StatusInternalServerError, "could not load categories")
			return
		}
		writeJSON(w, http.StatusOK, cats)

	case http.MethodPut:
		var cats []Cat
		if err := json.NewDecoder(r.Body).Decode(&cats); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := replaceCats(a.db, uid, cats); err != nil {
			log.Printf("replace cats: %v", err)
			writeErr(w, http.StatusInternalServerError, "could not save categories")
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *app) peopleHandler(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)

	switch r.Method {
	case http.MethodGet:
		people, err := listPeople(a.db, uid)
		if err != nil {
			log.Printf("list people: %v", err)
			writeErr(w, http.StatusInternalServerError, "could not load people")
			return
		}
		writeJSON(w, http.StatusOK, people)

	case http.MethodPut:
		var people map[string][]Person
		if err := json.NewDecoder(r.Body).Decode(&people); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := replacePeople(a.db, uid, people); err != nil {
			log.Printf("replace people: %v", err)
			writeErr(w, http.StatusInternalServerError, "could not save people")
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// prefsHandler reads and writes the account's UI-state blob (currently the last
// active category tab). The body is stored verbatim as JSON after being decoded
// once to reject anything that is not a JSON object.
func (a *app) prefsHandler(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)

	switch r.Method {
	case http.MethodGet:
		p, err := getPrefs(a.db, uid)
		if err != nil {
			log.Printf("get prefs: %v", err)
			writeErr(w, http.StatusInternalServerError, "could not load preferences")
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(p))

	case http.MethodPut:
		var obj map[string]any
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&obj); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid request body")
			return
		}
		raw, err := json.Marshal(obj)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid preferences")
			return
		}
		if err := setPrefs(a.db, uid, string(raw)); err != nil {
			log.Printf("set prefs: %v", err)
			writeErr(w, http.StatusInternalServerError, "could not save preferences")
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *app) resetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := resetUser(a.db, userID(r)); err != nil {
		log.Printf("reset: %v", err)
		writeErr(w, http.StatusInternalServerError, "could not reset board")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
