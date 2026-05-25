package goview

import (
	"encoding/json"
	"fmt"
	"strings"
)

func mustJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// Renderer is a function that converts data into an HTML string.
// Use any templating approach: text/template, html/template, templ, fmt.Sprintf.
type Renderer[T any] func(T) string

// Component binds an Observable to a DOM element identified by a CSS selector.
// When the Observable changes, the Component re-renders and patches the DOM,
// preserving input state and focus.
//
// Equivalent to a JavaFX control bound to a Property.
type Component[T any] struct {
	selector string
	obs      *Observable[T]
	render   func(T) Patch
	eval     func(string)
}

// NewComponent creates a Component. Call Mount() after the webview DOM is ready.
// The render function returns a plain string; the Component wraps it in an
// InnerPatch so the DOM element's innerHTML is replaced on every update.
//
//	comp := goview.NewComponent("#user-list", vm.Users, renderUserList, eval)
//	comp.Mount()
func NewComponent[T any](
	selector string,
	obs *Observable[T],
	render Renderer[T],
	eval func(string),
) *Component[T] {
	return &Component[T]{
		selector: selector,
		obs:      obs,
		render:   func(v T) Patch { return Inner(render(v)) },
		eval:     eval,
	}
}

// NewPatchComponent creates a Component whose render function returns an
// explicit Patch. Use this for imperative JS objects (charts, editors, maps)
// that observe a container attribute via MutationObserver.
//
//	comp := goview.NewPatchComponent("#chart", vm.ChartData, renderChart, eval)
//	comp.Mount()
func NewPatchComponent[T any](
	selector string,
	obs *Observable[T],
	render func(T) Patch,
	eval func(string),
) *Component[T] {
	return &Component[T]{
		selector: selector,
		obs:      obs,
		render:   render,
		eval:     eval,
	}
}

// Mount registers the component with its observable and triggers the initial render.
// Must be called once, after the webview has loaded the HTML scaffold.
func (c *Component[T]) Mount() {
	// initial render
	c.apply(c.obs.Get())

	// re-render on every change
	c.obs.OnChange(func(data T) {
		c.apply(data)
	})
}

// apply renders data and sends the patch to the DOM via __goviewPatch.
func (c *Component[T]) apply(data T) {
	patch := c.render(data)
	switch p := patch.(type) {
	case InnerPatch:
		escaped := strings.ReplaceAll(p.HTML, "`", "\\`")
		c.eval(fmt.Sprintf(
			`__goviewPatch(%q, %q, %s)`,
			c.selector, "inner",
			"`"+escaped+"`",
		))
	case AttrPatch:
		c.eval(fmt.Sprintf(
			`__goviewPatch(%s, %s, %s, %s)`,
			mustJSON(c.selector), mustJSON("attr"), mustJSON(p.Name), mustJSON(p.Value),
		))
	}
}
