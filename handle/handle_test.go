package handle_test

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/gregwebs/try/assert"
	"github.com/gregwebs/try/handle"
	"github.com/gregwebs/try/try"
)

type zeroStruct struct{}

func TestZero(t *testing.T) {
	assert.Equal(false, handle.Zero[bool]())
	assert.That(nil == handle.Zero[interface{}](), "zero interface{}")
	assert.Equal(zeroStruct{}, handle.Zero[zeroStruct](), "zero struct")
}

func throw() (string, error) {
	return "", fmt.Errorf("this is an ERROR")
}

func twoStrNoThrow() (string, string, error)        { return "test", "test", nil }
func intStrNoThrow() (int, string, error)           { return 1, "test", nil }
func boolIntStrNoThrow() (bool, int, string, error) { return true, 1, "test", nil }
func noThrow() (string, error)                      { return "test", nil }

func recursion(a int) (r int, err error) {
	defer handle.Do(&err, nil)

	if a == 0 {
		return 0, nil
	}
	s, err := noThrow()
	try.Check(err)
	_ = s
	r, err = recursion(a - 1)
	try.Check(err)
	r += a
	return r, nil
}

func cleanRecursion(a int) int {
	if a == 0 {
		return 0
	}
	s, err := noThrow()
	try.Check(err)
	_ = s
	return a + cleanRecursion(a-1)
}

func recursionWithErrorCheck(a int) (int, error) {
	if a == 0 {
		return 0, nil
	}
	s, err := noThrow()
	if err != nil {
		return 0, err
	}
	_ = s
	v, err := recursionWithErrorCheck(a - 1)
	if err != nil {
		return 0, err
	}
	return a + v, nil
}

func errHandlefOnly() (err error) {
	defer handle.Format(&err, "handle top")
	defer handle.Format(&err, "handle error")
	_, err = throw()
	try.Check(err)
	defer handle.Format(&err, "handle error")
	_, err = throw()
	try.Check(err)
	defer handle.Format(&err, "handle error")
	_, err = throw()
	try.Check(err)
	return err
}

func errTry1_Fmt() (err error) {
	defer handle.Format(&err, "handle top")
	// _ = try.Check1(throw())(func(err error) error { return fmt.Errorf("handle error: %v", err) })
	_, err = throw()
	try.Checkf(err, "handle error")
	_, err = throw()
	try.Checkf(err, "handle error")
	_, err = throw()
	try.Checkf(err, "handle error")
	return err
}

func errId(err error) error { return err }
func empty(_ error) error   { return nil }

func errTry1_id() (err error) {
	defer handle.Format(&err, "handle top")
	_, err = throw()
	try.Check(err, errId)
	_, err = throw()
	try.Check(err, errId)
	_, err = throw()
	try.Check(err, errId)
	return err
}

func errHandle_Only() (err error) {
	defer handle.Format(&err, "handle top")
	defer handle.Do(&err, empty)
	_, err = throw()
	try.Check(err)
	defer handle.Do(&err, empty)
	_, err = throw()
	try.Check(err)
	defer handle.Do(&err, empty)
	_, err = throw()
	try.Check(err)
	return err
}

func errTry1_inlineHandler() (err error) {
	defer handle.Format(&err, "handle top")
	_, err = throw()
	try.Check(err, func(err error) error { return fmt.Errorf("handle error: %v", err) })
	_, err = throw()
	try.Check(err, func(err error) error { return fmt.Errorf("handle error: %v", err) })
	_, err = throw()
	try.Check(err, func(err error) error { return fmt.Errorf("handle error: %v", err) })
	return err
}

func noErr() error {
	return nil
}

func TestTry_noError(t *testing.T) {
	_, err := noThrow()
	try.Check(err)
	_, _, err = twoStrNoThrow()
	try.Check(err)
	_, _, err = intStrNoThrow()
	try.Check(err)
	_, _, _, err = boolIntStrNoThrow()
	try.Check(err)
}

func TestDefault_Error(t *testing.T) {
	var err error
	defer handle.Do(&err, nil)

	_, err = throw()
	try.Check(err)

	t.Fail() // If everything works we are newer here
}

func TestTry_Error(t *testing.T) {
	var err error
	defer handle.Do(&err, nil)

	_, err = throw()
	try.Check(err)

	t.Fail() // If everything works we are newer here
}

func TestPanickingCatchHandlePanic(t *testing.T) {
	type args struct {
		f func()
	}
	tests := []struct {
		name  string
		args  args
		wants error
	}{
		{"general panic",
			args{
				func() {
					defer handle.CatchHandlePanic(func(err error) {}, func(v any) {})
					panic("panic")
				},
			},
			nil,
		},
		{"runtime.error panic",
			args{
				func() {
					defer handle.CatchHandlePanic(func(err error) {}, func(v any) {})
					var b []byte
					b[0] = 0
				},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() != nil {
					t.Error("panics should not fly thru")
				}
			}()
			tt.args.f()
		})
	}
}

