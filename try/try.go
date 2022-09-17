/*
Package try is a package for reducing error handling verbosity

Instead of 'x, err := f(); if err != nil { return handler(err) }'
One writes: 'x := Try1(f(), handler)

If the error is not nil it is automatically thrown via panic.
It is then caught by 'Handle'

	  import (
		"github.com/gregwebs/err3"
		_ "github.com/gregwebs/try"
	  )

	  func do() (err error) {
	    defer err3.Handlew(&err, "do")

	    x := Try1(f())(Formatw("called f"))
	  }

Package try is a package for try.TryX functions that implement the error
checking. try.TryX functions check 'if err != nil' and if it throws the err to the
error handlers, which are implemented by the err3 package.

All of the try package functions should be as fast as the simple 'if err != nil {'
statement, thanks to the compiler inlining and optimization.
Currently though there is an

Note that try.ToX function names end to a number (x) because:

	"No variadic type parameters. There is no support for variadic type parameters,
	which would permit writing a single generic function that takes different
	numbers of both type parameters and regular parameters." - Go Generics

The leading number at the end of the To2 tells that To2 takes two different
non-error arguments, and the third one must be an error value.

Currently only To, To1, To2, and To3 are implemented, but more could be added.
*/
package try

import (
	"fmt"
	"github.com/pingcap/errors"
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
// 	if err != nil {
//		return handler(err)
//	}
//
// With this code:
//
// 	try.Try(f(), handler)
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

// Try1 operates similar to 'Try'
// The 1 indicates that one non-error value will be passed through.
// Try takes handler functions directly as arguments
// Due to limitations of the Go language, Try1 cannot.
// Instead Try1 returns a function that handlers are applied to.
// It replaces the following code:
//
//	x, err := f()
// 	if err != nil {
//		return handler(err)
//	}
//
// With this code:
//
// 	x := try.Try1(f())(handler)
func Try1[T any, E error](v T, err E) func(func(E) error, ...func(error) error) T {
	return func(handler func(E) error, handlers ...func(error) error) T {
		if error(err) != nil {
			Try[E](err, handler, handlers...)
		}

		return v
	}
}

// Try2 is the same as Try1 but passes through 2 values
func Try2[T, U any, E error](v1 T, v2 U, err E) func(func(E) error, ...func(error) error) (T, U) {
	return func(handler func(E) error, handlers ...func(error) error) (T, U) {
		if error(err) != nil {
			Try[E](err, handler, handlers...)
		}
		return v1, v2
	}
}

// Try2 is the same as Try1 but passes through 3 values
func Try3[T, U, V any, E error](v1 T, v2 U, v3 V, err E) func(func(E) error, ...func(error) error) (T, U, V) {
	return func(handler func(E) error, handlers ...func(error) error) (T, U, V) {
		if error(err) != nil {
			Try[E](err, handler, handlers...)
		}
		return v1, v2, v3
	}
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

// Check1 is the same as Check but passes along one extra value
func Check1[T any](v T, err error) T {
	Check(err)
	return v
}

// Check2 is the same as Check but passes along two extra values
func Check2[T, U any](v1 T, v2 U, err error) (T, U) {
	Check(err)
	return v1, v2
}

// Check2 is the same as Check but passes along three extra values
func Check3[T, U, V any](v1 T, v2 U, v3 V, err error) (T, U, V) {
	Check(err)
	return v1, v2, v3
}
