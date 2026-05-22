package goview

import "fmt"

// DOM provides direct Go→DOM operations for cases where re-rendering a full
// component is overkill — toggling visibility, adding CSS classes, etc.
type DOM struct {
	eval func(string)
}

// NewDOM creates a DOM helper backed by the given eval function.
func NewDOM(eval func(string)) *DOM {
	return &DOM{eval: eval}
}

func (d *DOM) Show(selector string) {
	d.eval(fmt.Sprintf(`__goview.show(%s)`, mustJSON(selector)))
}

func (d *DOM) Hide(selector string) {
	d.eval(fmt.Sprintf(`__goview.hide(%s)`, mustJSON(selector)))
}

func (d *DOM) AddClass(selector, class string) {
	d.eval(fmt.Sprintf(`__goview.addClass(%s,%s)`, mustJSON(selector), mustJSON(class)))
}

func (d *DOM) RemoveClass(selector, class string) {
	d.eval(fmt.Sprintf(`__goview.removeClass(%s,%s)`, mustJSON(selector), mustJSON(class)))
}

// Eval sends arbitrary JS to the webview. Use sparingly — prefer Observable
// and Component for state-driven updates.
func (d *DOM) Eval(js string) {
	d.eval(js)
}
