package main

import (
	"time"
)

// Task is the core domain type.
type Task struct {
	ID        int
	Title     string
	Done      bool
	CreatedAt time.Time
}

// TaskRepo is a simple in-memory repository.
// When wired through a serial dispatch queue (see main.go) it is safe to use
// from any goroutine without explicit locking.
type TaskRepo struct {
	tasks  []Task
	nextID int
}

func NewTaskRepo() *TaskRepo {
	r := &TaskRepo{nextID: 1}
	// seed with a few tasks
	r.add("Buy groceries")
	r.add("Write goview docs")
	r.add("Build something cool")
	return r
}

func (r *TaskRepo) add(title string) Task {
	t := Task{ID: r.nextID, Title: title, CreatedAt: time.Now()}
	r.tasks = append(r.tasks, t)
	r.nextID++
	return t
}

func (r *TaskRepo) All() []Task {
	out := make([]Task, len(r.tasks))
	copy(out, r.tasks)
	return out
}

func (r *TaskRepo) Add(title string) Task {
	return r.add(title)
}

func (r *TaskRepo) Toggle(id int) {
	for i := range r.tasks {
		if r.tasks[i].ID == id {
			r.tasks[i].Done = !r.tasks[i].Done
			return
		}
	}
}

func (r *TaskRepo) Delete(id int) {
	for i, t := range r.tasks {
		if t.ID == id {
			r.tasks = append(r.tasks[:i], r.tasks[i+1:]...)
			return
		}
	}
}
