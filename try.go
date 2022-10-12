/*
Package try is a package for reducing error handling verbosity

Instead of 'x, err := f(); if err != nil { return handler(err) }'
One writes: 'x, err := f(); try.Try(err, handler)'

If the error is not nil it is automatically thrown via panic.
It is then caught by 'Handle' functions, which are required at the top of every function.

	  import (
		"github.com/gregwebs/try"
	  )

	  func do() (err error) {
	    defer try.Handlew(&err, "do")

	    x, err := f()
	    try.Try(err, try.Fmtw("called f"))
	  }

Package try is a package for try.Try and try.Check functions that implement the error
checking. Additionally, there are helper functions for creating handlers: Fmtw, Fmt, and Cleanup
*/
package try

import (
	"fmt"

	"github.com/gregwebs/errors"
)

var AddStackTrace bool = true

func Fmtw(format string, args ...any) func(error) error {
	return func(err error) error {
		args = append(args, err)
		return fmt.Errorf(format+": %w", args...)
	}
}

func Fmt(format string, args ...any) func(error) error {
	return func(err error) error {
		args = append(args, err)
		return fmt.Errorf(format+": %v", args...)
	}
}

func Cleanup(handler func()) func(error) error {
	return func(err error) error {
		handler()
		return err
	}
}

// Try is a helper function to return error values without adding a large if statement.
// It replaces the following code:
//
//	err := f()
//	if err != nil {
//		return handler(err)
//	}
//
// With this code:
//
//	try.Try(f(), handler)
//
// If the error value nil, it is a noop
// If the error value is non-nil, the handler functions will be applied to the error
// Then the non-nil error will be given to panic.
// You must use err3.Handle... at the top of your function to recover the error and return it instead of letting the panic continue to unwind
//
// By default, Try will wrap the error so that it has a stack trace
// This can be disabled by setting the var AddStackTrace = false
func Try[E error](errE E, handler func(E) error, handlers ...func(error) error) {
	if error(errE) == nil {
		return
	}
	err := error(errE)
	handlers = append([](func(error) error){any(handler).(func(error) error)}, handlers...)
	for _, handler := range handlers {
		if handler == nil {
			continue
		}
		errHandled := handler(err)
		// This both handles the fact that we allow cleanup functions
		// that intentionally return nil,
		// and doesn't allow a handler to accidentally eliminate the error by returning nil
		if error(errHandled) != nil {
			err = errHandled
		}
	}

	if AddStackTrace {
		err = errors.AddStack(err)
	}

	panic(err)
}

// Check is a helper function to immediately return error values without adding an if statement with a return.
// If an error occurs, it panics the error.
// You must use err3.Handle... at the top of your function to catch the error and return it instead of continuing the panic.
// the Try... functions an be used instead of Check... to add an error handler
//
// By default, Check will wrap the error so that it has a stack trace
// This can be disabled by setting the var AddStackTrace = false
func Check(err error) {
	if err != nil {
		if AddStackTrace {
			err = errors.AddStack(err)
		}
		panic(err)
	}
}
