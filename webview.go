package goview

import (
	"bytes"
	"text/template"
)

// SafeEval wraps a webview's Eval and Dispatch functions to ensure JS is always
// executed on the webview thread. Pass the result as the eval argument to
// NewComponent and NewDOM.
//
//	eval := goview.SafeEval(w.Dispatch, w.Eval)
func SafeEval(dispatch func(func()), eval func(string)) func(string) {
	return func(js string) {
		dispatch(func() {
			eval(js)
		})
	}
}

// RenderTemplate executes a text/template and returns the HTML string.
// Errors are rendered as visible HTML fragments so they surface during development
// rather than silently producing blank content.
func RenderTemplate(tmpl *template.Template, data any) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
