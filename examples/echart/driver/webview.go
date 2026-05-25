//go:build !linux

package driver

import (
	"github.com/abemedia/go-webview"
	_ "github.com/abemedia/go-webview/embedded"
)

type webviewDriver struct {
	w webview.WebView
}

func newDriver() Driver {
	w := webview.New(true)
	return &webviewDriver{w: w}
}

func (d *webviewDriver) Eval(js string)              { d.w.Eval(js) }
func (d *webviewDriver) Dispatch(f func())            { d.w.Dispatch(f) }
func (d *webviewDriver) SetHTML(html string)          { d.w.SetHtml(html) }
func (d *webviewDriver) Run()                         { d.w.Run() }
func (d *webviewDriver) Destroy()                     { d.w.Destroy() }
func (d *webviewDriver) SetTitle(title string)        { d.w.SetTitle(title) }
func (d *webviewDriver) SetSize(w, h int)             { d.w.SetSize(w, h, webview.HintNone) }
func (d *webviewDriver) BindString(name string, fn func(string)) { d.w.Bind(name, fn) }
func (d *webviewDriver) BindInt(name string, fn func(int))       { d.w.Bind(name, fn) }
func (d *webviewDriver) BindVoid(name string, fn func())        { d.w.Bind(name, fn) }
