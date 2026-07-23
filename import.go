package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"
)

// runImport loads a pre-database tasks.json into one account. It is a one-shot
// migration path for anyone upgrading from the single-user version, invoked as:
//
//	tasktracker -import ./tasks.json -user you@example.com
//
// People are deliberately not created here. In the old design their names lived
// only in the browser's localStorage, so the file carries bare ids like
// "p_1777358573109". The task_people rows are written with those ids and left
// dangling; the first browser sign-in uploads the real names against the same
// ids and the assignments resolve. The UI already skips unknown ids
// (static/index.html:1854), so nothing breaks in between.
func runImport(a *app, path, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return errors.New("-user is required: which account should these tasks belong to?")
	}

	var uid int64
	if err := a.db.QueryRow(`SELECT id FROM users WHERE email=?`, email).Scan(&uid); err != nil {
		return fmt.Errorf("no account found for %s -- sign up in the app first, then re-run this", email)
	}

	var existing int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE user_id=?`, uid).Scan(&existing); err != nil {
		return err
	}
	if existing > 0 {
		return fmt.Errorf("%s already has %d tasks; refusing to import on top of them", email, existing)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// The legacy file's JSON tags are identical to Task's, so it decodes directly.
	var legacy struct {
		Tasks []*Task `json:"tasks"`
	}
	if err := json.Unmarshal(raw, &legacy); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	if len(legacy.Tasks) == 0 {
		return fmt.Errorf("%s contains no tasks", path)
	}

	// Categories previously lived in localStorage, so the file references them
	// by id with no labels. Recreating them here matters: the board filters on
	// category (static/index.html:1334), so imported tasks whose category has
	// no row would be invisible.
	seen := map[string]bool{}
	cats := []Cat{}
	for _, t := range legacy.Tasks {
		if t.Category == "" || seen[t.Category] {
			continue
		}
		seen[t.Category] = true
		cats = append(cats, Cat{ID: t.Category, Label: labelize(t.Category)})
	}
	if err := replaceCats(a.db, uid, cats); err != nil {
		return fmt.Errorf("create categories: %w", err)
	}

	for _, t := range legacy.Tasks {
		if t.People == nil {
			t.People = []string{}
		}
		if t.Events == nil {
			t.Events = []Event{}
		}
		if err := createTask(a.db, uid, t); err != nil {
			return fmt.Errorf("import task %q: %w", t.Title, err)
		}
	}

	log.Printf("imported %d tasks and %d categories into %s", len(legacy.Tasks), len(cats), email)
	log.Printf("sign in from the browser you used before and your people will sync across")
	return nil
}

// labelize turns a category id such as "work" or "side_projects" into a
// display label: "Work", "Side projects".
func labelize(id string) string {
	s := strings.ReplaceAll(id, "_", " ")
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
