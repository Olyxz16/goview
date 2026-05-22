package main

import (
	"bytes"
	"fmt"
	"html/template"
)

// ── Templates ─────────────────────────────────────────────────────────────────

var taskListTmpl = template.Must(template.New("task-list").Parse(`
{{range .}}
<li class="task-item{{if .Done}} done{{end}}" data-id="{{.ID}}">
  <button class="toggle-btn"
          w-call="ToggleTask"
          w-args="{{.ID}}"
          aria-label="Toggle task">
    <span class="check">{{if .Done}}✓{{end}}</span>
  </button>
  <span class="task-title">{{.Title}}</span>
  <button class="delete-btn"
          w-call="DeleteTask"
          w-args="{{.ID}}"
          aria-label="Delete task">×</button>
</li>
{{else}}
<li class="empty-state">No tasks here.</li>
{{end}}
`))

var statusTmpl = template.Must(template.New("status").Parse(`
{{if .}}
<span class="status-error">{{.}}</span>
{{end}}
`))

var countTmpl = template.Must(template.New("count").Parse(`
<span>{{.}} remaining</span>
`))

var filterTmpl = template.Must(template.New("filters").Parse(`
<div class="filters">
  <button class="filter-btn{{if eq . "all"}} active{{end}}"
          w-call="SetFilter" w-args="all">All</button>
  <button class="filter-btn{{if eq . "active"}} active{{end}}"
          w-call="SetFilter" w-args="active">Active</button>
  <button class="filter-btn{{if eq . "done"}} active{{end}}"
          w-call="SetFilter" w-args="done">Done</button>
</div>
`))

// ── Renderer functions ────────────────────────────────────────────────────────
// These are plain functions with the signature func(T) string.
// They are passed to goview.NewComponent — the library has no opinion on how
// you produce HTML.

func renderTaskList(tasks []Task) string {
	var buf bytes.Buffer
	if err := taskListTmpl.Execute(&buf, tasks); err != nil {
		return fmt.Sprintf(`<li class="render-error">%s</li>`, err.Error())
	}
	return buf.String()
}

func renderStatus(msg string) string {
	var buf bytes.Buffer
	statusTmpl.Execute(&buf, msg)
	return buf.String()
}

func renderCount(n int) string {
	var buf bytes.Buffer
	countTmpl.Execute(&buf, n)
	return buf.String()
}

func renderFilters(filter string) string {
	var buf bytes.Buffer
	if err := filterTmpl.Execute(&buf, filter); err != nil {
		return fmt.Sprintf(`<div class="render-error">%s</div>`, err.Error())
	}
	return buf.String()
}
