# err3

The package provides tools for compact and composeable error handling.
Instead of the traditional:

``` go
x, err := f()
if err != nil {
	return err
}
```

You can write:

``` go
x := try.Check1(f())
```

## Fork

This is a fork of github.com/lainio/err2 with a different user facing API.
Internally the panic/recovery mechanism of propagating errors is the same.
However, there is no automatic mechanism for printing panics.


## Structure

err3 has the following package structure:
- The top-level main package can be imported as err3 which includes both the err3 and try packages
- The `err3/err3` package includes declarative error handling functions.
- The `err3/try` package offers error checking functions.


## Error handling

Every function which uses err3 for error-checking should have at least one
`err3.Handle*` function declared with `defer`. If this is ommitted, an error will panic up the stack until it finds such a function that will recover.

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

There are also helpers that are useful for catching errors and panics mostly in top-level functions: `Catch`, `CatchAll`, `CatchTrace`. Generally a program can use `CatchAll` at the top-level.


## Error checks

The `try` package provides convenient helpers to check the errors. Since the Go
1.18 we have been using generics to have fast and convenient error checking.

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
b := try.Try1(ioutil.ReadAll(r))
...
```

but not without an error handler (`Return`, `Annotate`, `Handle`) or it just
panics your app if you don't have a `recovery` call in the current call stack.
However, you can put your error handlers where ever you want in your call stack.
That can be handy in the internal packages and certain types of algorithms.

We think that panicking for the errors at the start of the development is far
better than not checking errors at all.


## Background

The original `err2` implements similar error handling mechanism as drafted in the original
[check/handle
proposal](https://go.googlesource.com/proposal/+/master/design/go2draft-error-handling-overview.md).

`err3` encourages the single use of a `defer` statement at the top of the function and then using the `try.TryX` functions to explicitly declare the error handlers for a function, similar to this [new proposal](https://github.com/golang/go/issues/55026). 

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

This form introduces minimal overhead (technically 6x slower, but just 1.7 nanoseconds slower):

``` go
_ = err3.Try1(f())
```

#### Automatic And Optimized Stack Tracing

TODO
