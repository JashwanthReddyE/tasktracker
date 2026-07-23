package main

import (
	"database/sql"
	"time"
)

// The JSON shapes below are the existing wire contract with static/index.html.
// Field names and tags must not drift without a matching frontend change.

type Event struct {
	Type string    `json:"type"`
	Time time.Time `json:"time"`
	Text string    `json:"text"`
}

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Notes     string    `json:"notes,omitempty"`
	Priority  string    `json:"priority"`
	Row       string    `json:"row"`
	Category  string    `json:"category"`
	DueDate   string    `json:"dueDate,omitempty"`
	People    []string  `json:"people"`
	Events    []Event   `json:"events"`
	CreatedAt time.Time `json:"createdAt"`
	Archived  bool      `json:"archived,omitempty"`
	Order     int       `json:"order"`
}

type Person struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Hue  int    `json:"hue"`
}

type Cat struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

const timeFmt = time.RFC3339Nano

func parseTime(s string) time.Time {
	t, err := time.Parse(timeFmt, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// ── Tasks ─────────────────────────────────────────────

// listTasks returns every task for one user, with assignees and events
// attached. It issues three queries rather than one-per-task: the join tables
// are fetched wholesale and stitched in memory, so cost stays flat as the
// board grows.
func listTasks(db *sql.DB, userID int64) ([]*Task, error) {
	rows, err := db.Query(`
		SELECT id, title, notes, priority, status, category_id,
		       due_date, created_at, archived, position
		  FROM tasks
		 WHERE user_id = ?
		 ORDER BY position, id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []*Task{}
	byID := map[int]*Task{}
	for rows.Next() {
		t := &Task{People: []string{}, Events: []Event{}}
		var created string
		if err := rows.Scan(&t.ID, &t.Title, &t.Notes, &t.Priority, &t.Row,
			&t.Category, &t.DueDate, &created, &t.Archived, &t.Order); err != nil {
			return nil, err
		}
		t.CreatedAt = parseTime(created)
		tasks = append(tasks, t)
		byID[t.ID] = t
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return tasks, nil
	}

	// Assignees. The join through tasks is what enforces ownership.
	pr, err := db.Query(`
		SELECT tp.task_id, tp.person_id
		  FROM task_people tp
		  JOIN tasks t ON t.id = tp.task_id
		 WHERE t.user_id = ?
		 ORDER BY tp.position, tp.rowid`, userID)
	if err != nil {
		return nil, err
	}
	defer pr.Close()
	for pr.Next() {
		var taskID int
		var personID string
		if err := pr.Scan(&taskID, &personID); err != nil {
			return nil, err
		}
		if t := byID[taskID]; t != nil {
			t.People = append(t.People, personID)
		}
	}
	if err := pr.Err(); err != nil {
		return nil, err
	}

	er, err := db.Query(`
		SELECT e.task_id, e.type, e.time, e.text
		  FROM events e
		  JOIN tasks t ON t.id = e.task_id
		 WHERE t.user_id = ?
		 ORDER BY e.id`, userID)
	if err != nil {
		return nil, err
	}
	defer er.Close()
	for er.Next() {
		var taskID int
		var ev Event
		var ts string
		if err := er.Scan(&taskID, &ev.Type, &ts, &ev.Text); err != nil {
			return nil, err
		}
		ev.Time = parseTime(ts)
		if t := byID[taskID]; t != nil {
			t.Events = append(t.Events, ev)
		}
	}
	return tasks, er.Err()
}

func createTask(db *sql.DB, userID int64, t *Task) error {
	return tx(db, func(tr *sql.Tx) error {
		res, err := tr.Exec(`
			INSERT INTO tasks (user_id, title, notes, priority, status,
			                   category_id, due_date, created_at, archived, position)
			VALUES (?,?,?,?,?,?,?,?,?,?)`,
			userID, t.Title, t.Notes, t.Priority, t.Row, t.Category,
			t.DueDate, t.CreatedAt.Format(timeFmt), t.Archived, t.Order)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		t.ID = int(id)
		if err := writePeople(tr, t.ID, t.People); err != nil {
			return err
		}
		return writeEvents(tr, t.ID, t.Events)
	})
}

// updateTask replaces a task in place. Nil People or Events mean "leave as is",
// matching the previous JSON-file behaviour where an omitted field preserved
// the stored value rather than clearing it. Returns false if the task does not
// exist or belongs to someone else -- callers surface that as 404.
func updateTask(db *sql.DB, userID int64, id int, t *Task) (bool, error) {
	found := false
	err := tx(db, func(tr *sql.Tx) error {
		res, err := tr.Exec(`
			UPDATE tasks
			   SET title=?, notes=?, priority=?, status=?, category_id=?,
			       due_date=?, archived=?, position=?
			 WHERE id=? AND user_id=?`,
			t.Title, t.Notes, t.Priority, t.Row, t.Category,
			t.DueDate, t.Archived, t.Order, id, userID)
		if err != nil {
			return err
		}
		n, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			return nil // not ours; leave found false
		}
		found = true

		if t.People != nil {
			if _, err := tr.Exec(`DELETE FROM task_people WHERE task_id=?`, id); err != nil {
				return err
			}
			if err := writePeople(tr, id, t.People); err != nil {
				return err
			}
		}
		if t.Events != nil {
			if _, err := tr.Exec(`DELETE FROM events WHERE task_id=?`, id); err != nil {
				return err
			}
			if err := writeEvents(tr, id, t.Events); err != nil {
				return err
			}
		}
		return nil
	})
	return found, err
}

func deleteTask(db *sql.DB, userID int64, id int) (bool, error) {
	res, err := db.Exec(`DELETE FROM tasks WHERE id=? AND user_id=?`, id, userID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// addEvent appends to a task's log. The ownership check is folded into the
// INSERT's SELECT so there is no read-then-write gap.
func addEvent(db *sql.DB, userID int64, taskID int, ev Event) (bool, error) {
	res, err := db.Exec(`
		INSERT INTO events (task_id, type, time, text)
		SELECT id, ?, ?, ? FROM tasks WHERE id=? AND user_id=?`,
		ev.Type, ev.Time.Format(timeFmt), ev.Text, taskID, userID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func writePeople(tr *sql.Tx, taskID int, ids []string) error {
	for i, pid := range ids {
		if _, err := tr.Exec(`
			INSERT OR IGNORE INTO task_people (task_id, person_id, position)
			VALUES (?,?,?)`, taskID, pid, i); err != nil {
			return err
		}
	}
	return nil
}

func writeEvents(tr *sql.Tx, taskID int, evs []Event) error {
	for _, e := range evs {
		if _, err := tr.Exec(`
			INSERT INTO events (task_id, type, time, text) VALUES (?,?,?,?)`,
			taskID, e.Type, e.Time.Format(timeFmt), e.Text); err != nil {
			return err
		}
	}
	return nil
}

// ── Categories ────────────────────────────────────────

func listCats(db *sql.DB, userID int64) ([]Cat, error) {
	rows, err := db.Query(`
		SELECT id, label FROM categories
		 WHERE user_id=? ORDER BY position, rowid`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cats := []Cat{}
	for rows.Next() {
		var c Cat
		if err := rows.Scan(&c.ID, &c.Label); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

// replaceCats swaps the whole list atomically. Whole-collection replace mirrors
// how the client mutates this state -- it re-serializes the full array on every
// add, rename, reorder and delete.
func replaceCats(db *sql.DB, userID int64, cats []Cat) error {
	return tx(db, func(tr *sql.Tx) error {
		if _, err := tr.Exec(`DELETE FROM categories WHERE user_id=?`, userID); err != nil {
			return err
		}
		for i, c := range cats {
			if _, err := tr.Exec(`
				INSERT INTO categories (user_id, id, label, position) VALUES (?,?,?,?)`,
				userID, c.ID, c.Label, i); err != nil {
				return err
			}
		}
		return nil
	})
}

// ── People ────────────────────────────────────────────

// listPeople returns people keyed by category id, matching the client's
// allPeople[categoryId] structure.
func listPeople(db *sql.DB, userID int64) (map[string][]Person, error) {
	rows, err := db.Query(`
		SELECT category_id, id, name, hue FROM people
		 WHERE user_id=? ORDER BY position, rowid`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string][]Person{}
	for rows.Next() {
		var cat string
		var p Person
		if err := rows.Scan(&cat, &p.ID, &p.Name, &p.Hue); err != nil {
			return nil, err
		}
		out[cat] = append(out[cat], p)
	}
	return out, rows.Err()
}

func replacePeople(db *sql.DB, userID int64, people map[string][]Person) error {
	return tx(db, func(tr *sql.Tx) error {
		if _, err := tr.Exec(`DELETE FROM people WHERE user_id=?`, userID); err != nil {
			return err
		}
		for cat, list := range people {
			for i, p := range list {
				if _, err := tr.Exec(`
					INSERT OR REPLACE INTO people
						(user_id, id, category_id, name, hue, position)
					VALUES (?,?,?,?,?,?)`,
					userID, p.ID, cat, p.Name, p.Hue, i); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// ── Reset ─────────────────────────────────────────────

// resetUser clears one user's board. Tasks cascade to task_people and events
// via the foreign keys, which is why foreign_keys(ON) is set in the DSN.
func resetUser(db *sql.DB, userID int64) error {
	return tx(db, func(tr *sql.Tx) error {
		for _, q := range []string{
			`DELETE FROM tasks WHERE user_id=?`,
			`DELETE FROM people WHERE user_id=?`,
			`DELETE FROM categories WHERE user_id=?`,
		} {
			if _, err := tr.Exec(q, userID); err != nil {
				return err
			}
		}
		return nil
	})
}
