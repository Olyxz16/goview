package goview

import (
	"fmt"
	"os"
	"strings"
)

// WebViewDriver is the minimal interface goview needs from a webview backend.
type WebViewDriver interface {
	Eval(js string)
	BindString(name string, fn func(string))
	BindInt(name string, fn func(int))
	BindVoid(name string, fn func())
	SetHTML(html string)
	Run()
}

// Mounter is anything that can be mounted into the DOM.
type Mounter interface {
	Mount()
}

// App orchestrates the goview lifecycle: component registration, binding
// registration, HTML injection, and automatic mounting when the DOM is ready.
//
// Use it to eliminate the manual mount shim and HTML injection boilerplate.
//
//	app := goview.NewApp(d, eval, vm.Dispatch())
//	app.Mount(taskList, counter, statusBar)
//	app.Bind("AddTask", vm.BindString(vm.AddTask))
//	app.Run("index.html")
type App struct {
	driver     WebViewDriver
	eval       func(string)
	dispatch   func(func())
	components []Mounter
}

// NewApp creates an App with the given driver, eval function, and dispatch queue.
//
//	eval should be a SafeEval-wrapped function (see [SafeEval]).
//	dispatch should be the serial action queue (see [NewActionQueue]).
func NewApp(d WebViewDriver, eval func(string), dispatch func(func())) *App {
	return &App{
		driver:   d,
		eval:     eval,
		dispatch: dispatch,
	}
}

// Mount registers components to be mounted automatically when the DOM is ready.
func (a *App) Mount(components ...Mounter) {
	a.components = append(a.components, components...)
}

// Bind registers a named Go function that can be called from JS via w-call.
// The fn argument must be one of: func(string), func(int), or func().
func (a *App) Bind(name string, fn interface{}) {
	switch f := fn.(type) {
	case func(string):
		a.driver.BindString(name, f)
	case func(int):
		a.driver.BindInt(name, f)
	case func():
		a.driver.BindVoid(name, f)
	default:
		panic(fmt.Sprintf(
			"goview.App.Bind: unsupported type for %s: %T (expected func(string), func(int), or func())",
			name, fn,
		))
	}
}

// Run reads the user's HTML scaffold, injects the goview runtime, registers the
// __goviewReady binding, and starts the webview.
func (a *App) Run(htmlPath string) {
	// Register the auto-mount callback. When goview.js finishes loading,
	// it calls __goviewReady on DOMContentLoaded.
	a.driver.BindVoid("__goviewReady", func() {
		a.dispatch(func() {
			for _, c := range a.components {
				c.Mount()
			}
		})
	})

	raw, err := os.ReadFile(htmlPath)
	if err != nil {
		panic(fmt.Sprintf("goview.App.Run: failed to read %s: %v", htmlPath, err))
	}
	html := string(raw)

	// Inject idiomorph and goview.js before </body>. If no </body>, append.
	runtime := fmt.Sprintf(
		"<script>%s</script>\n<script>%s</script>",
		IdiomorphJS,
		RuntimeJS,
	)

	if strings.Contains(html, "</body>") {
		html = strings.Replace(html, "</body>", runtime+"\n</body>", 1)
	} else {
		html = html + "\n" + runtime
	}

	a.driver.SetHTML(html)
	a.driver.Run()
}
