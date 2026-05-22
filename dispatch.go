package goview

// NewActionQueue creates a serial dispatch queue: a goroutine that executes
// submitted functions one at a time, in order. Pass the returned function as
// the dispatch argument to NewObservableWithDispatch, Component, and Bind*
// wrappers. This makes all UI state changes and renders happen on a single
// goroutine, so user code (ViewModels, repositories) needs no mutexes.
//
//	q := goview.NewActionQueue()
//	obs := goview.NewObservableWithDispatch(0, q)
//
// The queue is a buffered channel. If the buffer (default 256) fills up,
// submitters block until the queue goroutine catches up.
func NewActionQueue() func(func()) {
	return NewActionQueueSize(256)
}

// NewActionQueueSize is like NewActionQueue but lets you set the channel buffer.
func NewActionQueueSize(size int) func(func()) {
	ch := make(chan func(), size)
	go func() {
		for fn := range ch {
			fn()
		}
	}()
	return func(fn func()) {
		ch <- fn
	}
}

// BindString wraps a string callback so it always runs on the dispatch queue.
// Use it with webview.Bind to ensure Go-bound JS functions execute serially.
//
//	d.Bind("Greet", goview.BindString(q, vm.Greet))
func BindString(dispatch func(func()), fn func(string)) func(string) {
	return func(s string) {
		dispatch(func() { fn(s) })
	}
}

// BindInt wraps an int callback so it always runs on the dispatch queue.
func BindInt(dispatch func(func()), fn func(int)) func(int) {
	return func(n int) {
		dispatch(func() { fn(n) })
	}
}

// BindVoid wraps a void callback so it always runs on the dispatch queue.
func BindVoid(dispatch func(func()), fn func()) func() {
	return func() {
		dispatch(func() { fn() })
	}
}
