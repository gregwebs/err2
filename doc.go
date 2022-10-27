/*
Package try provides two main functionality:
 1. handlers for error recovery and handling
 2. checks for returning non-nil errors

The traditional error handling idiom in Go is roughly akin to

	if err != nil { return err }

The try package drives programmers to focus more on error handling rather than
checking errors. We think that checks should be so easy that we never forget
them. The CopyFile example shows how it works:

	// CopyFile copies source file to the given destination. If any error occurs it
	// returns error value describing the reason.
	func CopyFile(src, dst string) (err error) {
	     // Add first error handler just to annotate the error properly.
	     defer try.Handlef(&err, "copy %s %s", src, dst)

	     // Try to open the file. If error occurs now, err will be annotated and
	     // returned properly thanks to above try.Check
	     r, err := os.Open(src)
	     try.Check(err)
	     defer r.Close()

	     // Try to create a file. If error occurs now, err will be annotated and
	     // returned properly.
	     w, err := os.Create(dst)
	     try.Try(err, try.Cleanup(func() {
	     	os.Remove(dst)
	     })
	     defer w.Close()

	     // Try to copy the file. If error occurs now, all previous error handlers
	     // will be called in the reversed order. And final return error is
	     // properly annotated in all the cases.
	     _, err = io.Copy(w, r)
	     try.Check(err)

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

	b, err := ioutil.ReadAll(r)
	try.Check(err)

# Stack Tracing

By default, try.Try and try.Check will wrap the error so that it has a stack trace
This can be disabled by setting the `AddStackTrace = false`

# Error handling

The beginning of every function should contain an `try.Handle*` to ensure that
errors are caught. Otherwise errors will escape the function as a panic and you
will be relying on calling functions to properly recover from panics.
*/
package try