func TestPanickingCatchTrace(t *testing.T) {
	noPanic := func(v any) {}
	noError := func(err error) {}

	type args struct {
		f func()
	}
	tests := []struct {
		name string
		args args
	}{
		{"general panic",
			args{
				func() {
					defer handle.CatchHandlePanic(noError, noPanic)
					panic("panic")
				},
			},
		},
		{"runtime.error panic",
			args{
				func() {
					defer handle.CatchHandlePanic(noError, noPanic)
					var b []byte
					b[0] = 0
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() != nil {
					t.Error("panics should NOT carry on when tracing")
				}
			}()
			tt.args.f()
		})
	}
}

func TestPanicking_Handle(t *testing.T) {
	annotate := func(err error) error {
		return fmt.Errorf("annotate %v", err)
	}
	type args struct {
		f func() error
	}
	tests := []struct {
		name string
		args args
	}{
		{"general panic",
			args{
				func() (err error) {
					defer func() {
						if err == nil {
							t.Errorf("err is nil")
						}
					}()
					defer handle.Do(&err, annotate)
					panic("general panic")
				},
			},
		},
		{"runtime.error panic",
			args{
				func() (err error) {
					defer handle.Do(&err, annotate)
					var b []byte
					b[0] = 0
					return nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				assertRecoveredHandle(t, r, "annotate")
			}()
			if err := tt.args.f(); err == nil {
				t.Error("panics should be caught and returned as errors")
			}
		})
	}
}

func TestPanicking_Handlef(t *testing.T) {
	type args struct {
		f func() error
	}
	tests := []struct {
		name string
		args args
	}{
		{"general panic",
			args{
				func() (err error) {
					defer func() {
						if err == nil {
							t.Errorf("err is nil")
						}
					}()
					defer handle.Format(&err, "handlef")
					handle.AnnotatePanics = true
					panic("general panic")
				},
			},
		},
		{"runtime.error panic",
			args{
				func() (err error) {
					defer handle.Format(&err, "handlef")
					var b []byte
					b[0] = 0
					return nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				assertRecoveredHandle(t, r, "handlef")
			}()
			if err := tt.args.f(); err == nil {
				t.Error("panics should be caught and returned as errors")
			}
		})
	}
}

func assertRecoveredHandle(t *testing.T, r any, panicMsg string) {
	t.Helper()
	if r == nil {
		t.Error("panics should be re-thrown")
	}
	err, ok := r.(handle.PanicAnnotated)
	if !ok {
		t.Errorf("expected panic to be re-thrown as PanicAnnotated")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, panicMsg) {
		t.Errorf("expected the handler message %s in the Error(), got %v", panicMsg, errMsg)
	}
}

func TestPanicking_Handlew(t *testing.T) {
	type args struct {
		f func() error
	}
	tests := []struct {
		name string
		args args
	}{
		{"general panic",
			args{
				func() (err error) {
					defer func() {
						if err == nil {
							t.Errorf("err is nil")
						}
					}()
					defer handle.Wrap(&err, "handlew")
					panic("panic")
				},
			},
		},
		{"runtime.error panic",
			args{
				func() (err error) {
					defer handle.Wrap(&err, "handlew")
					var b []byte
					b[0] = 0
					return nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				assertRecoveredHandle(t, r, "handlew")
			}()
			if err := tt.args.f(); err == nil {
				t.Error("panics should be caught and returned as errors")
			}
			tt.args.f()
		})
	}
}

func TestPanicking_Catch(t *testing.T) {
	type args struct {
		f func()
	}
	tests := []struct {
		name  string
		args  args
		wants error
	}{
		{"general panic",
			args{
				func() {
					defer handle.CatchError(func(err error) {})
					panic("panic")
				},
			},
			nil,
		},
		{"runtime.error panic",
			args{
				func() {
					defer handle.CatchError(func(err error) {})
					var b []byte
					b[0] = 0
				},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("panics should carry on")
				}
			}()
			tt.args.f()
		})
	}
}

func TestCatch_All(t *testing.T) {
	defer handle.CatchAll(func(err error) {
		//fmt.Printf("error and defer handling:%s\n", err)
	})

	_, err := throw()
	try.Check(err)

	t.Fail() // If everything works we are never here
}

