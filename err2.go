package err2

import (
	"errors"
	"fmt"
	"io"
	"runtime/debug"
)

// Try is deprecated. Use try.To functions from try package instead.
// Try is as similar as proposed Go2 Try macro, but it's a function and it
// returns slice of interfaces. It has quite big performance penalty when
// compared to Check function.
func Try(args ...any) []any {
	check(args)
	return args
}

// Check is deprecated. Use try.To function instead.
// Check performs error check for the given argument. If the err is nil, it does
// nothing. According the measurements, it's as fast as
//  if err != nil {
//      return err
//  }
// on happy path.
func Check(err error) {
	if err != nil {
		panic(err)
	}
}

// FilterTry is deprecated. Use try.Is function instead.
// FilterTry performs filtered error check for the given argument. It's same
// as Check but before throwing an error it checks if error matches the filter.
// The return value false tells that there are no errors and true that filter is
// matched.
func FilterTry(filter, err error) bool {
	if err != nil {
		if errors.Is(filter, err) {
			return true
		}
		panic(err)
	}
	return false
}

// TryEOF is deprecated. Use try.IsEOF function instead.
// TryEOF checks errors but filters io.EOF from the exception handling and
// returns boolean which tells if io.EOF is present. See more info from
// FilterCheck.
func TryEOF(err error) bool {
	return FilterTry(io.EOF, err)
}

// Checks the error status of the last argument. It panics with "wrong
// signature" if the last calling parameter is not error. In case of error it
// delivers it by panicking.
func check(args []any) {
	argCount := len(args)
	last := argCount - 1
	if args[last] != nil {
		err, ok := args[last].(error)
		if !ok {
			panic("wrong signature")
		}
		panic(err)
	}
}

// Handle is for adding an error handler to a function by defer. It's for
// functions returning errors them self. For those functions that doesn't
// return errors there is a Catch function. Note! The handler function f is
// called only when err != nil.
func Handle(err *error, f func()) {
	// This and Catch are similar but we need to call recover() here because
	// how it works with defer. We cannot refactor these to use same function.

	// We put real panic objects back and keep only those which are
	// carrying our errors. We must also call all of the handlers in defer
	// stack.
	switch r := recover(); r.(type) {
	case nil:
		// Defers are in the stack and the first from the stack gets the
		// opportunity to get panic object's error (below). We still must
		// call handler functions to the rest of the handlers if there is
		// an error.
		if *err != nil {
			f()
		}
	case error:
		// We or someone did transport this error thru panic.
		*err = r.(error)
		f()
	default:
		panic(r)
	}
}

// Catch is a convenient helper to those functions that doesn't return errors.
// Go's main function is a good example. Note! There can be only one deferred
// Catch function per non error returning function. See Handle for more
// information.
func Catch(f func(err error)) {
	// This and Handle are similar but we need to call recover here because how
	// it works with defer. We cannot refactor these 2 to use same function.

	if r := recover(); r != nil {
		e, ok := r.(error)
		if !ok {
			panic(r)
		}
		f(e)
	}
}

// CatchAll is a helper function to catch and write handlers for all errors and
// all panics thrown in the current go routine.
func CatchAll(errorHandler func(err error), panicHandler func(v any)) {
	// This and Handle are similar but we need to call recover here because how
	// it works with defer. We cannot refactor these 2 to use same function.

	if r := recover(); r != nil {
		e, ok := r.(error)
		if ok {
			errorHandler(e)
		} else {
			panicHandler(r)
		}
	}
}

// CatchTrace is a helper function to catch and handle all errors. It recovers a
// panic as well and prints its call stack. This is preferred helper for go
// workers on long running servers.
func CatchTrace(errorHandler func(err error)) {
	// This and Handle are similar but we need to call recover here because how
	// it works with defer. We cannot refactor these 2 to use same function.

	if r := recover(); r != nil {
		e, ok := r.(error)
		if ok {
			errorHandler(e)
		} else {
			println(r)
			debug.PrintStack()
		}
	}
}

// Return is same as Handle but it's for functions which don't wrap or annotate
// their errors. If you want to annotate errors see Annotate for more
// information.
func Return(err *error) {
	// This and Handle are similar but we need to call recover here because how
	// it works with defer. We cannot refactor these two to use same function.

	if r := recover(); r != nil {
		e, ok := r.(error)
		if !ok {
			panic(r) // Not ours, carry on panicking
		}
		*err = e
	}
}

// Returnw wraps an error. It's similar to fmt.Errorf, but it's called only if
// error != nil. Note! If you don't want to wrap the error use Returnf instead.
func Returnw(err *error, format string, args ...any) {
	// This and Handle are similar but we need to call recover here because how
	// it works with defer. We cannot refactor these two to use same function.

	if r := recover(); r != nil {
		e, ok := r.(error)
		if !ok {
			panic(r) // Not ours, carry on panicking
		}
		*err = fmt.Errorf(format+": %w", append(args, e)...)
	} else if *err != nil { // if other handlers call recovery() we still..
		*err = fmt.Errorf(format+": %w", append(args, *err)...)
	}
}

// Annotatew is for annotating an error. It's similar to Returnf but it takes only
// two arguments: a prefix string and a pointer to error. It adds ": " between
// the prefix and the error text automatically.
func Annotatew(prefix string, err *error) {
	// This and Handle are similar but we need to call recover here because how
	// it works with defer. We cannot refactor these two to use same function.

	if r := recover(); r != nil {
		e, ok := r.(error)
		if !ok {
			panic(r) // Not ours, carry on panicking
		}
		*err = e
		format := prefix + ": %w"
		*err = fmt.Errorf(format, e)
	} else if *err != nil { // if other handlers call recovery() we still..
		format := prefix + ": %w"
		*err = fmt.Errorf(format, (*err))
	}
}

// Returnf builds an error. It's similar to fmt.Errorf, but it's called only if
// error != nil. Note! It doesn't use %w to wrap the error. Use Returnw for
// that.
func Returnf(err *error, format string, args ...any) {
	// This and Handle are similar but we need to call recover here because how
	// it works with defer. We cannot refactor these two to use same function.

	if r := recover(); r != nil {
		e, ok := r.(error)
		if !ok {
			panic(r) // Not ours, carry on panicking
		}
		*err = fmt.Errorf(format+": %v", append(args, e)...)
	} else if *err != nil { // if other handlers call recovery() we still..
		*err = fmt.Errorf(format+": %v", append(args, *err)...)
	}
}

// Annotate is for annotating an error. It's similar to Returnf but it takes only
// two arguments: a prefix string and a pointer to error. It adds ": " between
// the prefix and the error text automatically.
func Annotate(prefix string, err *error) {
	// This and Handle are similar but we need to call recover here because how
	// it works with defer. We cannot refactor these two to use same function.

	if r := recover(); r != nil {
		e, ok := r.(error)
		if !ok {
			panic(r) // Not ours, carry on panicking
		}
		*err = e
		format := prefix + ": %v"
		*err = fmt.Errorf(format, e)
	} else if *err != nil { // if other handlers call recovery() we still..
		format := prefix + ": %v"
		*err = fmt.Errorf(format, (*err))
	}
}

type _empty struct{}

// Empty is deprecated. Use try.To functions instead.
// Empty is a helper variable to demonstrate how we could build 'type wrappers'
// to make Try function as fast as Check.
var Empty _empty

// Try is deprecated. Use try.To functions instead.
// Try is a helper method to call func() (string, error) functions with it and
// be as fast as Check(err).
func (s _empty) Try(_ any, err error) {
	Check(err)
}
