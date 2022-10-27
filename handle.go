package try

import (
	"fmt"
	"runtime"

	"github.com/gregwebs/errors"
)

var AnnotatePanics bool = true

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
func Handle(err *error, handlerFn func(err error) error) {
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
	handleRecover(r, err, func(_ error) error {
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
	if AddStackTrace {
		handleRecover(r, err, func(err error) error {
			args = append(args, err)
			return errors.Errorf(prefix+": %v", args...)
		})
	} else {
		handleRecover(r, err, func(err error) error {
			args = append(args, err)
			return fmt.Errorf(prefix+": %v", args...)
		})
	}
}

// Handlew is for annotating an error.
// Must be used as a `defer`.
// It wraps the error with a message, similar to using "%w" in a format string
// This function will convert panics to errors
func Handlew(err *error, prefix string, args ...any) {
	// We need to call `recover` here because of how it works with defer.
	r := recover()
	if AddStackTrace {
		handleRecover(r, err, func(err error) error {
			return errors.Wrapf(err, prefix, args...)
		})
	} else {
		handleRecover(r, err, func(err error) error {
			args = append(args, err)
			return fmt.Errorf(prefix+": %w", args...)
		})
	}
}

// Annotate panics with information from Handle* functions.
// A dummy error will be created with the Panic as a string
type PanicAnnotated struct {
	Panic any
	// This error is the same as the panic
	// It allows functions that expect to annotate an error
	// to provide their annotation
	Err error
}

func (p PanicAnnotated) Error() string {
	// %+v should be available to get a stack,
	// but we shouldn't need it because this
	// should get thrown in a stack trace
	return fmt.Sprintf("%+v, %v", p.Panic, p.Err)
}

// This function will convert panics to errors
func handleRecover(r any, err *error, handlerFn func(err error) error) {
	// Call the handlerFn if possible if the recovery is not nil
	// If a non-runtime error, use the error and don't panic
	// Otherwise panic again.
	// Panic again with PanicAnnotated so that errors can be annotated
	var panicked *PanicAnnotated

	switch r := r.(type) {
	case PanicAnnotated:
		if !AnnotatePanics {
			// This case isn't possible unless this flag
			// is changed while the program is running
			panic(r)
		}

		panicked = &r
		*err = r.Err

	case runtime.Error:
		// A Go panic
		if !AnnotatePanics {
			panic(r)
		}

		// Rethrow the panic, but first allow it to be annotated by attaching an error
		if *err == nil {
			// Convert to an error that has the stack trace
			*err = errors.New(fmt.Sprintf("%+v", r))
		}
		panicked = &PanicAnnotated{
			Panic: r,
		}
	case error:
		// try.Check or try.Try threw an error
		// assert *err == nil
		*err = r

	case nil:
		// There is nothing to recover from
		// There are multiple Handle* functions.
		// One may have already dealt with the error
		// Still run this handler if err != nil

	default:
		// A Go panic
		if !AnnotatePanics {
			panic(r)
		}

		// Rethrow the panic, but first allow it to be annotated by attaching an error
		if *err == nil {
			// Convert to an error that has the stack trace
			*err = errors.New(fmt.Sprintf("%+v", r))
		}
		panicked = &PanicAnnotated{
			Panic: r,
		}
	}

	if handlerFn != nil && *err != nil {
		if newErr := handlerFn(*err); newErr != nil {
			*err = newErr
		}
	}

	if panicked != nil {
		if *err != nil {
			panicked.Err = *err
		}
		panic(*panicked)
	}
}

// CatchAll can be used in a function that does not return an error.
// Must be used with defer
// Converts a panic to an error
func CatchAll(handlerFn func(error)) {
	r := recover()
	var err error
	handleRecover(r, &err, func(err error) error {
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