func Example_copyFile() {
	copyFile := func(src, dst string) (err error) {
		defer handle.Format(&err, "copy %s %s", src, dst)

		// These try.To() checkers are as fast as `if err != nil {}`

		r, err := os.Open(src)
		try.Check(err)
		defer r.Close()

		w, err := os.Create(dst)
		defer handle.Cleanup(&err, func() {
			os.Remove(dst)
		})
		defer w.Close()
		_, err = io.Copy(w, r)
		try.Checkw(err, "copy failure")
		return nil
	}

	err := copyFile("/notfound/path/file.go", "/notfound/path/file.bak")
	if err != nil {
		fmt.Println(err)
	}
	// Output: copy /notfound/path/file.go /notfound/path/file.bak: open /notfound/path/file.go: no such file or directory
}

func ExampleDo() {
	var err error
	defer handle.Do(&err, nil)
	_, err = noThrow()
	try.Check(err)
	// Output:
}

func ExampleFormat() {
	annotated := func() (err error) {
		defer handle.Format(&err, "annotated")
		_, err = throw()
		try.Check(err)
		return err
	}
	err := annotated()
	fmt.Printf("%v", err)
	// Output: annotated: this is an ERROR
}

func ExampleFormat_format_args() {
	annotated := func() (err error) {
		defer handle.Format(&err, "annotated: %s", "handle")
		_, err = throw()
		try.Check(err)
		return err
	}
	err := annotated()
	fmt.Printf("%v", err)
	// Output: annotated: handle: this is an ERROR
}

func ExampleFormat_panic() {
	type fn func(v int) int
	var recursion fn
	const recursionLimit = 77 // 12+11+10+9+8+7+6+5+4+3+2+1 = 78

	recursion = func(i int) int {
		if i > recursionLimit { // simulated error case
			panic(fmt.Errorf("helper failed at: %d", i))
		} else if i == 0 {
			return 0 // recursion without error ends here
		}
		return i + recursion(i-1)
	}

	annotated := func() (err error) {
		defer handle.Format(&err, "annotated: %s", "handle")

		r := recursion(12) // call recursive algorithm successfully
		recursion(r)       // call recursive algorithm unsuccessfully
		return err
	}
	err := annotated()
	fmt.Printf("%v", err)
	// Output: annotated: handle: helper failed at: 78
}

func ExampleFormat_deferStack() {
	annotated := func() (err error) {
		defer handle.Format(&err, "3rd")
		defer handle.Format(&err, "2nd")
		_, err = throw()
		try.Checkf(err, "1st")
		return err
	}
	err := annotated()
	fmt.Printf("%v", err)
	// Output: 3rd: 2nd: 1st: this is an ERROR
}

func ExampleDo_with_handler() {
	doSomething := func(a, b int) (err error) {
		defer handle.Do(&err, func(err error) error {
			return fmt.Errorf("error with (%d, %d): %v", a, b, err)
		})
		_, err = throw()
		try.Check(err)
		return err
	}
	err := doSomething(1, 2)
	fmt.Printf("%v", err)
	// Output: error with (1, 2): this is an ERROR
}

func BenchmarkOldErrorCheckingWithIfClause(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, err := noThrow()
		if err != nil {
			return
		}
	}
}

func Benchmark_Err_HandleNil(b *testing.B) {
	for n := 0; n < b.N; n++ {
		errHandle_Only()
	}
}

func Benchmark_Err_Try1_id(b *testing.B) {
	for n := 0; n < b.N; n++ {
		errTry1_id()
	}
}

func Benchmark_Err_HandlersOnly(b *testing.B) {
	for n := 0; n < b.N; n++ {
		errHandlefOnly()
	}
}

func Benchmark_Err_Try1_Fmt(b *testing.B) {
	for n := 0; n < b.N; n++ {
		errTry1_Fmt()
	}
}

func Benchmark_NoErr_Check1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, err := noThrow() // a slight slow-dow
		try.Check(err)
	}
}

func Benchmark_Nohandle_Check_NilErr(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, err := noThrow()
		try.Check(err) // no slow-down
	}
}

func Benchmark_NoErr_Check2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _, err := twoStrNoThrow()
		try.Check(err)
	}
}

func Benchmark_NoErr_Check(b *testing.B) {
	for n := 0; n < b.N; n++ {
		try.Check(noErr())
	}
}

func Benchmark_NoErr_Check_NilErr(b *testing.B) {
	for n := 0; n < b.N; n++ {
		err := noErr()
		try.Check(err)
	}
}

func BenchmarkCleanRecursionWithTryCall(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = cleanRecursion(100)
	}
}

func BenchmarkRecursionWithCheckAndDefer(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = recursion(100)
	}
}

func BenchmarkRecursionWithOldErrorCheck(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, err := recursionWithErrorCheck(100)
		if err != nil {
			return
		}
	}
}
