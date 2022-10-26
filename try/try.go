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

func Fmtw(format string, args ...interface{}) func(error) error {
	return func(err error) error {
		args = append(args, err)
		return fmt.Errorf(format+": %w", args...)
	}
}

func Fmt(format string, args ...interface{}) func(error) error {
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

// Check is a helper function to return error values without adding a large if statement.
// It replaces the following code:
//
//	err := f()
//	if err != nil {
//		return handler(err)
//	}
//
// With this code:
//
//	try.Check(f(), handler)
//
// Using a handler is optional. Most of the time you should use `try.Checkw` or `try.Checkf`.
//
// If the error value nil, it is a noop
// If the error value is non-nil, the handler functions will be applied to the error
// Then the non-nil error will be given to panic.
// You must use try.Handle at the top of your function to recover the error and return it instead of letting the panic continue to unwind
//
// By default, Check will wrap the error so that it has a stack trace
// This can be disabled by setting the var AddStackTrace = false
func Check(err error, handlers ...func(error) error) {
	if err == nil {
		return
	}
	for _, handler := range handlers {
		if handler == nil {
			continue
		}
		
		// This both handles the fact that we allow cleanup functions
		// that intentionally return nil,
		// and doesn't allow a handler to accidentally eliminate the error by returning nil
		if errHandled := handler(err); errHandled != nil {
			err = errHandled
		}
	}

	if AddStackTrace {
		err = errors.AddStack(err)
	}

	panic(err)
}

func Checkw(err error, format string, args ...interface{}) {
	Check(err, Fmtw(format, args...))
}

func Checkf(err error, format string, args ...interface{}) {
	Check(err, Fmt(format, args...))
}

func CheckCleanup(err error, cleanupHandler func()) {
	Check(err, Cleanup(cleanupHandler))
}
