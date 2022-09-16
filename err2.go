package err2

import (
	"fmt"
	"io"
	"os"

	"github.com/lainio/err2/internal/handler"
)

// StackTraceWriter allows to set automatic stack tracing.
//
//	err2.StackTraceWriter = os.Stderr // write stack trace to stderr
//	 or
//	err2.StackTraceWriter = log.Writer() // stack trace to std logger
var StackTraceWriter io.Writer

// Every function using err2/try must defer a Handle* function.
// If no additional error annotation is desired, 'nil' may be given as the handlerFn
// Handle is for adding an error handler to a function by deferring. It's for
// functions returning errors themself. For those functions that don't return
// errors, there are a CatchXxxx functions. The handler is called only when err
// != nil. There is no limit how many Handle functions can be added to defer
// stack. They all are called if an error has occurred and they are in deferred.
func Handle(err *error, handlerFn func()) {
	// We need to call `recover` here because of how it works with defer.
	r := recover()
	handleRecover(r, err, handlerFn)
}

// Handlef is for annotating an error.
// It's similar to Annotate but it appends ": %v" to the format string
func Handlef(err *error, prefix string, args ...any) {
	// We need to call `recover` here because of how it works with defer.
	r := recover()
	formatHandler(r, err, prefix+": %v", args...)
}

// Handlew is for annotating an error.
// It's similar to Annotate but it appends ": %w" to the format string
func Handlew(err *error, prefix string, args ...any) {
	// We need to call `recover` here because of how it works with defer.
	r := recover()
	formatHandler(r, err, prefix+": %w", args...)
}

func formatHandler(r any, err *error, format string, args ...any) {
	handleRecover(r, err, func() {
		args = append(args, *err)
		*err = fmt.Errorf(format, args...)
	})
}

func handleRecover(r any, err *error, handlerFn func()) {
	// We put real panic objects back and keep only those which are
	// carrying our errors. We must also call all of the handlers in defer
	// stack.
	if handlerFn == nil {
		handler.Process(handler.Info{
			Trace:        StackTraceWriter,
			Any:          r,
			ErrorHandler: func(e error) { *err = e },
		})
	} else {
		handler.Process(handler.Info{
			Trace: StackTraceWriter,
			Any:   r,
			NilHandler: func() {
				// Defers are in the stack and the first from the stack gets the
				// opportunity to get panic object's error (below). We still must
				// call handler functions to the rest of the handlers if there is
				// an error.
				if *err != nil {
					handlerFn()
				}
			},
			ErrorHandler: func(e error) {
				// We or someone did transport this error thru panic.
				*err = e
				handlerFn()
			},
		})
	}
}

// Catch is a convenient helper to those functions that doesn't return errors.
// There can be only one deferred Catch function per non error returning
// function like main(). It doesn't stop panics and runtime errors. If that's
// important use CatchAll or CatchTrace instead. See Handle for more
// information.
func Catch(f func(err error)) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	handler.Process(handler.Info{
		Trace:        StackTraceWriter,
		Any:          r,
		ErrorHandler: f,
	})
}

// CatchAll is a helper function to catch and write handlers for all errors and
// all panics thrown in the current go routine. It and CatchTrace are preferred
// helperr for go workers on long running servers, because they stop panics as
// well.
func CatchAll(errorHandler func(err error), panicHandler func(v any)) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	handler.Process(handler.Info{
		Trace:        StackTraceWriter,
		Any:          r,
		ErrorHandler: errorHandler,
		PanicHandler: panicHandler,
	})
}

// CatchTrace is a helper function to catch and handle all errors. It also
// recovers a panic and prints its call stack. It and CatchAll are preferred
// helpers for go-workers on long-running servers because they stop panics as
// well.
func CatchTrace(errorHandler func(err error)) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	handler.Process(handler.Info{
		Trace:        os.Stderr,
		Any:          r,
		ErrorHandler: errorHandler,
		PanicHandler: func(v any) {}, // suppress panicking
	})
}
