package try_test

import (
	"fmt"
	"io"
	"os"

	"github.com/gregwebs/err2/err3"
	"github.com/gregwebs/err2/try"
)

var (
	errForTesting = fmt.Errorf("error for %s", "testing")
)

func Example_copyFile() {
	copyFile := func(src, dst string) (err error) {
		defer err3.Handlef(&err, "copy %s %s", src, dst)

		// These try package helpers are as fast as Check() calls which is as
		// fast as `if err != nil {}`

		r := try.Check1(os.Open(src))
		defer r.Close()

		w := try.Check1(os.Create(dst))
		defer err3.HandleCleanup(&err, func() {
			os.Remove(dst)
		})
		defer w.Close()
		try.Check1(io.Copy(w, r))
		return nil
	}

	err := copyFile("/notfound/path/file.go", "/notfound/path/file.bak")
	if err != nil {
		fmt.Println(err)
	}
	// Output: copy /notfound/path/file.go /notfound/path/file.bak: open /notfound/path/file.go: no such file or directory
}

func Example_copyFile_try() {
	copyFile := func(src, dst string) (err error) {
		defer err3.Handlef(&err, "copy %s %s", src, dst)

		// These try package helpers are as fast as Check() calls which is as
		// fast as `if err != nil {}`

		r := try.Check1(os.Open(src))
		defer r.Close()

		w := try.Try1(os.Create(dst))(func(err error) error {
			os.Remove(dst)
			return err
		})
		defer w.Close()
		try.Check1(io.Copy(w, r))
		return nil
	}

	err := copyFile("/notfound/path/file.go", "/notfound/path/file.bak")
	if err != nil {
		fmt.Println(err)
	}
	// Output: copy /notfound/path/file.go /notfound/path/file.bak: open /notfound/path/file.go: no such file or directory
}
