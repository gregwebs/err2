package ifelse

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/gregwebs/try"
)

const idxTimeFmt = "2006-01-02 15:04:05.99999"

type addresses struct {
	ID        string
	CreatedAt string
	UpdatedAt string
}

func (i addresses) switchParse(b []byte) (_ interface{}, err error) {
	var d addresses
	err = json.Unmarshal(b, &d)
	if err != nil {
		return try.Zero[interface{}](), err
	}
	id, err := strconv.ParseInt(d.ID, 10, 64)
	if err != nil {
		return try.Zero[interface{}](), err
	}
	cr, err := time.Parse(idxTimeFmt, d.CreatedAt)
	if err != nil {
		return try.Zero[interface{}](), err
	}
	ud, err := time.Parse(idxTimeFmt, d.UpdatedAt)
	if err != nil {
		return try.Zero[interface{}](), err
	}
	s := struct {
		id int64
		cr time.Time
		ud time.Time
	}{
		id: id,
		cr: cr,
		ud: ud,
	}
	return &s, nil
}

func ifErr() (_ bool, err error) {
	err = errors.New("test if")
	if err != nil {
		return try.Zero[bool](), err
	}
	return true, nil
}
