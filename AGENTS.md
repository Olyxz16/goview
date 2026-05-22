# goview — Agent Reference

This document is for AI agents and automation tools that generate or modify code using the `github.com/Olyxz16/goview` library.

---

## What goview is

goview is a Go-native MVVM binding layer for desktop webview applications. The architecture is:

- **Go owns all state and logic.**
- **HTML is a static scaffold** with named container elements.
- **JS is a thin runtime** (~150 lines) that patches the DOM and routes user events back to Go bindings.
- **No JS bundler, no npm, no frontend framework.**

All rendering happens in Go. Components push HTML strings into the DOM via `Eval`. The JS runtime morphs the DOM with Idiomorph to preserve focus, input values, and scroll position.

---

## Mental model for agents

When building a goview app, think in this order:

1. **Domain / Repository** — plain Go structs and methods, zero UI code.
2. **ViewModel** — embed `goview.BaseVM`, declare `*goview.Observable[T]` fields for every piece of state the UI depends on.
3. **Renderer** — pure `func(T) string` functions that turn state into HTML.
4. **HTML scaffold** — static file with empty `<div id="...">` containers and `w-call` buttons.
5. **main.go** — wire driver → VM → components → bindings → `app.Run("index.html")`.

Never put UI logic in the ViewModel. Never put business logic in the renderer.

---

## Core API Reference

### Observable[T]

```go
obs := goview.Observe(initialValue, vm)       // dispatch-aware (recommended)
obs := goview.NewObservable(initialValue)       // synchronous listeners
obs := goview.NewObservableWithDispatch(v, q)   // explicit queue

obs.Set(newValue)   // updates value, notifies listeners
obs.Get() T         // reads current value
obs.OnChange(fn)    // register a listener
```

**Rule:** All UI-bound Observables must be created with `Observe(..., vm)` or `NewObservableWithDispatch` so listeners run on the serial queue.

### Component[T]

```go
comp := goview.NewComponent(selector, observable, renderer, eval)
comp.Mount()
```

- `selector` — CSS selector (e.g. `"#task-list"`).
- `observable` — the `*Observable[T]` this component renders.
- `renderer` — `func(T) string`. Use `html/template`, `text/template`, `templ`, or `fmt.Sprintf`.
- `eval` — the result of `vm.Eval(d.Eval)` (SafeEval wrapper).

`Mount` must be called once, after the DOM is ready. When using `goview.App`, components are mounted automatically via the `__goviewReady` callback.

### Computed / Join

```go
derived := goview.Computed(source, func(src S) T { ... })
joined  := goview.Join(src1, src2, func(a A, b B) C { ... })
```

These are read-only. Do not call `Set` on them. The dispatch queue is inherited from the source(s).

### BaseVM

```go
type MyVM struct {
    goview.BaseVM
    // ... Observable fields
}

vm := &MyVM{BaseVM: goview.NewBaseVM()}
```

Methods provided by embedding:

| Method | Returns | Usage |
|--------|---------|-------|
| `vm.Dispatch()` | `func(func())` | Pass to `NewApp`, `NewObservableWithDispatch` |
| `vm.Eval(rawEval)` | `func(string)` | Pass as `eval` arg to `NewComponent` |
| `vm.BindString(fn)` | `func(string)` | Wrap webview string callbacks |
| `vm.BindInt(fn)` | `func(int)` | Wrap webview int callbacks |
| `vm.BindVoid(fn)` | `func()` | Wrap webview void callbacks |

### App

```go
app := goview.NewApp(driver, eval, dispatch)
app.Mount(components...)
app.Bind("Name", fn)   // fn must be func(string), func(int), or func()
app.Run("index.html")
```

`app.Run` reads the HTML, injects the JS runtime before `</body>`, registers `__goviewReady`, and starts the webview.

### DOM

For lightweight operations that do not need a full re-render:

```go
dom := goview.NewDOM(eval)
dom.Show("#spinner")
dom.Hide("#spinner")
dom.AddClass("#btn", "active")
dom.RemoveClass("#btn", "active")
dom.Eval("customJS()")   // escape hatch
```

### SafeEval

```go
eval := goview.SafeEval(dispatch, rawEval)
```

