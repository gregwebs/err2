# err3

The package provides tools for compact and composeable error handling.
Instead of the traditional:

``` go
x, err := f()
if err != nil {
	return fmt.Sprintf("annotate %v", err)
}
```

You can write:

``` go
x := err3.Try1(f())(err3.Fmt("annotate"))
```

## Fork

This is a fork of github.com/lainio/err2 with a different user facing API.
Internally the panic/recovery mechanism of propagating errors is the same.

Tracing is handled differently.
There is no automatic mechanism for printing panics, instead users should create
their own standard way of doing this.
Errors themselves are wrapped up with a stack trace that can be recovered
and printed with "%+v".
The original stack trace printing code is still available under the stackprint module.


## Structure

err3 has the following package structure:
- The top-level main package err3 can be imported as err3 which combines both the err3/err3 and err3/try packages
- The `err3/err3` package includes declarative error handling functions.
- The `err3/try` package offers error checking functions.
- The `stackprint` package contains the original code from `err2` to help print stack traces
- The `assert` package contains the original code from `err2` to help with assertions.


## Error checks

The functions `CheckX` and `TryX` are used for checking and handling errors.
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
b := err3.Check1(ioutil.ReadAll(r))
...
```

But they do require a deferred error handler at the top of the function.


## Error handling

Every function which uses err3 for error-checking should have at least one
`err3.Handle*` function declared with `defer`. These functions recover the error. If this is ommitted, an error will panic up the stack until there is a recover.

This is the simplest form of `err3.Handle*`.

```go
func do() error {
	defer err3.Handlef(&err, "do")
	...
}
```

There is also
* `Handlew`: wrap the error with %w instead of %v
* `Handle`: call a function with the error
* `HandleCleanup`: call a cleanup function

There are also helpers `CatchError`, `CatchAll`, and `ErrorFromRecovery` that are useful for catching errors and panics in functions that do not return errors. These are generally callbacks, goroutines, and main.


## Background

The original `err2` implements similar error handling mechanism as drafted in the original
[check/handle
proposal](https://go.googlesource.com/proposal/+/master/design/go2draft-error-handling-overview.md).

`err3` encourages the single use of a `defer` statement at the top of the function and then using the `err3.TryX` functions to explicitly declare the error handlers for a function, similar to this [new proposal](https://github.com/golang/go/issues/55026). 

The package accomplishes error handling internally by using `panic/recovery`, which is less than ideal.
However, it works out well because:

* benchmarks show that when there is no error, overhead is minimal
* it helps with properly handling panics

In general code should not pass errors in performance sensitive paths. Normally if it does (for example APIs that use `EOF` as an error), the API is not well designed.

The mandatory use of the `defer` might prevent some code optimisations like function inlining.
If you have highly performance sensitive code it is best not to use this library, particularly in functions that are not benchmarked.

The following form introduces no overhead:

``` go
x, err := f()
err3.Check(Err)
```

This form introduces minimal overhead.
On My Mac M1 it shows as taking an additional 1.7 nanoseconds, which is 6x slower than the original.

``` go
_ = err3.Check1(f())
```

#### Automatic And Optimized Stack Tracing

By default, TryX and CheckX will wrap the error so that it has a stack trace
This can be disabled by setting the `AddStackTrace = false`
