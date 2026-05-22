//go:build linux

package driver

import "github.com/webui-dev/go-webui/v2"

type webuiDriver struct {
	w webui.Window
}

func newDriver() Driver {
	return &webuiDriver{w: webui.NewWindow()}
}

func (d *webuiDriver) Eval(js string)       { d.w.Run(js) }
func (d *webuiDriver) Dispatch(f func())   { f() } // thread-safe, no-op wrapper
func (d *webuiDriver) SetHTML(html string)  { d.w.Show(html) }
func (d *webuiDriver) Run()                { webui.Wait() }
func (d *webuiDriver) Destroy()            {} // cleanup handled by webui.Wait
func (d *webuiDriver) SetTitle(title string) {} // webui window title set via HTML
func (d *webuiDriver) SetSize(w, h int)      {} // sizing configured in HTML/JS

func (d *webuiDriver) BindString(name string, fn func(string)) {
	d.w.Bind(name, func(e webui.Event) any {
		arg, _ := webui.GetArg[string](e)
		fn(arg)
		return nil
	})
}

func (d *webuiDriver) BindInt(name string, fn func(int)) {
	d.w.Bind(name, func(e webui.Event) any {
		arg, _ := webui.GetArg[int](e)
		fn(arg)
		return nil
	})
}

func (d *webuiDriver) BindVoid(name string, fn func()) {
	d.w.Bind(name, func(e webui.Event) any {
		fn()
		return nil
	})
}