Always use this wrapper (or `vm.Eval`) when passing an `Eval` function to `NewComponent` or `NewDOM`. It ensures JS runs on the webview's thread.

### RenderTemplate

```go
html, err := goview.RenderTemplate(tmpl, data)
```

Executes a `text/template` and returns the string.

---

## HTML Attribute Reference (goview.js)

goview.js uses event delegation. These attributes work on elements rendered statically **and** on elements injected dynamically by Components.

| Attribute | Meaning |
|-----------|---------|
| `w-call="BindingName"` | Calls the Go binding on click (or `w-trigger`). |
| `w-args="raw"` | Sends a single string argument. |
| `w-args-json='{"key":"val"}'` | Sends a JSON object as the argument. |
| `w-value` | Reads the element's own `.value` and sends it as the argument. |
| `w-value="#selector"` | Reads `.value` from another element and sends it. |
| `w-clear` | After calling the binding, clears `.value` on the element and refocuses it. |
| `w-clear="#selector"` | After calling, clears `.value` on the matched element. |
| `w-trigger="event"` | Overrides the default `click` trigger. Common: `input`, `change`, `keydown`. |
| `w-key="Enter"` | With `w-trigger="keydown"`, only fires when this key is pressed. |
| `w-debounce="300"` | Waits N ms after the last event before calling the binding. |

**Argument priority:** `w-args-json` > `w-args` > `w-value`. Only one argument array is sent.

**Binding signature:** The Go function bound via `app.Bind` must match the argument shape:

- `w-args="foo"` → `func(string)`
- `w-args-json='{"id":1}'` → `func(string)` (JSON string) or parse inside Go
- No args → `func()`

The current driver adapters pass the first argument as a string when `w-args` is used. For structured data, use `w-args-json` and unmarshal in Go if needed.

---

## Code Generation Rules

### 1. ViewModel structure

Always embed `goview.BaseVM`. Declare Observables as pointer fields.

```go
type TaskVM struct {
    goview.BaseVM
    Tasks  *goview.Observable[[]Task]
    Filter *goview.Observable[string]
}
```

Never store webview references, HTML strings, or JS inside the ViewModel.

### 2. Business methods

Business methods mutate Observables. They never touch the webview directly.

```go
func (vm *TaskVM) AddTask(title string) {
    title = strings.TrimSpace(title)
    if title == "" {
        vm.Status.Set("Title cannot be empty")
        return
    }
    vm.repo.Add(title)
    vm.Tasks.Set(vm.repo.All())   // triggers re-render
    vm.Status.Set("")
}
```

### 3. Renderer functions

Renderers are pure `func(T) string`. They must not have side effects.

```go
func renderTasks(tasks []Task) string {
    var buf bytes.Buffer
    taskTmpl.Execute(&buf, tasks)
    return buf.String()
}
```

For simple cases, `fmt.Sprintf` is fine. For complex markup, prefer `html/template`.

### 4. main.go wiring

Follow this exact sequence:

```go
d := driver.New()
defer d.Destroy()
d.SetTitle("...")
d.SetSize(w, h)

vm := NewViewModel(...)
eval := vm.Eval(d.Eval)
app := goview.NewApp(d, eval, vm.Dispatch())

app.Mount(
    goview.NewComponent("#selector1", vm.Obs1, render1, eval),
    goview.NewComponent("#selector2", vm.Obs2, render2, eval),
)

app.Bind("MethodName", vm.BindVoid(vm.MethodName))
app.Run("index.html")
```

Do not call `comp.Mount()` manually when using `goview.App`.

### 5. HTML scaffold

Use empty container elements. Do not inline state or logic.

```html
<div id="task-list"></div>
<div id="status-bar"></div>
<button w-call="AddTask" w-args="hello">Add</button>
```

If the project uses `goview.App`, do **not** manually import `goview.js` or `idiomorph.min.js`. `app.Run` injects them automatically before `</body>`.

### 6. Driver packages

The `driver/` package should contain:

- `driver.go` — the `Driver` interface
- `webview.go` (build tag `!linux`) — adapter for `github.com/abemedia/go-webview`
- `webui.go` (build tag `linux`) — adapter for `github.com/webui-dev/go-webui/v2`

Do not put webview imports directly in `main.go`. Always go through the driver abstraction.

---

