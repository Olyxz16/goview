package goview

// DispatchProvider is anything that can provide a serial dispatch queue.
// [BaseVM] implements this interface, so any ViewModel that embeds BaseVM
// can be passed directly to [Observe].
type DispatchProvider interface {
	Dispatch() func(func())
}

// Observe creates a dispatch-aware Observable with the given initial value.
// The Observable inherits the dispatch queue from the provider, so all
// listener callbacks are serialised.
//
//	vm.Count = goview.Observe(0, vm)
func Observe[T any](initial T, dp DispatchProvider) *Observable[T] {
	return NewObservableWithDispatch(initial, dp.Dispatch())
}

// BaseVM is an embeddable struct that owns the serial dispatch queue.
// Compose it into your ViewModel to eliminate boilerplate around queue
// creation, Bind* wrappers, and SafeEval.
//
//	type MyVM struct {
//	    goview.BaseVM
//	    Count *goview.Observable[int]
//	}
//
//	vm := &MyVM{BaseVM: goview.NewBaseVM()}
//	vm.Count = goview.Observe(0, vm)
type BaseVM struct {
	dispatch func(func())
}

// NewBaseVM creates a BaseVM with its own internal action queue.
// The user never needs to call [NewActionQueue] directly.
func NewBaseVM() BaseVM {
	return BaseVM{dispatch: NewActionQueue()}
}

// NewBaseVMWithDispatch creates a BaseVM with an explicit dispatch queue.
// Useful for testing: pass a synchronous dispatch (func(f func()) { f() })
// to make Observable mutations execute immediately.
func NewBaseVMWithDispatch(dispatch func(func())) BaseVM {
	return BaseVM{dispatch: dispatch}
}

// Dispatch returns the serial queue owned by this BaseVM.
func (b BaseVM) Dispatch() func(func()) {
	return b.dispatch
}

// Eval creates an eval function that sends JS to the webview via the
// serial dispatch queue. This is the equivalent of calling
// [SafeEval] with this BaseVM's queue.
func (b BaseVM) Eval(eval func(string)) func(string) {
	return SafeEval(b.dispatch, eval)
}

// BindString wraps a string callback so it always runs on the dispatch queue.
func (b BaseVM) BindString(fn func(string)) func(string) {
	return BindString(b.dispatch, fn)
}

// BindInt wraps an int callback so it always runs on the dispatch queue.
func (b BaseVM) BindInt(fn func(int)) func(int) {
	return BindInt(b.dispatch, fn)
}

// BindVoid wraps a void callback so it always runs on the dispatch queue.
func (b BaseVM) BindVoid(fn func()) func() {
	return BindVoid(b.dispatch, fn)
}
