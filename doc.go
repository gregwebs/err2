/*
Package err3 provides three main functionality:
 1. err3 package includes helper functions for error recovery and handling
 2. try package is for error checking and handling

The traditional error handling idiom in Go is roughly akin to

	if err != nil { return err }

The err3 package drives programmers to focus more on error handling rather than
checking errors. We think that checks should be so easy that we never forget
them. The CopyFile example shows how it works:

	// CopyFile copies source file to the given destination. If any error occurs it
	// returns error value describing the reason.
	func CopyFile(src, dst string) (err error) {
	     // Add first error handler just to annotate the error properly.
	     defer err3.Handlef(&err, "copy %s %s", src, dst)

	     // Try to open the file. If error occurs now, err will be annotated and
	     // returned properly thanks to above err3.Returnf.
	     r := try.Check1(os.Open(src))
	     defer r.Close()

	     // Try to create a file. If error occurs now, err will be annotated and
	     // returned properly.
	     w := try.Try1(os.Create(dst))(try.Cleanup(func() {
	     	os.Remove(dst)
	     })
	     defer w.Close()

	     // Try to copy the file. If error occurs now, all previous error handlers
	     // will be called in the reversed order. And final return error is
	     // properly annotated in all the cases.
	     _ = try.Check1(io.Copy(w, r))

	     // All OK, just return nil.
	     return nil
	}

# Error checks

The try package provides convenient helpers to check the errors. For example,
instead of

	b, err := ioutil.ReadAll(r)
	if err != nil {
	   return err
	}

we can write

	b := try.Check1(ioutil.ReadAll(r))

Note that try.ToX functions are as fast as if err != nil statements. Please see
the try package documentation for more information about the error checks.

# Stack Tracing

By default, TryX and CheckX will wrap the error so that it has a stack trace
This can be disabled by setting the `AddStackTrace = false`

# Error handling

The beginning of every function should contain an `err3.Handle*` to ensure that
errors are caught. Otherwise errors will escape the function as a panic and you
will be relying on calling functions to properly recover from panics.
*/
package try
