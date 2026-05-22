// Package goview provides a Go-native MVVM binding layer for webview-based
// desktop applications.
//
// # Core primitives
//
//   - [Observable] — a generic reactive value. Notifies listeners on Set.
//   - [Component] — binds an Observable to a DOM element. Re-renders on change.
//   - [Computed]  — derives a read-only Observable from a source Observable.
//   - [Join]      — derives a read-only Observable from two source Observables.
//   - [Observe]   — creates a dispatch-aware Observable via a [DispatchProvider].
//   - [DOM]       — direct Go→DOM helpers (show, hide, addClass, etc.)
//   - [BaseVM]    — embeddable struct that owns the action queue and provides
//     Bind* helpers and an Eval bridge.
//
// # Architecture
//
// Go owns all application state via ViewModels (plain Go structs with Observable
// fields). HTML is a static scaffold of named containers. The JS runtime
// (goview.js, ~150 lines) handles DOM morphing and routes user interactions back
// to Go via webview.Bind.
//
// # Quick start
//
//	// 1. Create a ViewModel that embeds BaseVM.
//	vm := &AppVM{
//	    BaseVM: goview.NewBaseVM(),
//	}
//	vm.Count = goview.Observe(0, vm)
//
//	// 2. Bind components; re-renders are automatically serialised.
//	comp := goview.NewComponent("#counter", vm.Count, func(n int) string {
//	    return fmt.Sprintf(`<span>%d</span>`, n)
//	}, vm.Eval(d.Eval))
//	comp.Mount()
//
//	// 3. Wrap webview callbacks so they enter the queue before touching state.
//	d.Bind("Increment", vm.BindVoid(func() {
//	    vm.Count.Set(vm.Count.Get() + 1)
//	}))
//
// Because every state mutation and render runs on the same goroutine,
// your repositories and ViewModels need no mutexes. The ViewModel is
// fully unit-testable without a webview; use [NewBaseVMWithDispatch]
// with a synchronous dispatch for deterministic tests.
package goview
