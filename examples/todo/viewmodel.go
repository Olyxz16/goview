package main

import (
	"strings"

	"github.com/Olyxz16/goview"
)

// TaskViewModel holds all application state as Observables.
// It has zero knowledge of the webview, HTML, or JS.
// It is fully unit-testable without a webview.
type TaskViewModel struct {
	goview.BaseVM

	// Observable state
	Tasks  *goview.Observable[[]Task]
	Filter *goview.Observable[string] // "all" | "active" | "done"
	Status *goview.Observable[string]

	// Computed — derived automatically, never set manually
	Visible *goview.Observable[[]Task]
	Count   *goview.Observable[int] // count of remaining (not done) tasks

	repo *TaskRepo
}

func NewTaskViewModel(repo *TaskRepo) *TaskViewModel {
	vm := &TaskViewModel{
		BaseVM: goview.NewBaseVM(),
		repo:   repo,
	}

	// Primary observables — all state lives here.
	vm.Tasks = goview.Observe(repo.All(), vm)
	vm.Filter = goview.Observe("all", vm)
	vm.Status = goview.Observe("", vm)

	// Visible is computed from both Tasks and Filter.
	vm.Visible = goview.Join(vm.Tasks, vm.Filter, func(tasks []Task, filter string) []Task {
		if filter == "all" {
			return tasks
		}
		out := make([]Task, 0, len(tasks))
		for _, t := range tasks {
			if filter == "done" && t.Done {
				out = append(out, t)
			} else if filter == "active" && !t.Done {
				out = append(out, t)
			}
		}
		return out
	})

	// Count of remaining tasks
	vm.Count = goview.Computed(vm.Tasks, func(tasks []Task) int {
		n := 0
		for _, t := range tasks {
			if !t.Done {
				n++
			}
		}
		return n
	})

	return vm
}

// ── Business methods ──────────────────────────────────────────────────────────
// These are exposed to JS via webview.Bind. They mutate state; Observables
// notify Components; Components update the DOM. No UI code here.

func (vm *TaskViewModel) AddTask(title string) {
	title = strings.TrimSpace(title)
	if title == "" {
		vm.Status.Set("Task title cannot be empty")
		return
	}
	vm.repo.Add(title)
	vm.Tasks.Set(vm.repo.All())
	vm.Status.Set("")
}

func (vm *TaskViewModel) ToggleTask(id int) {
	vm.repo.Toggle(id)
	vm.Tasks.Set(vm.repo.All())
}

func (vm *TaskViewModel) DeleteTask(id int) {
	vm.repo.Delete(id)
	vm.Tasks.Set(vm.repo.All())
}

func (vm *TaskViewModel) SetFilter(filter string) {
	vm.Filter.Set(filter)
}

func (vm *TaskViewModel) ClearDone() {
	tasks := vm.repo.All()
	for _, t := range tasks {
		if t.Done {
			vm.repo.Delete(t.ID)
		}
	}
	vm.Tasks.Set(vm.repo.All())
}
