package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Event struct {
	Type string    `json:"type"`
	Time time.Time `json:"time"`
	Text string    `json:"text"`
}

type Task struct {
	ID        int      `json:"id"`
	Title     string   `json:"title"`
	Notes     string   `json:"notes,omitempty"`
	Priority  string   `json:"priority"`
	Row       string   `json:"row"`
	Category  string   `json:"category"`
	DueDate   string   `json:"dueDate,omitempty"`
	People    []string `json:"people"`
	Events    []Event  `json:"events"`
	CreatedAt time.Time `json:"createdAt"`
	Archived  bool     `json:"archived,omitempty"`
	Order     int      `json:"order"`
}

type Store struct {
	mu     sync.RWMutex
	Tasks  []*Task `json:"tasks"`
	NextID int     `json:"nextId"`
}

const dataFile = "tasks.json"

var db = &Store{NextID: 1}




func (s *Store) seed() {
	s.Tasks  = []*Task{}
	s.NextID = 1
}

func (s *Store) load() {
	data, err := os.ReadFile(dataFile)
	if err != nil {
		s.seed()
		s.save()
		return
	}
	if err := json.Unmarshal(data, s); err != nil {
		s.seed()
		s.save()
	}
}

func (s *Store) save() {
	s.mu.RLock()
	data, _ := json.MarshalIndent(s, "", "  ")
	s.mu.RUnlock()
	_ = os.WriteFile(dataFile, data, 0644)
}

func jsonHeader(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		h(w, r)
	}
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		db.mu.RLock()
		out := make([]*Task, len(db.Tasks))
		copy(out, db.Tasks)
		db.mu.RUnlock()
		if out == nil {
			out = []*Task{}
		}
		json.NewEncoder(w).Encode(out)

	case http.MethodPost:
		var t Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		t.CreatedAt = time.Now()
		t.Events = []Event{{Type: "created", Time: t.CreatedAt, Text: "Task created"}}
		if t.People == nil {
			t.People = []string{}
		}
		db.mu.Lock()
		t.ID = db.NextID
		db.NextID++
		db.Tasks = append(db.Tasks, &t)
		db.mu.Unlock()
		go db.save()
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(t)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")

	addEvent := strings.HasSuffix(path, "/events")
	if addEvent {
		path = strings.TrimSuffix(path, "/events")
	}

	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if addEvent {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var evt Event
		if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		evt.Time = time.Now()
		db.mu.Lock()
		for _, t := range db.Tasks {
			if t.ID == id {
				t.Events = append(t.Events, evt)
				break
			}
		}
		db.mu.Unlock()
		go db.save()
		json.NewEncoder(w).Encode(evt)
		return
	}

	switch r.Method {
	case http.MethodPut:
		var update Task
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		db.mu.Lock()
		for i, t := range db.Tasks {
			if t.ID == id {
				update.ID = id
				update.CreatedAt = t.CreatedAt
				if update.Events == nil {
					update.Events = t.Events
				}
				if update.People == nil {
					update.People = t.People
				}
				db.Tasks[i] = &update
				break
			}
		}
		db.mu.Unlock()
		go db.save()
		w.WriteHeader(http.StatusOK)

	case http.MethodDelete:
		db.mu.Lock()
		for i, t := range db.Tasks {
			if t.ID == id {
				db.Tasks = append(db.Tasks[:i], db.Tasks[i+1:]...)
				break
			}
		}
		db.mu.Unlock()
		go db.save()
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	db.load()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/tasks", jsonHeader(tasksHandler))
	mux.HandleFunc("/api/tasks/", jsonHeader(taskHandler))
	mux.HandleFunc("/api/reset", jsonHeader(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		db.mu.Lock()
		db.Tasks  = []*Task{}
		db.NextID = 1
		db.mu.Unlock()
		go db.save()
		w.WriteHeader(http.StatusNoContent)
	}))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	addr := ":8080"
	log.Printf("Task Tracker → http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
