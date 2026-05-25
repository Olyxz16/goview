// Package driver abstracts the underlying webview backend so that the
// application code in main.go stays platform-agnostic.
//
// Build tags select the concrete implementation:
//   - !linux  → go-webview (github.com/abemedia/go-webview)
//   - linux   → webui (github.com/webui-dev/go-webui/v2)
package driver

// Driver is the minimal surface area goview needs from a webview backend.
type Driver interface {
	Eval(js string)
	Dispatch(f func())
	BindString(name string, fn func(string))
	BindInt(name string, fn func(int))
	BindVoid(name string, fn func())
	SetHTML(html string)
	Run()
	Destroy()
	SetTitle(title string)
	SetSize(w, h int)
}

// New creates a Driver backed by the platform-specific webview implementation.
// The concrete type is chosen at compile time via build tags.
func New() Driver {
	return newDriver()
}
