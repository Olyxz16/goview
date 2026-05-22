package goview

import (
	"encoding/json"
	"fmt"
)

// Renderer is a function that converts data into an HTML string.
// Use any templating approach: text/template, html/template, templ, fmt.Sprintf.
type Renderer[T any] func(T) string

// Component binds an Observable to a DOM element identified by a CSS selector.
// When the Observable changes, the Component re-renders and patches the DOM
// via Idiomorph, preserving input state and focus.
//
// Equivalent to a JavaFX control bound to a Property.
type Component[T any] struct {
	selector   string
	observable *Observable[T]
	render     Renderer[T]
	eval       func(string)
}

// NewComponent creates a Component. Call Mount() after the webview DOM is ready.
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
		selector:   selector,
		observable: obs,
		render:     render,
		eval:       eval,
	}
}

// Mount registers the component with its observable and triggers the initial render.
// Must be called once, after the webview has loaded the HTML scaffold.
func (c *Component[T]) Mount() {
	// initial render
	c.push(c.observable.Get())

	// re-render on every change
	c.observable.OnChange(func(data T) {
		c.push(data)
	})
}

// push renders data and sends the result to the DOM via __goview.morph.
func (c *Component[T]) push(data T) {
	html := c.render(data)
	js := fmt.Sprintf(`__goview.morph(%s,%s)`, mustJSON(c.selector), mustJSON(html))
	c.eval(js)
}

func mustJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
