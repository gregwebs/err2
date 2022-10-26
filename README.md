# Go try

This Go package provides tools for compact and composeable error handling.
Instead of the traditional:

``` go
x, err := f()
if err != nil {
	return fmt.Sprintf("annotate %v", err)
}
```

You can write:

``` go
x, err := f()
try.Try(err, try.Fmt("annotate"))
```

The functions `Check` and `Try` are used for checking and handling errors.
For example, instead of

```go
b, err := ioutil.ReadAll(r)
if err != nil {
    return err
}
...
```

we can call

```go
b, err := ioutil.ReadAll(r)
try.Check(err)
```

These function require a deferred error handler at the top of the function.


## Error handling

Every function which uses try for error-checking should have at least one deferred
`try.Handle*` function. `try` propagates errors via a panic, and these functions recover the error. If this is ommitted, an error will panic up the stack until there is a recover.

```go
func do() (err error) {
	defer try.Handlew(&err, "calling do")
	...
}
```

There is also
* `Handlew`: annotate the error with a message and wrap it (like fmt.Errorf with %w)
* `Handlef`: annotate the error with a message and wrap it (like fmt.Errorf with %v)
* `Handle`: call a function with the error
* `HandleCleanup`: call a cleanup function

There are also helpers `Catch*`, and `ErrorFromRecovery` that are useful for catching errors and panics in functions that do not return errors. These are generally callbacks, goroutines, and main.


## Panic handling

The handler functions will also annotate panics and then rethrow them.
This behavior can be disabled by setting `AnnotatePanics = false`

## Fork

This is a fork of github.com/lainio/err2 with a different user facing API.
Internally the panic/recovery mechanism of propagating errors is the same.
Besides only returning errors and not passing through values, differences are:

* Tracing is handled differently.
* There is no automatic mechanism for printing panics, instead users should create
their own standard way of doing this. The original stack trace printing code is still available under the stackprint module.
* Errors themselves are wrapped up with a stack trace that can be recovered and printed with "%+v".
* Panics are annotated by the `Handle*` functions


## Trying it out

There are scripts under the ./codemod directory (upgrade.sh and downgrade.sh) to automatically translate code to use `try` or to downgrade it back to the original error handling.

## Background

The original `err2` implements similar error handling mechanism as drafted in the original
[check/handle
proposal](https://go.googlesource.com/proposal/+/master/design/go2draft-error-handling-overview.md).

`try` encourages the single use of a `defer` statement at the top of the function and then using the `try.Try` functions to explicitly declare the error handlers for a function, similar to this [new proposal](https://github.com/golang/go/issues/55026). 

The package accomplishes error handling internally by using `panic/recovery`, which is less than ideal.
However, it works out well because:

* benchmarks show that when there is no error the overhead is non-existant
* it helps with properly handling panics

In general code should not pass errors in performance sensitive paths. Normally if it does (for example APIs that use `EOF` as an error), the API is not well designed.

The mandatory use of the `defer` might prevent some code optimisations like function inlining.
If you have highly performance sensitive code it is best not to use this library, particularly in functions that are not benchmarked.

The following form introduces no overhead in all Go versions:

``` go
x, err := f()
try.Check(Err)
```

#### Settings for Automatic Stack Tracing and panic annotation

By default, Try and Check will wrap the error so that it has a stack trace
This can be disabled by setting the `AddStackTrace = false`


By default. `Handle*` will annotate panics as well.
This can be disabled by setting `AnnotatePanics = false`

## Structure

try has the following package structure:
- The top-level main package try can be imported as try which combines both the try/err3 and try/try packages
- The `try/err3` package includes declarative error handling functions.
- The `try/try` package offers error checking functions.
- The `stackprint` package contains the original code from `err2` to help print stack traces
- The `assert` package contains the original code from `err2` to help with assertions.