## Important Invariants

1. **Serial queue invariant:** All state mutations and all renders must run on the same goroutine. `BaseVM` provides this queue. `Bind*` wrappers ensure callbacks from the webview enter the queue. Never call `obs.Set` from a background goroutine without dispatching it.

2. **Renderer purity:** Renderers must be deterministic functions of their input. They must not read global state, write files, or call `Eval`.

3. **One eval wrapper:** Create `eval := vm.Eval(d.Eval)` once and reuse it for every Component and DOM helper.

4. **No manual JS imports:** `app.Run` embeds the JS runtime. Do not add `<script src="goview.js">` to the HTML.

5. **Selector uniqueness:** A CSS selector should only be targeted by one Component. If two Components write to the same selector, they will overwrite each other.

6. **Template errors must not crash:** `goview.RenderTemplate` renders errors as visible HTML. When writing your own renderer, return an error fragment string rather than panicking.

---

## Testing strategy for agents

When generating tests:

- Use `goview.NewBaseVMWithDispatch(func(f func()) { f() })` for a synchronous queue.
- Test ViewModel methods by asserting `obs.Get()` values after mutation.
- Do not instantiate a real webview in unit tests.
- Renderers can be tested independently by calling them with sample data and asserting the returned HTML string contains expected substrings.

---

## What NOT to do

| Bad pattern | Why | Fix |
|-------------|-----|-----|
| Calling `d.Eval` directly from a ViewModel | Breaks serial queue | Use `vm.Eval(d.Eval)` wrapper |
| Passing a bare `func(string)` to `d.Bind` | Runs on webview thread, races | Use `vm.BindString(...)` wrapper |
| Using `w-call` without `app.Bind` | JS error: binding not found | Register every `w-call` target in `main.go` |
| Multiple Components on one selector | Race / flicker | Merge into a single Component or use nested selectors |
| Blocking the dispatch queue | UI freezes | Offload heavy work to a goroutine, then `dispatch` the result |
| Mutating an Observable from a background goroutine | Data race | Send the mutation through the dispatch queue |

---

## Common task templates

### Add a new UI section

1. Add an `Observable[T]` field to the ViewModel.
2. Initialise it in the constructor with `goview.Observe(initial, vm)`.
3. Add a renderer `func(T) string`.
4. Add a container `<div id="new-section"></div>` to `index.html`.
5. Add `goview.NewComponent("#new-section", vm.NewField, renderNew, eval)` to `app.Mount(...)`.
6. If user interaction is needed, add a `w-call` button and bind it in `main.go`.

### Add a button that mutates state

1. Write a ViewModel method that mutates the relevant Observable.
2. Add `<button w-call="MethodName" w-args="...">` to the HTML (inside a Component template if dynamic).
3. In `main.go`: `app.Bind("MethodName", vm.BindVoid(vm.MethodName))` (or `BindString` / `BindInt`).

### Add a text input that submits on Enter

```html
<input type="text"
       w-call="AddItem"
       w-trigger="keydown"
       w-key="Enter"
       w-value
       w-clear>
```

1. ViewModel: `func (vm *VM) AddItem(title string) { ... }`
2. main.go: `app.Bind("AddItem", vm.BindString(vm.AddItem))`

### Add a filtered list

1. Observable for raw data: `vm.Items = goview.Observe(repo.All(), vm)`
2. Observable for filter: `vm.Filter = goview.Observe("all", vm)`
3. Computed visible list: `vm.Visible = goview.Join(vm.Items, vm.Filter, filterFn)`
4. Component renders `vm.Visible`.
5. Filter buttons: `<button w-call="SetFilter" w-args="active">Active</button>` bound to `vm.BindString(vm.SetFilter)`.

---

## File checklist for a complete goview app

```
project/
├── go.mod
├── main.go              # driver, VM, App wiring
├── viewmodel.go         # VM struct + methods
├── renderer.go          # templates + render funcs
├── repo.go              # domain / persistence (optional)
├── index.html           # static HTML scaffold
└── driver/
    ├── driver.go        # Driver interface
    ├── webview.go       //go:build !linux
    └── webui.go         //go:build linux
```

When asked to "build a goview app", generate these files. Do not generate `vendor/` or JS files — the runtime is embedded by `goview`.
