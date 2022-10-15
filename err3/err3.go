package err3

import (
	"fmt"
	"runtime"

	"github.com/gregwebs/errors"
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
// This function will convert panics to errors
func Handle(err *error, handlerFn func() error) {
	// We need to call `recover` here because of how it works with defer.
	r := recover()
	handleRecover(r, err, handlerFn)
}

// HandleCleanup is a convenience function for using a handler that does not return an error.
// Must be used as a `defer`.
// This function will convert panics to errors
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
// This function will convert panics to errors
func Handlef(err *error, prefix string, args ...any) {
	// We need to call `recover` here because of how it works with defer.
	r := recover()
	formatHandler(r, err, prefix+": %v", args...)
}

// Handlew is for annotating an error.
// Must be used as a `defer`.
// It appends ": %w" to the format string
// This function will convert panics to errors
func Handlew(err *error, prefix string, args ...any) {
	// We need to call `recover` here because of how it works with defer.
	r := recover()
	formatHandler(r, err, prefix+": %w", args...)
}

// This function will convert panics to errors
func formatHandler(r any, err *error, format string, args ...any) {
	handleRecover(r, err, func() error {
		args = append(args, *err)
		return fmt.Errorf(format, args...)
	})
}

// This function will convert panics to errors
func handleRecover(r any, err *error, handlerFn func() error) {
	// Call the handlerFn if possible if the recovery is not nil
	// If a non-runtime error, use the error and don't panic
	// Otherwise panic again.
	switch r := r.(type) {
	case runtime.Error:
		// A Go panic
		// Convert to an error that has the stack trace
		// Overwrite err: it should be unset unless there was an error during error handling
		*err = errors.AddStack(errors.New(fmt.Sprintf("%+v", r)))
	case error:
		// try.Check or try.Try threw an error
		// assert *err == nil
		*err = r
	case nil:
		// There are multiple Handle* functions.
		// One already dealt with the error
	default:
		// A Go panic
		*err = errors.AddStack(errors.New(fmt.Sprintf("%v", r)))
	}
	if handlerFn != nil && *err != nil {
		if newErr := handlerFn(); newErr != nil {
			*err = newErr
		}
	}
}

// CatchAll can be used in a function that does not return an error.
// Must be used with defer
// Converts a panic to an error
func CatchAll(handlerFn func(error)) {
	r := recover()
	var err error
	handleRecover(r, &err, func() error {
		handlerFn(err)
		return nil
	})
}

// CatchHandlePanic can be used in a function that does not return an error
// Must be used with defer
//
// CatchHandlePanic stops panics and gives the panic to the panicHandler
// It uses ErrorFromRecover to extract errors and give them to the errorHandler
func CatchHandlePanic(errorHandler func(error), panicHandler func(v any)) {
	r := recover()
	if r == nil {
		return
	}
	if err := ErrorFromRecover(r); err != nil {
		errorHandler(err)
	} else {
		if panicHandler == nil {
			panic(r)
		} else {
			panicHandler(r)
		}
	}
}

// CatchError can be used in a function that does not return an error
// Must be used with defer
// Does not handle panics
func CatchError(errorHandler func(error)) {
	CatchHandlePanic(errorHandler, nil)
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

// Zero produces a zero value for a type
// This is made available to automate downgrading from this package
// When using the try.Check/Try functions, zero values are generated automatically by Golang
// When downgrading, you need to either manually construct zero values or you can automate with the help of this function
// For example, given the following code: function(err error) (bool, error) { try.Check(err), ... }
// The try.Check can be translated to if err != nil { return try.Zero(TYPE), err }
func Zero[Z any]() Z {
	var x Z
	return x
}
