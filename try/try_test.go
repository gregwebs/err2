package try_test

import (
	"fmt"
	"io"
	"os"

	"github.com/gregwebs/try/err3"
	"github.com/gregwebs/try/try"
)

func Example_copyFile() {
	copyFile := func(src, dst string) (err error) {
		defer err3.Handlef(&err, "copy %s %s", src, dst)

		// These try package helpers are as fast as Check() calls which is as
		// fast as `if err != nil {}`

		r, err := os.Open(src)
		try.Check(err)
		defer r.Close()

		w, err := os.Create(dst)
		try.Check(err)
		defer err3.HandleCleanup(&err, func() {
			os.Remove(dst)
		})
		defer w.Close()
		_, err = io.Copy(w, r)
		try.Check(err)
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

		r, err := os.Open(src)
		try.Check(err)
		defer r.Close()

		w, err := os.Create(dst)
		try.Try(err, func(err error) error {
			os.Remove(dst)
			return err
		})
		defer w.Close()
		_, err = io.Copy(w, r)
		try.Check(err)
		return nil
	}

	err := copyFile("/notfound/path/file.go", "/notfound/path/file.bak")
	if err != nil {
		fmt.Println(err)
	}
	// Output: copy /notfound/path/file.go /notfound/path/file.bak: open /notfound/path/file.go: no such file or directory
}
