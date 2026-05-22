package goview

import "sync"

// Observable is a generic reactive value. When Set is called, all registered
// listeners are notified.
//
// If created with NewObservableWithDispatch, listener callbacks are
// dispatched to the provided serial queue. Otherwise they run synchronously
// on the calling goroutine.
//
// Equivalent to JavaFX's Property<T>.
type Observable[T any] struct {
	mu        sync.RWMutex
	value     T
	listeners []func(T)
	dispatch  func(func()) // serial queue; nil means synchronous
}

// NewObservable creates an Observable with an initial value.
// Listener callbacks run synchronously on the goroutine that calls Set.
func NewObservable[T any](initial T) *Observable[T] {
	return &Observable[T]{value: initial}
}

// NewObservableWithDispatch creates an Observable whose listeners are
// automatically dispatched to the given serial queue when Set is called.
// This is the recommended constructor for UI-bound Observables.
//
//	q := goview.NewActionQueue()
//	obs := goview.NewObservableWithDispatch(0, q)
func NewObservableWithDispatch[T any](initial T, dispatch func(func())) *Observable[T] {
	return &Observable[T]{value: initial, dispatch: dispatch}
}

// Set updates the value and notifies all listeners.
// If a dispatch queue was provided, listener callbacks are sent there;
// otherwise they run synchronously on the calling goroutine.
func (o *Observable[T]) Set(v T) {
	o.mu.Lock()
	o.value = v
	ls := make([]func(T), len(o.listeners))
	copy(ls, o.listeners)
	o.mu.Unlock()

	if o.dispatch != nil {
		for _, l := range ls {
			listener := l // capture for closure
			o.dispatch(func() { listener(v) })
		}
	} else {
		for _, l := range ls {
			l(v)
		}
	}
}

// Get returns the current value.
func (o *Observable[T]) Get() T {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.value
}

// OnChange registers a listener that is called whenever the value changes.
// The listener receives the new value.
func (o *Observable[T]) OnChange(fn func(T)) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.listeners = append(o.listeners, fn)
}

// onChange is the untyped variant used internally by Computed and Join.
func (o *Observable[T]) onChange(fn func()) {
	o.OnChange(func(_ T) { fn() })
}

// Dispatch returns the serial queue associated with this Observable, or nil.
func (o *Observable[T]) Dispatch() func(func()) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.dispatch
}

// Computed creates a read-only Observable whose value is derived from a source
// Observable. It re-evaluates fn whenever the source changes.
//
//	filtered := goview.Computed(vm.Users, func(users []User) []User {
//	    return filterByName(users, vm.Filter.Get())
//	})
//
// If the source has a dispatch queue, Computed inherits it so its own
// listeners are also serialised.
func Computed[S, T any](source *Observable[S], fn func(S) T) *Observable[T] {
	dispatch := source.Dispatch()
	var out *Observable[T]
	if dispatch != nil {
		out = NewObservableWithDispatch(fn(source.Get()), dispatch)
	} else {
		out = NewObservable(fn(source.Get()))
	}
	source.OnChange(func(v S) {
		out.Set(fn(v))
	})
	return out
}

// Join creates a read-only Observable from two source Observables.
// It re-evaluates fn whenever either source changes.
//
//	visible := goview.Join(vm.Tasks, vm.Filter, func(tasks []Task, filter string) []Task {
//	    return applyFilter(tasks, filter)
//	})
//
// The returned Observable inherits the dispatch queue from src1 (falling back
// to src2 if src1 has none), so its listeners are serialised.
func Join[S1, S2, T any](src1 *Observable[S1], src2 *Observable[S2], fn func(S1, S2) T) *Observable[T] {
	dispatch := src1.Dispatch()
	if dispatch == nil {
		dispatch = src2.Dispatch()
	}
	recompute := func() T { return fn(src1.Get(), src2.Get()) }
	var out *Observable[T]
	if dispatch != nil {
		out = NewObservableWithDispatch(recompute(), dispatch)
	} else {
		out = NewObservable(recompute())
	}
	src1.onChange(func() { out.Set(recompute()) })
	src2.onChange(func() { out.Set(recompute()) })
	return out
}
