package goview

// Patch represents a declarative DOM update. Concrete implementations are
// InnerPatch (replaces innerHTML) and AttrPatch (sets an element attribute).
type Patch interface {
	patch() // unexported, seals the interface
}

// InnerPatch replaces the element's innerHTML.
// This is the existing Component behavior, now made explicit.
type InnerPatch struct {
	HTML string
}

func (p InnerPatch) patch() {}

// AttrPatch sets a single attribute on the element.
// Used for imperative JS objects (charts, maps, editors) that observe their
// container via MutationObserver and manage their own lifecycle.
type AttrPatch struct {
	Name  string
	Value string
}

func (p AttrPatch) patch() {}

// Inner creates a Patch that replaces innerHTML.
func Inner(html string) Patch { return InnerPatch{HTML: html} }

// Attr creates a Patch that sets a single element attribute.
func Attr(name, value string) Patch { return AttrPatch{Name: name, Value: value} }
