package main

import (
	"github.com/Olyxz16/goview"
	"github.com/Olyxz16/goview-tasks/driver"
)

func main() {
	repo := NewTaskRepo()
	vm := NewTaskViewModel(repo)

	d := driver.New()
	defer d.Destroy()
	d.SetTitle("Tasks")
	d.SetSize(800, 600)

	eval := vm.Eval(d.Eval)
	app := goview.NewApp(d, eval, vm.Dispatch())

	app.Mount(
		goview.NewComponent("#task-list", vm.Visible, renderTaskList, eval),
		goview.NewComponent("#status-bar", vm.Status, renderStatus, eval),
		goview.NewComponent("#task-count", vm.Count, renderCount, eval),
		goview.NewComponent("#filter-bar", vm.Filter, renderFilters, eval),
	)

	app.Bind("AddTask", vm.BindString(vm.AddTask))
	app.Bind("ToggleTask", vm.BindInt(vm.ToggleTask))
	app.Bind("DeleteTask", vm.BindInt(vm.DeleteTask))
	app.Bind("SetFilter", vm.BindString(vm.SetFilter))
	app.Bind("ClearDone", vm.BindVoid(vm.ClearDone))

	app.Run("index.html")
}
