package err3

import (
	"fmt"
	"runtime"
)

// Handle handles any errors with a given handler function
//
// Every function using Try*/Check* must defer a Handle* function.
// Handle* functions must be used with `defer`
//
// If no additional error annotation is desired, 'nil' may be given as the handlerFn
// Handle is for adding an error handler to a function by deferring. It's for
// functions returning errors themself. For those functions that don't return
// errors, there are a CatchXxxx functions. The handler is called only when err
// != nil. There is no limit how many Handle functions can be added to defer
// stack. They all are called if an error has occurred and they are in deferred.
func Handle(err *error, handlerFn func() error) {
	// We need to call `recover` here because of how it works with defer.
	r := recover()
	handleRecover(r, err, handlerFn)
}

// HandleCleanup is a convenience function for using a handler that does not return an error.
// Must be used as a `defer`.
func HandleCleanup(err *error, handlerFn func()) {
	// We need to call `recover` here because of how it works with defer.
	r := recover()
	handleRecover(r, err, func() error {
		handlerFn()
		return nil
	})
}

// Handlef is for handling errors by annotating them with a format string.
// Must be used as a `defer`.
// It appends ": %v" to the format string
func Handlef(err *error, prefix string, args ...any) {
	// We need to call `recover` here because of how it works with defer.
	r := recover()
	formatHandler(r, err, prefix+": %v", args...)
}

// Handlew is for annotating an error.
// Must be used as a `defer`.
// It appends ": %w" to the format string
func Handlew(err *error, prefix string, args ...any) {
	// We need to call `recover` here because of how it works with defer.
	r := recover()
	formatHandler(r, err, prefix+": %w", args...)
}

func formatHandler(r any, err *error, format string, args ...any) {
	handleRecover(r, err, func() error {
		args = append(args, *err)
		return fmt.Errorf(format, args...)
	})
}

func handleRecover(r any, err *error, handlerFn func() error) {
	// Call the handlerFn if possible if the recovery is not nil
	// If a non-runtime error, use the error and don't panic
	// Otherwise panic again.
	shouldPanic := true
	switch r := r.(type) {
	case runtime.Error:
		// A normal Go panic
	case error:
		// A try.CheckX or try.TryX threw an error
		// assert *err == nil
		*err = r
		shouldPanic = false
	case nil:
		// There are multiple Handle* functions.
		// One already dealt with the error
		shouldPanic = false
	}
	if handlerFn != nil && *err != nil {
		if newErr := handlerFn(); newErr != nil {
			*err = newErr
		}
	}
	if shouldPanic {
		panic(r)
	}
}

// CatchError can be used in a function that does not return an error
// Must be used with defer
func CatchError(handlerFn func(error)) {
	r := recover()
	var err error
	handleRecover(r, &err, func() error {
		handlerFn(err)
		return nil
	})
}

// CatchAll can be used in a function that does not return an error
// Must be used with defer
//
// CatchAll stops panics and gives the panic to the panicHandler
// It uses ErrorFromRecover to extract errors and give them to the errorHandler
func CatchAll(errorHandler func(error), panicHandler func(v any)) {
	r := recover()
	if err := ErrorFromRecover(r); err != nil {
		errorHandler(err)
	} else {
		panicHandler(r)
	}
}

// ErrorFromRecover extracts a non-runtime error from the recovery object
// Otherwise it returns nil
func ErrorFromRecover(r any) error {
	switch r := r.(type) {
	case runtime.Error:
		return nil
	case error:
		return r
	default:
		return nil
	}
}
