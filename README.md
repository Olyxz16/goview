# goview

Go-native MVVM for desktop webview applications.

**goview** lets you build desktop GUI apps in pure Go. You write ViewModels with reactive state, plain HTML templates for the view, and a ~150-line JS runtime handles the DOM. No React, no Svelte, no build steps — just Go structs and `html/template`.

```go
vm.Count = goview.Observe(0, vm)
comp := goview.NewComponent("#counter", vm.Count, func(n int) string {
    return fmt.Sprintf(`<span>%d</span>`, n)
}, vm.Eval(d.Eval))
comp.Mount()
```

---

## Philosophy

Most desktop-app frameworks force you to split your brain: Go for the backend, JavaScript for the UI. goview says the Go side can own everything. The HTML is a static scaffold. The JS runtime is a dumb switchboard that routes clicks back to Go and patches the DOM with Idiomorph. All state, all logic, all rendering lives in Go.

## Features

- **Observable[T]** — generic reactive values with serial dispatch
- **Component[T]** — binds an Observable to a DOM element via CSS selector; auto re-renders on change
- **Computed / Join** — derived Observables from one or two sources
- **BaseVM** — embeddable ViewModel base with action queue, `Bind*` helpers, and `SafeEval`
- **App** — orchestrates component registration, binding setup, runtime injection, and auto-mount
- **DOM** — direct Go→DOM helpers when full re-renders are overkill
- **Idiomorph** — surgical DOM diffing preserves focus, input values, and scroll position
- **Zero build tools** — no npm, no bundler, no Vite. The JS runtime is vendored and embedded.
- **Cross-platform drivers** — works with [go-webview](https://github.com/abemedia/go-webview) and [go-webui](https://github.com/webui-dev/go-webui)

## Installation

```bash
go get github.com/Olyxz16/goview
```

## Quick Start

### 1. Write your HTML scaffold

`index.html` contains empty containers. No logic, no state.

```html
<!DOCTYPE html>
<html>
<head><title>Counter</title></head>
<body>
  <div id="counter"></div>
  <button w-call="Increment">+1</button>
</body>
</html>
```

### 2. Write a ViewModel

```go
type CounterVM struct {
    goview.BaseVM
    Count *goview.Observable[int]
}

func NewCounterVM() *CounterVM {
    vm := &CounterVM{BaseVM: goview.NewBaseVM()}
    vm.Count = goview.Observe(0, vm)
    return vm
}

func (vm *CounterVM) Increment() {
    vm.Count.Set(vm.Count.Get() + 1)
}
```

### 3. Wire it up in `main.go`

```go
func main() {
    vm := NewCounterVM()

    d := driver.New() // see Platform Drivers below
    defer d.Destroy()
    d.SetTitle("Counter")
    d.SetSize(400, 300)

    eval := vm.Eval(d.Eval)
    app := goview.NewApp(d, eval, vm.Dispatch())

    app.Mount(goview.NewComponent("#counter", vm.Count,
        func(n int) string {
            return fmt.Sprintf("<span>%d</span>", n)
        }, eval))

    app.Bind("Increment", vm.BindVoid(vm.Increment))
    app.Run("index.html")
}
```

`app.Run` injects the goview JS runtime, registers the mount callback, and starts the webview.

## Core Concepts

### Observable

An `Observable[T]` is a thread-safe reactive box. When you call `Set`, all listeners are notified.

```go
name := goview.Observe("Ada", vm) // dispatch-aware
name.Set("Grace")
name.Get() // "Grace"
```

If created via `Observe` (which uses a `DispatchProvider` like `BaseVM`), listeners run on a serial queue. This means your ViewModels and repositories need **no mutexes**.

### Component

A `Component[T]` wires an `Observable[T]` to a DOM selector. On every change it re-renders and patches the DOM with Idiomorph.

```go
comp := goview.NewComponent("#users", vm.Users, renderUsers, eval)
comp.Mount()
```

`renderUsers` is any `func([]User) string`. Use `html/template`, `text/template`, `templ`, or plain `fmt.Sprintf`.

### Computed and Join

Derive read-only Observables from sources:

```go
vm.Visible = goview.Join(vm.Tasks, vm.Filter, func(tasks []Task, filter string) []Task {
    // ...
})

vm.Count = goview.Computed(vm.Tasks, func(tasks []Task) int {
    // ...
})
```

The result inherits the dispatch queue from its source, so updates are still serialised.

### BaseVM

Embed `BaseVM` into your ViewModel to get:

- `Dispatch()` — the serial action queue
- `Eval(eval) func(string)` — SafeEval wrapper
- `BindString(fn) func(string)` — queue-wrapped callback
- `BindInt(fn) func(int)` — queue-wrapped callback
- `BindVoid(fn) func()` — queue-wrapped callback

This removes all boilerplate around goroutine safety.

### App

`App` ties everything together. You register components and bindings, then call `Run`.

```go
app := goview.NewApp(driver, eval, dispatch)
app.Mount(comp1, comp2, comp3)
app.Bind("AddTask", vm.BindString(vm.AddTask))
app.Run("index.html")
```

`Run` reads your HTML, injects the vendored JS runtime before `</body>`, and starts the webview.

## HTML Attributes

goview.js recognizes these declarative attributes on any element:

| Attribute | Description |
|-----------|-------------|
| `w-call="Name"` | Calls the Go binding named `Name` on click (or the trigger event). |
| `w-args="raw"` | Passes a single string argument to the binding. |
| `w-args-json='{"k":"v"}'` | Passes a JSON object as a single argument. |
| `w-value` | Reads the element's own `.value` as the argument. |
| `w-value="#id"` | Reads `.value` from another element as the argument. |
| `w-clear` | Clears the element's `.value` after the call, then refocuses. |
| `w-clear="#id"` | Clears another element's `.value` after the call. |
| `w-trigger="event"` | Changes the trigger from `click` (e.g. `input`, `change`, `keydown`). |
| `w-key="Enter"` | When `w-trigger="keydown"`, only fires on this key. |
| `w-debounce="300"` | Debounces the call by N milliseconds. |

Because goview.js uses event delegation, dynamically rendered content (inside a Component) gets these handlers for free.

## Platform Drivers

goview expects a minimal driver interface:

```go
type WebViewDriver interface {
    Eval(js string)
    BindString(name string, fn func(string))
    BindInt(name string, fn func(int))
    BindVoid(name string, fn func())
    SetHTML(html string)
    Run()
}
```

The `example/` directory includes ready-made adapters for:

- **go-webview** (`github.com/abemedia/go-webview`) — default on macOS/Windows
- **go-webui** (`github.com/webui-dev/go-webui/v2`) — default on Linux

You can write your own adapter for Wails, Lorca, or any other webview backend that exposes `Eval` and `Bind`.

## Rendering

goview has no opinion on how you produce HTML strings. Common approaches:

**1. `html/template` or `text/template`**

```go
var userTmpl = template.Must(template.New("users").Parse(`...`))

func renderUsers(users []User) string {
    var buf bytes.Buffer
    userTmpl.Execute(&buf, users)
    return buf.String()
}
```

**2. `templ`**

```go
import "github.com/a-h/templ"

func renderUsers(users []User) string {
    var buf bytes.Buffer
    templ.Component(usersList(users)).Render(context.Background(), &buf)
    return buf.String()
}
```

**3. Plain `fmt.Sprintf`**

```go
func renderCount(n int) string {
    return fmt.Sprintf("<span>%d remaining</span>", n)
}
```

Use whichever fits the complexity of the view.

## Testing

Because ViewModels are pure Go structs with no webview dependency, they are fully unit-testable.

```go
func TestCounter(t *testing.T) {
    vm := NewCounterVM()
    syncDispatch := func(f func()) { f() }
    vm.BaseVM = goview.NewBaseVMWithDispatch(syncDispatch)

    vm.Increment()
    if vm.Count.Get() != 1 {
        t.Fatalf("expected 1, got %d", vm.Count.Get())
    }
}
```

`NewBaseVMWithDispatch` lets you inject a synchronous queue so Observable updates happen immediately in tests.

## Project Layout

A typical goview project looks like this:

```
myapp/
├── main.go          # wiring: driver, VM, App, Run
├── viewmodel.go     # ViewModel struct + business methods
├── renderer.go      # html/template definitions + render funcs
├── repo.go          # domain repository (optional)
├── index.html       # static HTML scaffold
└── driver/
    ├── driver.go    # Driver interface
    ├── webview.go   # go-webview adapter (!linux)
    └── webui.go     # go-webui adapter (linux)
```

## License

MIT
